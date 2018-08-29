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
	Connections     int
	RobotsUserAgent string
	Include         []string
	Exclude         []string
	Start           string
	RespectNofollow bool
	MaxDepth        int
	WaitTime        string
}

type Crawler struct {
	depth           int
	connections     chan bool
	queue           []*Link
	nextqueue       []*Link
	mu              sync.Mutex      // guards nextqueue
	wg              sync.WaitGroup  // watches for merging new links
	Seen            map[string]bool // Full text of address
	results         chan *Result
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
	first := MakeAbsoluteLink(config.Start, "", false)
	return CrawlList(config, []*Link{first})
}

func CrawlList(config *Config, q []*Link) *Crawler {
	// FIXME: Should be configurable
	// also probably handle error
	wait, _ := time.ParseDuration(config.WaitTime)

	tr := &http.Transport{
		MaxIdleConns:    config.Connections,
		IdleConnTimeout: 30 * time.Second,
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: tr,
	}

	c := &Crawler{
		connections: make(chan bool, config.Connections),
		Seen:        make(map[string]bool),
		results:     make(chan *Result, config.Connections),
		queue:       q,
		Config:      config,
		client:      client,
		wait:        wait,
		robots:      make(map[string]*robotstxt.RobotsData),
	}
	c.preparePatterns(config.Include, config.Exclude)

	for _, v := range c.queue {
		c.Seen[v.Address.String()] = true
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
	c.LastRequestTime = time.Now()
}

func crawlStartQueue(c *Crawler) crawlfn {
	if len(c.queue) > 0 {
		return crawlStart
	}
	return nil
}

func crawlNextQueue(c *Crawler) crawlfn {
	c.queue = c.nextqueue
	c.nextqueue = nil
	c.depth++
	return crawlStartQueue
}

func crawlStart(c *Crawler) crawlfn {
	node := c.queue[0] // If this panics, there is a logic error.
	switch {
	case time.Since(c.LastRequestTime) < c.wait:
		return crawlWait
	case !c.WillCrawl(node.Address.String()):
		return crawlNext
	default:
		return crawlDo
	}
}

func crawlCheckRobots(c *Crawler) crawlfn {
	node := c.queue[0]
	if _, ok := c.robots[node.Address.Host]; !ok {
		c.addRobots(node.Address.String())
	}
	if !c.robots[node.Address.Host].TestAgent(node.Address.RobotsPath(), c.Config.RobotsUserAgent) {
		result := MakeResult(node.Address, c.depth)
		result.Status = "Blocked by robots.txt"
		c.results <- result
	}
	return crawlDo
}

func crawlWait(c *Crawler) crawlfn {
	time.Sleep(c.wait - time.Since(c.LastRequestTime))
	return crawlStart
}

func crawlNext(c *Crawler) crawlfn {
	c.queue = c.queue[1:]
	if len(c.queue) > 0 {
		return crawlStart
	}
	return crawlAwait
}

func crawlDo(c *Crawler) crawlfn {
	node := c.queue[0]

	// This allows me to spawn no more than 20 fetches
	c.connections <- true
	c.resetWait()
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer func() { <-c.connections }()
		c.fetch(node) // FIXME: implement actual queue?
	}()

	return crawlNext
}

func crawlAwait(c *Crawler) crawlfn {
	c.wg.Wait()
	return crawlNextQueue
}

func (c *Crawler) merge(links []*Link) {
	if !(c.depth < c.MaxDepth) {
		return
	}
	for _, link := range links {
		if link.Address == nil || !c.WillCrawl(link.Address.String()) {
			continue
		}
		c.mu.Lock()
		if _, ok := c.Seen[link.Address.String()]; !ok {
			if !(link.Nofollow && c.RespectNofollow) {
				c.Seen[link.Address.String()] = true
				c.nextqueue = append(c.nextqueue, link)
			}
		}
		c.mu.Unlock()
	}
}

func (c *Crawler) fetch(link *Link) {
	result := MakeResult(link.Address, c.depth)

	resp, err := c.client.Get(link.Address.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()

	result.Hydrate(resp)
	links := result.Links
	result.ResolvesTo = result.Address

	// If redirect, add target to list
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		result.ResolvesTo = MakeAddressFromRelative(link.Address, resp.Header.Get("Location"))
		links = []*Link{
			MakeLink(
				link.Address,
				resp.Header.Get("Location"),
				"",
				false,
			),
		}
	}
	c.merge(links)
	c.results <- result
}
