package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/benjaminestes/crawl/src/crawler"
)

var config = &crawler.Config{
	RobotsUserAgent: "Crawler",
	WaitTime:        "100ms",
	MaxDepth:        -1,
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Expected command and config file path.")
		os.Exit(1)
	}

	configJSON, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(configJSON, config)
	if err != nil {
		fmt.Println("config error")
		os.Exit(0)
	}

	var c *crawler.Crawler

	switch os.Args[1] {
	case "site":
		c = crawlSite(config)
	case "list":
		c = crawlList(config)
	default:
		fmt.Fprintln(os.Stderr, "Invalid command.")
		os.Exit(1)
	}

	count := 0
	start := time.Now()
	for n := c.Next(); n != nil; n = c.Next() {
		j, _ := json.Marshal(n)
		fmt.Printf("%s\n", j)
		count++
		fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 65))
		fmt.Fprintf(
			os.Stderr,
			"\r%s : %d crawled",
			time.Since(start).Round(time.Second),
			count,
		)
	}

	fmt.Fprintf(os.Stderr, "\n")
}

func crawlSite(config *crawler.Config) *crawler.Crawler {
	return crawler.Crawl(config)
}

func crawlList(config *crawler.Config) *crawler.Crawler {
	var queue []*crawler.Node
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		n := crawler.MakeNode(0, scanner.Text())
		queue = append(queue, n)
	}
	config.MaxDepth = 0
	return crawler.CrawlList(config, queue)
}
