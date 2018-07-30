package crawler

import (
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/temoto/robotstxt"
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
	queue           []*Node
	nextqueue       []*Node
	n               int
	mu              sync.Mutex      // guards nextqueue
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
		queue:       []*Node{first},
		Config:      config,
		wait:        wait,
		robots:      make(map[string]*robotstxt.RobotsData),
	}
	c.preparePatterns(config.Include, config.Exclude)
	c.Seen[first.Address.String()] = true

	go func() {
		for len(c.queue) > 0 {
			c.work()
			c.queue = c.nextqueue
			c.nextqueue = nil
		}
		close(c.results)
	}()

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

func (c *Crawler) WillCrawl(u string) bool {
	for _, p := range c.exclude {
		if p.MatchString(u) {
			return false
		}
	}

	for _, p := range c.include {
		if p.MatchString(u) {
			return true
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

	// Now we've "seen" this host.
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
	var wg, wg2 sync.WaitGroup
	for _, node := range c.queue {
		switch {
		case node.Depth > c.MaxDepth && c.MaxDepth >= 0:
			continue
		case !c.WillCrawl(node.Address.String()):
			continue
		case c.robots[node.Address.Host] == nil:
			if _, ok := c.robots[node.Address.Host]; !ok {
				c.addRobots(node.Address.String())
			}
			if !c.robots[node.Address.Host].TestAgent(node.Address.RobotsPath(), c.Config.RobotsUserAgent) {
				result := MakeResult(node.Address, node.Depth)
				result.Status = "Blocked by robots.txt"
				c.results <- result
				continue
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
		n := node // Ensure binding doesn't get used by multiple gofuns
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.fetch(n)
		}()
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			c.merge()
		}()
	}
	wg.Wait()
	wg2.Wait()
}

func (c *Crawler) merge() {
	nodes := <-c.newnodes
	for _, node := range nodes {
		if node.Address == nil {
			continue
		}
		c.mu.Lock()
		if _, ok := c.Seen[node.Address.String()]; !ok {
			if !(node.Nofollow && c.RespectNofollow) {
				c.Seen[node.Address.String()] = true
				c.nextqueue = append(c.nextqueue, node)
			}
		}
		c.mu.Unlock()
	}
}

func (c *Crawler) fetch(node *Node) {
	c.connections <- true
	defer func() { <-c.connections }()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	result := MakeResult(node.Address, node.Depth)

	resp, err := client.Get(node.Address.String())
	if err != nil {
		go func() { c.newnodes <- []*Node{} }()
		// TODO: Couldn't fetch
		return
	}
	defer resp.Body.Close()

	result.Hydrate(resp)
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
	}()
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
