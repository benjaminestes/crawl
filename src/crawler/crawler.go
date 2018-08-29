package crawler

import (
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/temoto/robotstxt"
)

type Crawler struct {
	depth           int
	connections     chan bool
	queue           []*Address
	nextqueue       []*Address
	mu              sync.Mutex      // guards nextqueue
	wg              sync.WaitGroup  // watches for fetches
	seen            map[string]bool // Full text of address
	results         chan *Result
	robots          map[string]*robotstxt.RobotsData
	lastRequestTime time.Time
	wait            time.Duration
	include         []*regexp.Regexp
	exclude         []*regexp.Regexp
	client          *http.Client
	*Config
}

func Crawl(config *Config) *Crawler {
	// This should anticipate a failure condition
	first := MakeAddressFromString(config.Start)
	return CrawlList(config, []*Address{first})
}

func CrawlList(config *Config, q []*Address) *Crawler {
	// FIXME: Should handle error
	wait, _ := time.ParseDuration(config.WaitTime)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			MaxIdleConns:    config.Connections,
			IdleConnTimeout: 30 * time.Second,
		},
	}

	c := &Crawler{
		connections: make(chan bool, config.Connections),
		seen:        make(map[string]bool),
		results:     make(chan *Result, config.Connections),
		queue:       q,
		Config:      config,
		client:      client,
		wait:        wait,
		robots:      make(map[string]*robotstxt.RobotsData),
	}
	c.preparePatterns(config.Include, config.Exclude)

	for _, addr := range c.queue {
		c.seen[addr.Full] = true
	}

	go func() {
		for f := crawlStartQueue; f != nil; {
			f = f(c)
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
	if err != nil || resp.StatusCode != 200 {
		return
	}
	defer resp.Body.Close()

	robots, err := robotstxt.FromResponse(resp)
	if err != nil {
		return
	}
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
	c.lastRequestTime = time.Now()
}

func (c *Crawler) merge(links []*Link) {
	// This is how the crawler terminates — it will encounter an empty queue.
	if !(c.depth < c.MaxDepth) {
		return
	}
	for _, link := range links {
		if link.Address == nil || !c.WillCrawl(link.Address.Full) {
			continue
		}
		c.mu.Lock()
		if _, ok := c.seen[link.Address.Full]; !ok {
			if !(link.Nofollow && c.RespectNofollow) {
				c.seen[link.Address.Full] = true
				c.nextqueue = append(c.nextqueue, link.Address)
			}
		}
		c.mu.Unlock()
	}
}

func (c *Crawler) fetch(addr *Address) {
	result := MakeResult(addr, c.depth)

	resp, err := c.client.Get(addr.Full)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	result.Hydrate(resp)
	links := result.Links
	result.ResolvesTo = result.Address

	// If redirect, add target to list
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		result.ResolvesTo = MakeAddressFromRelative(addr, resp.Header.Get("Location"))
		links = []*Link{MakeLink(addr, resp.Header.Get("Location"), "", false)}
	}
	c.merge(links)
	c.results <- result
}
