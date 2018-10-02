// Copyright 2018 Benjamin Estes. All rights reserved.  Use of this
// source code is governed by an MIT-style license that can be found
// in the LICENSE file.

package crawler

import (
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/benjaminestes/crawl/src/crawler/data"
	"github.com/temoto/robotstxt"
)

type Crawler struct {
	depth   int
	queue   []*data.Address
	seen    map[string]bool // key = full text of URL
	results chan *data.Result

	// robots maintains a robots.txt matcher for every encountered
	// domain
	robots map[string]*robotstxt.RobotsData

	// mu guards nextqueue when multiple fetches may try to write
	// to it simultaneously
	nextqueue []*data.Address
	mu        sync.Mutex

	// wg waits for all spawned fetches to complete before
	// crawling the next level
	wg sync.WaitGroup

	// connections is a semaphore ensuring no more than
	// Config.Connections connections are active
	connections chan bool

	// wait is the parsed version of Config.WaitTime
	wait            time.Duration
	lastRequestTime time.Time

	// (in|ex)clude are the compiled versions of
	// Config.(In|Ex)clude, which are []string.
	include []*regexp.Regexp
	exclude []*regexp.Regexp

	client *http.Client
	*Config
}

// Crawl creates and starts a Crawler, and returns a pointer to it.
// The Crawler is a state machine running in its own
// goroutine. Therefore, calling this function may initiate many
// network requests, even before any results are requested from it.
func Crawl(config *Config) *Crawler {
	// This should anticipate a failure condition
	first := data.MakeAddressFromString(config.Start)
	return CrawlList(config, []*data.Address{first})
}

// initializeClient uses a config object to create an http.Client
// that conforms to the end-user's requirements.
func initializedClient(config *Config) *http.Client {
	return &http.Client{
		// Because we're checking the behavior of specific
		// URLs to understand whether they behave as expected,
		// we do not want to follow redirects.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			MaxIdleConns: config.Connections,
			// FIXME: make configurable
			IdleConnTimeout: 30 * time.Second,
		},
	}
}

// CrawlList starts and returns a *Crawler that is working from an
// input slice of Addresses rather than the start URL specified in
// config.
func CrawlList(config *Config, q []*data.Address) *Crawler {
	// FIXME: Should handle error
	wait, _ := time.ParseDuration(config.WaitTime)

	client := initializedClient(config)

	c := &Crawler{
		connections: make(chan bool, config.Connections),
		seen:        make(map[string]bool),
		results:     make(chan *data.Result, config.Connections),
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

func (c *Crawler) willCrawl(u string) bool {
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

// Returns the next result from the crawl. Results are guaranteed to come
// out in order ascending by depth. Within a "level" of depth, there is
// no guarantee as to which URLs will be crawled first.
//
// Result objects are suitable for Marshling into JSON format and conform
// to the schema exported by the crawler.Schema package.
func (c *Crawler) Next() *data.Result {
	node, ok := <-c.results
	if !ok {
		return nil
	}
	return node
}

func (c *Crawler) resetWait() {
	c.lastRequestTime = time.Now()
}

func (c *Crawler) merge(links []*data.Link) {
	// This is how the crawler terminates — it will encounter an empty queue.
	if !(c.depth < c.MaxDepth) {
		return
	}
	for _, link := range links {
		if link.Address == nil || !c.willCrawl(link.Address.Full) {
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

func (c *Crawler) fetch(addr *data.Address) {
	result := data.MakeResult(addr, c.depth)

	req, err := http.NewRequest("GET", addr.Full, nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", c.Config.UserAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	result.Hydrate(resp)
	links := result.Links
	result.ResolvesTo = result.Address

	// If redirect, add target to list
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		result.ResolvesTo = data.MakeAddressFromRelative(addr, resp.Header.Get("Location"))
		links = []*data.Link{data.MakeLink(addr, resp.Header.Get("Location"), "", false)}
	}
	c.merge(links)
	c.results <- result
}
