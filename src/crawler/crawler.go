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
	depth           int
	connections     chan bool
	queue           []*data.Address
	nextqueue       []*data.Address
	mu              sync.Mutex      // guards nextqueue
	wg              sync.WaitGroup  // watches for fetches
	seen            map[string]bool // Full text of address
	results         chan *data.Result
	robots          map[string]*robotstxt.RobotsData
	lastRequestTime time.Time
	wait            time.Duration
	include         []*regexp.Regexp
	exclude         []*regexp.Regexp
	client          *http.Client
	*Config
}

// Starts and returns a pointer to an active crawler. Crawling is done
// asynchronously, so many connections may be made just by instantiating
// a crawler with this function.
func Crawl(config *Config) *Crawler {
	// This should anticipate a failure condition
	first := data.MakeAddressFromString(config.Start)
	return CrawlList(config, []*data.Address{first})
}

// Behaves the same as Crawl(), but ignores the concept of depth. The
// returned crawl object fetches the URLs in the provided queue only.
func CrawlList(config *Config, q []*data.Address) *Crawler {
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
