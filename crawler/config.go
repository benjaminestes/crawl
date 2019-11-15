package crawler

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

var defaultCrawler = Crawler{
	Connections:     1,
	MaxDepth:        0,
	RobotsUserAgent: "Crawler",
	WaitTime:        "100ms",
	IdleConnTimeout: 30,
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
