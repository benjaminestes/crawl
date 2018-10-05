// Copyright 2018 Benjamin Estes. All rights reserved.  Use of this
// source code is governed by an MIT-style license that can be found
// in the LICENSE file.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/benjaminestes/crawl/crawler"
	"github.com/benjaminestes/crawl/crawler/data"
	"github.com/benjaminestes/crawl/crawler/schema"
)

var config = &crawler.Config{
	RobotsUserAgent: "Crawler",
	WaitTime:        "100ms",
	MaxDepth:        -1,
}

func main() {
	schemaCommand := flag.NewFlagSet("schema", flag.ExitOnError)
	spiderCommand := flag.NewFlagSet("spider", flag.ExitOnError)
	listCommand := flag.NewFlagSet("list", flag.ExitOnError)
	listType := listCommand.String("format", "text", "format of input for list mode: {text|xml}")

	if len(os.Args) < 2 {
		log.Fatal(fmt.Errorf("expected command"))
		os.Exit(1)
	}

	var c *crawler.Crawler

	switch os.Args[1] {
	case "schema":
		schemaCommand.Parse(os.Args[2:])
		doSchema()
		return
	case "spider":
		spiderCommand.Parse(os.Args[2:])
		if spiderCommand.NArg() < 1 {
			log.Fatal(fmt.Errorf("expected location of config file"))
		}
		config := configFromFile(spiderCommand.Arg(0))
		c = crawlSpider(config)
	case "list":
		listCommand.Parse(os.Args[2:])
		if listCommand.NArg() < 1 {
			log.Fatal(fmt.Errorf("expected location of config file"))
		}
		config := configFromFile(listCommand.Arg(0))
		queue := listFromReader(os.Stdin)
		if *listType == "xml" {
			queue = listFromReader(os.Stdin)
		}
		c = crawlList(config, queue)
	default:
		flag.PrintDefaults()
		log.Fatal(fmt.Errorf("unknown command"))
	}

	doCrawl(c)
}

func configFromFile(name string) *crawler.Config {
	config := &crawler.Config{}

	configJSON, err := ioutil.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(configJSON, config)
	if err != nil {
		log.Fatal(err)
	}

	return config
}

func doCrawl(c *crawler.Crawler) {
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

func doSchema() {
	j, _ := json.MarshalIndent(schema.BQ, "", "\t")
	fmt.Printf("%s\n", j)
}

func listFromReader(in io.Reader) (queue []*data.Address) {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		n := data.MakeAddressFromString(scanner.Text())
		queue = append(queue, n)
	}
	return
}

func crawlSpider(config *crawler.Config) *crawler.Crawler {
	return crawler.Crawl(config)
}

func crawlList(config *crawler.Config, queue []*data.Address) *crawler.Crawler {
	config.MaxDepth = 0
	return crawler.CrawlList(config, queue)
}
