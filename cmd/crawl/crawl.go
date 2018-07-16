package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/benjaminestes/crawl/src/crawler"
)

func main() {
	if len(os.Args) == 0 {
		fmt.Println("Please provide a base URL.")
		os.Exit(0)
	}

	base, err := url.Parse(os.Args[1])
	if err != nil {
		fmt.Println("Broken initial URL.")
		os.Exit(0)
	}

	c := crawler.Crawl(base)

	for n := c.Next(); n != nil; n = c.Next() {
		j, _ := json.Marshal(n)
		fmt.Printf("%s\n", j)
	}
}
