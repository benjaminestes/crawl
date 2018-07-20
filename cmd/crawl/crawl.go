package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/benjaminestes/crawl/src/crawler"
)

var config = &crawler.Config{
	RobotsUserAgent: "Crawler",
	WaitTime:        "100ms",
}

func main() {
	configJSON, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(configJSON, config)
	if err != nil {
		fmt.Println("config error")
		os.Exit(0)
	}

	c := crawler.Crawl(config)

	count := 0
	start := time.Now()
	for n := c.Next(); n != nil; n = c.Next() {
		j, _ := json.Marshal(n)
		fmt.Printf("%s\n", j)
		count++
		fmt.Fprintf(
			os.Stderr,
			"\r%s : %d crawled : %d queued : %d seen",
			time.Since(start).Round(time.Second),
			count,
			len(c.Queue),
			len(c.Seen),
		)
	}

	fmt.Fprintf(os.Stderr, "\n")
}
