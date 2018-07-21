package crawler

import (
	"net/http"

	"golang.org/x/net/html"
)

type Pair struct {
	Key string
	Val string
}

type Result struct {
	Depth   int
	Content struct {
		Description string
		Title       string
		H1          string
		Robots      string
		Canonical   string
	}
	Response struct {
		Status        string
		StatusCode    int
		Proto         string
		ProtoMajor    int
		ProtoMinor    int
		ContentLength int64
		Header        []*Pair
	}
	Links    []*Link
	Hreflang []*Hreflang
	*Address
}

func MakeResult(addr *Address, depth int) *Result {
	return &Result{
		Address: addr,
		Depth:   depth,
	}
}

func (r *Result) Hydrate(resp *http.Response, doc *html.Node) {
	for k := range resp.Header {
		r.Response.Header = append(r.Response.Header, &Pair{k, resp.Header.Get(k)})
	}

	// Populate response fields
	r.Response.ContentLength = resp.ContentLength
	r.Response.Status = resp.Status
	r.Response.StatusCode = resp.StatusCode
	r.Response.Proto = resp.Proto
	r.Response.ProtoMajor = resp.ProtoMajor
	r.Response.ProtoMinor = resp.ProtoMinor

	// Populate Content fields
	scrape(r, doc)

	// Populate Hreflang fields
	r.Hreflang = getHreflang(r.URL, doc)

	// Populate and update links
	r.Links = getLinks(r.URL, doc)
}
