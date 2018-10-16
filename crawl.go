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
	"log"
	"os"
	"strings"
	"time"

	"github.com/benjaminestes/crawl/crawler"
	"github.com/benjaminestes/crawl/schema"
)

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
		fmt.Println(schema.BigQueryJSON())
		return
	case "spider":
		spiderCommand.Parse(os.Args[2:])
		if spiderCommand.NArg() < 1 {
			log.Fatal(fmt.Errorf("expected location of config file"))
		}
		config, err := os.Open(spiderCommand.Arg(0))
		if err != nil {
			log.Fatal(fmt.Errorf("%v", err))
		}
		c, err = crawler.FromJSON(config)
		if err != nil {
			log.Fatal(fmt.Errorf("%v", err))
		}
	case "list":
		listCommand.Parse(os.Args[2:])
		if listCommand.NArg() < 1 {
			log.Fatal(fmt.Errorf("expected location of config file"))
		}
		config, err := os.Open(listCommand.Arg(0))
		if err != nil {
			log.Fatal(fmt.Errorf("%v", err))
		}
		queue := listFromReader(os.Stdin)
		// FIXME: Here to justify listType existence.
		if *listType == "xml" {
			queue = listFromReader(os.Stdin)
		}
		c, err := crawler.FromJSON(config)
		if err != nil {
			log.Fatal(fmt.Errorf("%v", err))
		}
		c.From = queue
		c.MaxDepth = 0
	default:
		flag.PrintDefaults()
		log.Fatal(fmt.Errorf("unknown command"))
	}

	doCrawl(c)
}

func doCrawl(c *crawler.Crawler) {
	err := c.Start()
	if err != nil {
		// FIXME: need a way to signal error
		panic("couldn't start crawler")
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

func listFromReader(in io.Reader) []string {
	var queue []string
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		queue = append(queue, scanner.Text())
	}
	return queue
}
