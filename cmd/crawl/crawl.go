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
}

func main() {
	configJSON, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) == 0 {
		fmt.Println("Please provide a base URL.")
		os.Exit(0)
	}

	url := os.Args[1]

	err = json.Unmarshal(configJSON, config)
	if err != nil {
		fmt.Println("config error")
		os.Exit(0)
	}
	c := crawler.Crawl(url, config)

	count := 0
	for n := c.Next(); n != nil; n = c.Next() {
		j, _ := json.Marshal(n)
		fmt.Printf("%s\n", j)
		count++
		fmt.Fprintf(os.Stderr, "\r%s %d/%d", time.Now().Format("2006/01/02 03:04:05"), count, len(c.Seen))
	}
}
