package crawler

import (
	"time"

	"github.com/benjaminestes/crawl/src/crawler/data"
)

type crawlfn func(*Crawler) crawlfn

func crawlStartQueue(c *Crawler) crawlfn {
	if len(c.queue) > 0 {
		return crawlStart
	}
	return nil
}

func crawlStart(c *Crawler) crawlfn {
	if time.Since(c.lastRequestTime) < c.wait {
		return crawlWait
	}
	return crawlCheckRobots
}

func crawlWait(c *Crawler) crawlfn {
	time.Sleep(c.wait - time.Since(c.lastRequestTime))
	return crawlStart
}

func crawlCheckRobots(c *Crawler) crawlfn {
	addr := c.queue[0]
	if _, ok := c.robots[addr.Host]; !ok {
		c.addRobots(addr.Full)
	}
	if c.robots[addr.Host] != nil && !c.robots[addr.Host].TestAgent(addr.RobotsPath(), c.Config.RobotsUserAgent) {
		result := data.MakeResult(addr, c.depth)
		result.Status = "Blocked by robots.txt"
		c.results <- result
		return crawlNext
	}
	return crawlDo
}

func crawlDo(c *Crawler) crawlfn {
	addr := c.queue[0]

	// This allows me to spawn no more than 20 fetches
	c.connections <- true
	c.resetWait()
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer func() { <-c.connections }()
		c.fetch(addr) // FIXME: implement actual queue?
	}()

	return crawlNext
}

func crawlNext(c *Crawler) crawlfn {
	c.queue[0] = nil
	c.queue = c.queue[1:]
	if len(c.queue) > 0 {
		return crawlStart
	}
	return crawlAwait
}

func crawlAwait(c *Crawler) crawlfn {
	c.wg.Wait()
	return crawlNextQueue
}

func crawlNextQueue(c *Crawler) crawlfn {
	c.queue = c.nextqueue
	c.nextqueue = nil
	c.depth++
	return crawlStartQueue
}
