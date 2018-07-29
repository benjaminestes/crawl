package crawler

import (
	"fmt"
	"log"
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
	connections     chan bool
	newnodes        chan []*Node
	n               int
	Seen            map[string]bool // Full text of address
	results         chan *Result
	robots          map[string]*robotstxt.RobotsData
	LastRequestTime time.Time
	wait            time.Duration
	include         []*regexp.Regexp
	exclude         []*regexp.Regexp
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
		connections: make(chan bool, 20),
		Seen:        make(map[string]bool),
		results:     make(chan *Result, 20),
		newnodes:    make(chan []*Node),
		Config:      config,
		wait:        wait,
		robots:      make(map[string]*robotstxt.RobotsData),
	}
	c.preparePatterns(config.Include, config.Exclude)

	c.n++
	go func() { c.newnodes <- []*Node{first} }()
	go c.work()
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

	resp, err := http.Get(url.Scheme + "://" + url.Host + "/robots.txt")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	robots, _ := robotstxt.FromResponse(resp)
	c.robots[url.Host] = robots
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

func (c *Crawler) work() {
	for ; c.n > 0; c.n-- {
		nodes := <-c.newnodes
		for _, node := range nodes {
			switch {
			case node.Depth > c.MaxDepth && c.MaxDepth >= 0:
				continue
			case node.Address == nil:
				continue
			case !c.WillCrawl(node.Address.String()):
				continue
			case c.Seen[node.Address.String()]:
				continue
			case node.Nofollow && c.RespectNofollow:
				continue
			case c.robots[node.Address.Host] == nil:
				if _, ok := c.robots[node.Address.Host]; !ok {
					log.Printf("fetch robots %s\n", node.String())
					c.addRobots(node.Address.String())
				}
			case !c.robots[node.Address.Host].TestAgent(node.Address.RobotsPath(), c.Config.RobotsUserAgent):
				result := MakeResult(node.Address, node.Depth)
				result.Status = "Blocked by robots.txt"
				c.results <- result
				continue
			case time.Since(c.LastRequestTime) < c.wait:
				time.Sleep(c.wait - time.Since(c.LastRequestTime))
			}
			c.resetWait()
			log.Printf("released %s\n", node.String())
			c.Seen[node.Address.String()] = true
			c.n++
			go c.fetch(node)
		}
	}
	close(c.results)
}

func (c *Crawler) fetch(node *Node) {
	c.connections <- true
	defer func() { <-c.connections }()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	log.Printf("fetched %s\n", node.String())
	result := MakeResult(node.Address, node.Depth)

	resp, err := client.Get(node.Address.String())
	if err != nil {
		go func() { c.newnodes <- []*Node{} }()
		fmt.Fprintf(os.Stderr, "Couldn't fetch %s\n", node.Address)
		return
	}
	defer resp.Body.Close()

	tree, err := html.Parse(resp.Body)
	if err != nil {
		go func() { c.newnodes <- []*Node{} }()
		fmt.Fprintf(os.Stderr, "Couldn't parse %s\n", node.Address)
		return
	}

	result.Hydrate(resp, tree)
	links := result.Links

	// If redirect, add target to list
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		links = []*Link{
			MakeLink(
				node.Address,
				resp.Header.Get("Location"),
				"",
				false,
			),
		}
	}
	go func() {
		c.newnodes <- linksToNodes(node.Depth+1, links)
		log.Printf("added list len %d\n", len(links))
	}()
	log.Printf("sent result %s\n", node.String())
	c.results <- result
}

func linksToNodes(depth int, links []*Link) (nodes []*Node) {
	for _, link := range links {
		nodes = append(nodes, &Node{
			Depth: depth,
			Link:  link,
		})
	}
	return
}
