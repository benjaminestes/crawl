package crawler

import (
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/benjaminestes/crawl/version"
)

var defaultCrawler = Crawler{
	Connections:     1,
	MaxDepth:        0,
	MaxPages:		 0,
	UserAgent:       version.UserAgent(),
	RobotsUserAgent: "Crawler",

	// These fields must be set to avoid time parsing errors,
	// and to keep non-zero defaults colocated in this file.
	WaitTime:        "100ms",
	Timeout:         "30s",
}

func FromJSON(in io.Reader) (*Crawler, error) {
	config := defaultCrawler

	configJSON, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(configJSON, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
