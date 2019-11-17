// Copyright 2018 Benjamin Estes. All rights reserved.  Use of this
// source code is governed by an MIT-style license that can be found
// in the LICENSE file.

package crawler

import (
	"time"

	"github.com/benjaminestes/crawl/crawler/data"
	"github.com/benjaminestes/robots/v2"
)

// A crawlfn represents a state of the crawler state machine.  Its
// return value is the next state.
type crawlfn func(*Crawler) crawlfn

// crawlStartQueue is the initial state. If the current queue is
// empty, it returns nil. This is the ultimate termination condition.
func crawlStartQueue(c *Crawler) crawlfn {
	if len(c.queue) > 0 {
		return crawlStart
	}
	return nil
}

// crawlStart is the beginning of the process of crawling a single
// URL.
func crawlStart(c *Crawler) crawlfn {
	if time.Since(c.lastRequestTime) < c.wait {
		return crawlWait
	}
	return crawlCheckRobots
}

// crawlWait pauses if c.WaitTime has not elapsed since spawning the
// last request.
func crawlWait(c *Crawler) crawlfn {
	time.Sleep(c.wait - time.Since(c.lastRequestTime))
	return crawlStart
}

// crawlcheckrobots verifies that the domain being crawled allows the
// URL to be requested. If we get here, it means we've already decided
// the URL is in the scope of the crawl as defined by the end user.
func crawlCheckRobots(c *Crawler) crawlfn {
	addr := c.queue[0]
	rtxtURL, err := robots.Locate(addr.String())
	if err != nil {
		// Couldn't parse URL. Is this the desired behavior?
		return crawlNext
	}
	if _, ok := c.robots[rtxtURL]; !ok {
		c.addRobots(rtxtURL)
	}
	if !c.robots[rtxtURL](addr.String()) {
		// FIXME: Can this be some sort of "emit error" func?
		result := data.MakeResult(addr.String(), c.depth, nil)
		result.Status = "Blocked by robots.txt"
		c.results <- result
		return crawlNext
	}
	return crawlDo
}

// crawlDo begins the process of fetching a URL. If we're here, we're
// determined to try to crawl. The next step is to secure resources to
// actually crawl the URL, and initiate fetching.
func crawlDo(c *Crawler) crawlfn {
	addr := c.queue[0]
	// This blocks when there are = c.Connections fetches active.
	// Otherwise, it secures a token.
	c.connections <- true
	c.resetWait()
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer func() { <-c.connections }() // Release token
		// This fetch triggers the crawling of a URL and
		// ultimately the extraction of the links on the
		// crawled page. Merging of newly discovered URLs
		// happens as part of this call.
		c.fetch(addr)
	}()
	return crawlNext
}

// crawlNext tries to crawl the next URL in the queue. If there are no
// more URLs in the current queue, we wait for all currently active
// fetches to complete.
func crawlNext(c *Crawler) crawlfn {
	c.queue = c.queue[1:]
	if len(c.queue) > 0 {
		return crawlStart
	}
	return crawlAwait
}

// crawlAwait waits for all currently active fetches to finish.  This
// is done so that the crawler can proceed linearly through the crawl,
// level by level.
func crawlAwait(c *Crawler) crawlfn {
	c.wg.Wait()
	return crawlNextQueue
}

// crawlNextQueue replace the current queue with the next and starts
// the process again. This next queue represents the accumulated URLs
// in the next level of the crawl that we haven't yet seen.
func crawlNextQueue(c *Crawler) crawlfn {
	c.queue = c.nextqueue
	c.nextqueue = nil
	c.depth++
	return crawlStartQueue
}
