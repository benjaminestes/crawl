package crawler

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/temoto/robotstxt"
	"golang.org/x/net/html"
)

type Config struct {
	RobotsUserAgent string
	Include         []string
	Exclude         []string
	Start           string
	RespectNofollow bool
	WaitTime        string
}

type Node struct {
	Depth int
	*Link
}

type Crawler struct {
	Base            *url.URL
	Current         *Node
	Queue           []*Node
	Seen            map[string]bool // Full text of address
	results         chan *Result
	result          *Result
	newlist         []*Link
	robots          *robotstxt.RobotsData
	LastRequestTime time.Time
	wait            time.Duration
	include         []*regexp.Regexp
	exclude         []*regexp.Regexp
	client          *http.Client
	*Config
}

type crawlfn func(*Crawler) crawlfn

func Crawl(config *Config) *Crawler {
	// This should anticipate a failure condition
	first := &Node{
		Depth: 0,
		Link:  MakeLink(config.Start, "", false),
	}

	// FIXME: Should be configurable
	// also probably handle error
	wait, _ := time.ParseDuration(config.WaitTime)

	c := &Crawler{
		Base:    first.URL,
		Current: first, // ←
		Queue: []*Node{ // Only one of these should need to be set...
			first,
		},
		Seen:    make(map[string]bool),
		results: make(chan *Result),
		Config:  config,
		wait:    wait,
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
	c.preparePatterns(config.Include, config.Exclude)
	c.Seen[first.Address.FullAddress] = true
	c.fetchRobots()
	go c.run()
	return c
}

// Methods

func (c *Crawler) preparePatterns(include, exclude []string) {
	for _, s := range include {
		p := regexp.MustCompile(s)
		c.include = append(c.include, p)
	}
	for _, s := range exclude {
		p := regexp.MustCompile(s)
		c.exclude = append(c.exclude, p)
	}
}

func (c *Crawler) WillCrawl(u string) bool { // Should this test addresses?
	for _, p := range c.include {
		if p.MatchString(u) {
			return true
		}
	}

	for _, p := range c.exclude {
		if p.MatchString(u) {
			return false
		}
	}

	if len(c.include) > 0 {
		return false
	}
	return true
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

func (c *Crawler) emit() {
	c.results <- c.result
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

func (c *Crawler) resetWait() {
	c.LastRequestTime = time.Now()
}

// State machine functions

func crawlWait(c *Crawler) crawlfn {
	time.Sleep(10 * time.Millisecond)
	return crawlTryFetch
}

func crawlStart(c *Crawler) crawlfn {
	switch {
	case !c.WillCrawl(c.Current.Address.FullAddress) || c.Current.Nofollow:
		// If a URL does not match our include and exclude patterns,
		// or it was pointed to by a nofollow link, there will be
		// no result for it.
		return crawlSkip
	case time.Since(c.LastRequestTime) < c.wait:
		return crawlWait
	default:
		return crawlTryFetch
	}
}

func crawlSkip(c *Crawler) crawlfn {
	c.Queue = c.Queue[1:]
	if len(c.Queue) == 0 {
		return nil
	}
	c.Current = c.Queue[0] // Still not sure how much this helps
	return crawlStart
}

func crawlTryFetch(c *Crawler) crawlfn {
	if !c.robots.TestAgent(c.Current.URL.String(), c.Config.RobotsUserAgent) {
		c.result = MakeResult(c.Current.Address, c.Current.Depth)
		c.result.Response.Status = "Blocked by robots.txt"
		c.emit()
		return crawlSkip
	}
	return crawlDoFetch
}

func crawlDoFetch(c *Crawler) crawlfn {
	c.resetWait()

	c.result = MakeResult(c.Current.Address, c.Current.Depth)

	resp, err := c.client.Get(c.Current.URL.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't fetch %s\n", c.Current.Address)
		return crawlSkip
	}
	defer resp.Body.Close()

	tree, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't parse %s\n", c.Current.Address)
		return crawlSkip
	}

	c.result.Hydrate(resp, tree)
	c.newlist = c.result.Links

	// If redirect, add target to list
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		c.newlist = append(c.newlist, MakeLink(resp.Header.Get("Location"), "", false))
	}

	c.emit()
	return crawlMerge
}

func crawlMerge(c *Crawler) crawlfn {
	for _, link := range c.newlist {
		if c.Seen[link.Address.FullAddress] == false {
			if !link.Nofollow || !c.RespectNofollow {
				node := &Node{
					Depth: c.Current.Depth + 1,
					Link:  link,
				}
				c.Queue = append(c.Queue, node)
			}
			c.Seen[link.Address.FullAddress] = true
		}
	}
	return crawlSkip
}
