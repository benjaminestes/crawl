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
	MaxDepth        int
	WaitTime        string
}

type Node struct {
	Depth int
	*Link
}

type Crawler struct {
	Current         *Node
	Queue           []*Node
	Seen            map[string]bool // Full text of address
	results         chan *Result
	result          *Result
	newlist         []*Link
	robots          map[string]*robotstxt.RobotsData
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
		Link:  MakeAbsoluteLink(config.Start, "", false),
	}

	// FIXME: Should be configurable
	// also probably handle error
	wait, _ := time.ParseDuration(config.WaitTime)

	c := &Crawler{
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
		robots: make(map[string]*robotstxt.RobotsData),
	}
	c.preparePatterns(config.Include, config.Exclude)
	c.Seen[first.Address.String()] = true
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

func (c *Crawler) addRobots(u string) {
	url, err := url.Parse(u)
	if err != nil {
		return
	}

	// No matter what, make an entry for this host
	// That way we know we've check at least once

	c.robots[url.Host] = nil

	resp, err := c.client.Get(url.Scheme + "://" + url.Host + "/robots.txt")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	robots, _ := robotstxt.FromResponse(resp)
	c.robots[url.Host] = robots
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
	return crawlFetch
}

func crawlAddRobots(c *Crawler) crawlfn {
	// Crawler already has this state — does it need to be passed?
	c.addRobots(c.Current.Address.String())
	fmt.Fprintf(os.Stderr, "%s\n", "Check robots.txt for: "+c.Current.Address.Host)
	return crawlStart
}

func crawlStart(c *Crawler) crawlfn {
	switch {
	// FIXME: put "has max depth" into a method
	case c.Current.Depth > c.MaxDepth && c.MaxDepth >= 0:
		return crawlNext
	case !c.WillCrawl(c.Current.Address.String()) || c.Current.Nofollow:
		// If a URL does not match our include and exclude patterns,
		// or it was pointed to by a nofollow link, there will be
		// no result for it.
		return crawlNext
	case c.robots[c.Current.Address.Host] == nil:
		if _, ok := c.robots[c.Current.Address.Host]; !ok {
			// We haven't read robots.txt for the current domain!
			return crawlAddRobots
		}
		// We previously failed to find a robots file!
		return crawlFetch
		// FIXME: Test for robots.txt of domain of current URL
	case !c.robots[c.Current.Address.Host].TestAgent(c.Current.Address.Path+"?"+c.Current.Address.RawQuery, c.Config.RobotsUserAgent):
		return crawlRobotsBlocked
	case time.Since(c.LastRequestTime) < c.wait:
		return crawlWait
	default:
		return crawlFetch
	}
}

func crawlNext(c *Crawler) crawlfn {
	c.Queue = c.Queue[1:]
	if len(c.Queue) == 0 {
		return nil
	}
	c.Current = c.Queue[0] // Still not sure how much this helps
	return crawlStart
}

func crawlRobotsBlocked(c *Crawler) crawlfn {
	c.result = MakeResult(c.Current.Address, c.Current.Depth)
	c.result.Status = "Blocked by robots.txt"
	c.emit()
	return crawlNext
}

func crawlFetch(c *Crawler) crawlfn {
	c.resetWait()

	c.result = MakeResult(c.Current.Address, c.Current.Depth)

	resp, err := c.client.Get(c.Current.Address.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't fetch %s\n", c.Current.Address)
		return crawlNext
	}
	defer resp.Body.Close()

	tree, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't parse %s\n", c.Current.Address)
		return crawlNext
	}

	c.result.Hydrate(resp, tree)
	c.newlist = c.result.Links

	// If redirect, add target to list
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		c.newlist = []*Link{MakeLink(c.Current.Address, resp.Header.Get("Location"), "", false)}
	}

	c.emit()
	return crawlMerge
}

func crawlMerge(c *Crawler) crawlfn {
	for _, link := range c.newlist {
		if link.Address == nil {
			continue
		}
		if c.Seen[link.Address.String()] == false {
			if !(link.Nofollow && c.RespectNofollow) {
				node := &Node{
					Depth: c.Current.Depth + 1,
					Link:  link,
				}
				c.Queue = append(c.Queue, node)
			}
			c.Seen[link.Address.String()] = true
		}
	}
	return crawlNext
}
