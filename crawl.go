// Copyright 2018 Benjamin Estes. All rights reserved.  Use of this
// source code is governed by an MIT-style license that can be found
// in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/benjaminestes/crawl/crawler"
	"github.com/benjaminestes/crawl/schema"
	"github.com/benjaminestes/crawl/sitemap"
)

var (
	spiderCommand = flag.NewFlagSet("spider", flag.ExitOnError)
	listCommand   = flag.NewFlagSet("list", flag.ExitOnError)
	listType      = listCommand.String("format",
		"text", "format of input for list mode: {text|xml}")
	sitemapCommand = flag.NewFlagSet("sitemap", flag.ExitOnError)
)

func main() {
	if len(os.Args) < 2 {
		doHelp()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "help":
		doHelp()
		os.Exit(0)
	case "schema":
		doSchema()
	case "spider":
		doSpider()
	case "list":
		doList()
	case "sitemap":
		doSitemap()
	default:
		fmt.Fprintf(os.Stderr, "unexpected command: %s\n", os.Args[1])
		fmt.Fprintf(os.Stderr, `run "crawl help" for usage`+"\n")
		os.Exit(1)
	}
}

func doSchema() {
	os.Stdout.Write(schema.BigQueryJSON())
	fmt.Println()
}

func doSpider() {
	spiderCommand.Parse(os.Args[2:])
	if spiderCommand.NArg() < 1 {
		log.Fatal(fmt.Errorf("expected location of config file"))
	}
	config, err := os.Open(spiderCommand.Arg(0))
	if err != nil {
		log.Fatal(fmt.Errorf("%v", err))
	}
	c, err := crawler.FromJSON(config)
	if err != nil {
		log.Fatal(fmt.Errorf("%v", err))
	}
	doCrawl(c)
}

func doSitemap() {
	sitemapCommand.Parse(os.Args[2:])
	if sitemapCommand.NArg() < 1 {
		log.Fatal(fmt.Errorf("expected location of config file"))
	}
	config, err := os.Open(sitemapCommand.Arg(0))
	if err != nil {
		log.Fatalf("%v", err)
	}
	if sitemapCommand.NArg() < 2 {
		log.Fatal(fmt.Errorf("expected sitemap URL"))
	}
	var queue []string
	queue, err = fetchAll(sitemapCommand.Arg(1))
	if err != nil {
		log.Fatal(fmt.Errorf("error fetching sitemap"))
	}
	c, err := crawler.FromJSON(config)
	if err != nil {
		log.Fatalf("couldn't parse JSON config: %v", err)
	}
	c.From = queue
	c.MaxDepth = 0
	doCrawl(c)
}

func doList() {
	listCommand.Parse(os.Args[2:])
	if listCommand.NArg() < 1 {
		log.Fatal(fmt.Errorf("expected location of config file"))
	}
	config, err := os.Open(listCommand.Arg(0))
	if err != nil {
		log.Fatal(fmt.Errorf("%v", err))
	}
	var queue []string
	switch *listType {
	case "text":
		queue = listFromReader(os.Stdin)
		// FIXME: Here to justify listType existence.
	case "xml":
		queue, err = sitemap.Parse(os.Stdin)
		if err != nil {
			log.Fatalf("couldn't parse sitemap from stdin: %v", err)
		}
	}
	c, err := crawler.FromJSON(config)
	if err != nil {
		log.Fatal(fmt.Errorf("%v", err))
	}
	c.From = queue
	c.MaxDepth = 0
	doCrawl(c)
}

func doCrawl(c *crawler.Crawler) {
	count, lastCount := 0, 0
	lastUpdate := time.Now()
	err := c.Start()
	if err != nil {
		// FIXME: need a way to signal error
		panic("couldn't start crawler")
	}
	log.Printf("crawl started")
	for n := c.Next(); n != nil; n = c.Next() {
		j, _ := json.Marshal(n)
		fmt.Printf("%s\n", j)
		count++
		if time.Since(lastUpdate) > 5*time.Second {
			lastUpdate = time.Now()
			rate := (count - lastCount) / 5
			lastCount = count
			log.Printf("crawled %d (~%d/sec)", count, rate)
		}
	}

	log.Printf("crawl complete, %d URLs total", count)
}

func listFromReader(in io.Reader) []string {
	var queue []string
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		queue = append(queue, scanner.Text())
	}
	return queue
}

// fetchAll recursively produces a list of all URLs represented by the
// sitemap (index?) at url. If url points to a sitemap index, all of
// the sitemaps within that index will be recursively
// requested. Requests are not concurrent.
func fetchAll(url string) ([]string, error) {
	log.Printf("retrieving sitemap %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("error retrieving sitemap %s: %v", url, err)
	}
	defer resp.Body.Close()

	// It's possible we will need to try to parse the response
	// body twice, so read to []byte.
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading content of sitemap %s: %v", url, err)
	}

	var urls []string

	urls, err = sitemap.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	if len(urls) > 0 {
		return urls, nil
	}

	sitemaps, err := sitemap.ParseIndex(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	for _, s := range sitemaps {
		newurls, err := fetchAll(s)
		if err != nil {
			return nil, err
		}
		urls = append(urls, newurls...)
	}

	return urls, nil
}

func doHelp() {
	fmt.Println("USAGE: crawl <command> [-flags] [args]")
	fmt.Println()
	fmt.Println("The following commands are valid:")
	fmt.Println("\thelp, list, schema, sitemap, spider")
	fmt.Println()
	fmt.Println("help\t\tPrint this message.")
	fmt.Println()
	fmt.Println("list\t\tCrawl a list of URLs provided on stdin.")
	fmt.Println()
	fmt.Println("\t\tThe -format={text|xml} flag determines the expected type.")
	fmt.Println()
	fmt.Println("\t\tExample:")
	fmt.Println("\t\tcrawl list config.json <url_list.txt >out.txt")
	fmt.Println("\t\tcrawl list -format=xml config.json <sitemap.xml >out.txt")
	fmt.Println()
	fmt.Println("schema\t\tPrint a BigQuery-compatible JSON schema to stdout.")
	fmt.Println()
	fmt.Println("\t\tExample:")
	fmt.Println("\t\tcrawl schema >schema.json")
	fmt.Println()
	fmt.Println("sitemap\t\tRecursively requests a sitemap or sitemap index from")
	fmt.Println("\t\ta URL provided as argument.")
	fmt.Println()
	fmt.Println("\t\tExample:")
	fmt.Println("\t\tcrawl sitemap http://www.example.com/sitemap.xml >out.txt")
	fmt.Println()
	fmt.Println("spider\t\tCrawl from the URLs specific in the configuration file.")
	fmt.Println()
	fmt.Println("\t\tExample:")
	fmt.Println("\t\tcrawl spider config.json >out.txt")
}
