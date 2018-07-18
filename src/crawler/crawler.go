package crawler

import (
	"net/http"
	"net/url"

	"github.com/temoto/robotstxt"
	"golang.org/x/net/html"
)

type Config struct {
	RobotsUserAgent string
}

type Node struct {
	Depth int
	*Link
}

type Crawler struct {
	Base    *url.URL
	Current *Node
	Queue   []*Node
	Seen    map[string]bool // Full text of address
	results chan *Result    // FIXME: rename this, maybe?
	newlist []*Link
	robots  *robotstxt.RobotsData
	*Config
}

type crawlfn func(*Crawler) crawlfn

func Crawl(u string, config *Config) *Crawler {
	// This should anticipate a failure condition
	first := &Node{
		Depth: 0,
		Link:  MakeLink(u, "", true),
	}

	c := &Crawler{
		Base:    first.URL,
		Current: first, // ←
		Queue: []*Node{ // Only one of these should need to be set...
			first,
		},
		Seen:    make(map[string]bool),
		results: make(chan *Result),
		Config:  config,
	}
	c.Seen[first.Address.FullAddress] = true
	c.fetchRobots()
	go c.run()
	return c
}

func (c *Crawler) fetchRobots() {
	resp, err := http.Get(c.Base.Scheme + "://" + c.Base.Host + "/robots.txt")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	robots, err := robotstxt.FromResponse(resp)
	if err != nil {
		return
	}

	c.robots = robots
}

// Methods

func (c *Crawler) emit(n *Result) {
	c.results <- n
}

func (c *Crawler) run() {
	for state := crawlStart; state != nil; {
		state = state(c)
	}
	close(c.results)
}

func (c *Crawler) Next() *Result {
	node, ok := <-c.results
	if !ok {
		return nil
	}
	return node
}

// State machine functions

func crawlFetch(c *Crawler) crawlfn {
	if c.Current.URL.Host != c.Base.Host {
		return crawlSkip
	}

	// Ridiculous — split this out into functions
	r := new(Result)
	r.Address = c.Current.Address
	if c.robots.TestAgent(c.Current.URL.String(), c.Config.RobotsUserAgent) {
		resp, err := http.Get(c.Current.URL.String())
		if err != nil {
			return nil
		}
		defer resp.Body.Close()

		tree, err := html.Parse(resp.Body)
		if err != nil {
			return nil
		}

		// Process response and fill node

		// Process header fields
		for k := range resp.Header {
			r.Response.Header = append(r.Response.Header, &Pair{k, resp.Header.Get(k)})
		}

		// Populate response fields
		r.Response.ContentLength = resp.ContentLength
		r.Response.Status = resp.Status
		r.Response.StatusCode = resp.StatusCode
		r.Response.Proto = resp.Proto
		r.Response.ProtoMajor = resp.ProtoMajor
		r.Response.ProtoMinor = resp.ProtoMinor

		// Populate Content fields
		scrape(r, tree)

		// Populate Hreflang fields
		r.Hreflang = getHreflang(r.URL, tree)

		// Populate and update links
		c.newlist = getLinks(r.URL, tree)
		r.Links = c.newlist // Dangerous possibility of mutation?
	} else {
		r.Response.Status = "Blocked by robots.txt"
	}

	c.emit(r)
	return crawlMerge
}

func crawlStart(c *Crawler) crawlfn {
	if len(c.Queue) == 0 {
		return nil
	}
	return crawlFetch
}

func crawlSkip(c *Crawler) crawlfn {
	c.Queue = c.Queue[1:]
	c.Current = c.Queue[0]
	return crawlStart
}

func crawlMerge(c *Crawler) crawlfn {
	for _, link := range c.newlist {
		if c.Seen[link.Address.FullAddress] == false {
			node := &Node{
				Depth: c.Current.Depth + 1,
				Link:  link,
			}
			c.Queue = append(c.Queue, node)
			c.Seen[link.Address.FullAddress] = true
		}
	}
	return crawlSkip
}
