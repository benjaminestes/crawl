// Package sitemap is an internal package of the tool Crawl,
// responsible for parsing XML sitemaps and indexes.
package sitemap

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// These unexported types represent the necessary and sufficient data
// to crawl URLs discovered in sitemaps and indexes.
//
// Specification: https://www.sitemaps.org/protocol.html

// Sitemap

type urlset struct {
	URLs []string `xml:"url>loc"`
}

// Sitemap index

type index struct {
	Sitemaps []string `xml:"sitemap>loc"`
}

// Parse interprets in as a sitemap. It returns the URLs in that
// sitemap if successful.
func Parse(in io.Reader) ([]string, error) {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("Parse couldn't read sitemap data: %v", err)
	}

	res := &urlset{}

	err = xml.Unmarshal(data, res)
	if err != nil {
		return nil, fmt.Errorf("Parse failed to unmarshal sitemap data: %v", err)
	}

	return res.URLs, nil
}

// ParseIndex interprets in as a sitemap index. It returns the sitemap
// URLs in the index if successful.
func ParseIndex(in io.Reader) ([]string, error) {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}

	res := &index{}

	err = xml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}

	return res.Sitemaps, nil
}

// Fetch is like Parse, but it also retrieves its data from the given
// URL.
func Fetch(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	urls, err := Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return urls, nil
}

// FetchIndex is like ParseIndex, but it also retrieves its data from
// the given URL.
func FetchIndex(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	sitemaps, err := ParseIndex(resp.Body)
	if err != nil {
		return nil, err
	}
	return sitemaps, nil
}
