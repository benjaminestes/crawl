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

type Crawler struct {
	Base    *url.URL
	Current *Node
	Queue   []*Node
	Seen    map[string]bool
	nodes   chan *Node
	newlist []*Address
	robots  *robotstxt.RobotsData
	*Config
}

type crawlfn func(*Crawler) crawlfn

func Crawl(base *url.URL, config *Config) *Crawler {
	c := &Crawler{
		Base: base,
		Current: &Node{
			Address: &Address{
				Full: base.String(),
				URL:  base,
			},
		},
		Queue: []*Node{
			&Node{
				Address: &Address{
					Full: base.String(),
					URL:  base,
				},
			},
		},
		Seen:   make(map[string]bool),
		nodes:  make(chan *Node),
		Config: config,
	}
	c.Seen[base.String()] = true
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

func (c *Crawler) emit(n *Node) {
	c.nodes <- n
}

func (c *Crawler) run() {
	for state := crawlStart; state != nil; {
		state = state(c)
	}
	close(c.nodes)
}

func (c *Crawler) Next() *Node {
	node, ok := <-c.nodes
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

		for k := range resp.Header {
			c.Current.Response.Header = append(c.Current.Response.Header, &Pair{k, resp.Header.Get(k)})
		}

		c.Current.Response.ContentLength = resp.ContentLength
		c.Current.Response.Status = resp.Status
		c.Current.Response.StatusCode = resp.StatusCode
		c.Current.Response.Proto = resp.Proto
		c.Current.Response.ProtoMajor = resp.ProtoMajor
		c.Current.Response.ProtoMinor = resp.ProtoMinor
		scrape(c.Current, tree)

		c.newlist = getLinks(c.Current.URL, tree)
		c.Current.Links = c.newlist
	} else {
		c.Current.Response.Status = "Blocked by robots.txt"
	}

	// Emit before mutating c.Current
	c.emit(c.Current)
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
		if c.Seen[link.Full] == false {
			node := &Node{
				Address: link,
				Depth:   c.Current.Depth + 1,
			}
			c.Queue = append(c.Queue, node)
			c.Seen[link.Full] = true
		}
	}
	return crawlSkip
}
