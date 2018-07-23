package crawler

import (
	"net/http"
	"net/url"

	"github.com/benjaminestes/crawl/src/scrape"
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
	scrapeResult(r, doc)

	// Populate Hreflang fields
	r.Hreflang = getHreflang(r.URL, doc)

	// Populate and update links
	r.Links = getLinks(r.URL, doc)
}

func scrapeResult(n *Result, doc *html.Node) {
	n.Content.Title = scrape.GetText(scrape.QueryNode("title", nil, doc))
	n.Content.H1 = scrape.GetText(scrape.QueryNode("h1", nil, doc))
	n.Content.Description = scrape.GetAttribute("content", scrape.QueryNode("meta", map[string]string{
		"name": "description",
	}, doc))
	n.Content.Robots = scrape.GetAttribute("content", scrape.QueryNode("meta", map[string]string{
		"name": "robots",
	}, doc))
	n.Content.Canonical = scrape.GetAttribute("href", scrape.QueryNode("link", map[string]string{
		"rel": "canonical",
	}, doc))
}

// FIXME: Should get the same URL resolving treatment as links
func getHreflang(base *url.URL, n *html.Node) (hreflang []*Hreflang) {
	nodes := scrape.QueryNodes("link", map[string]string{
		"rel": "alternate",
	}, n)

	for _, n := range nodes {
		lang := scrape.GetAttribute("hreflang", n)
		href := scrape.GetAttribute("href", n)
		if href != "" {
			hreflang = append(hreflang, MakeHreflang(href, lang))
		}
	}

	return
}

func getLinks(base *url.URL, n *html.Node) (links []*Link) {
	els := scrape.GetNodesByTagName("a", n)
	for _, a := range els {
		h, _ := url.Parse(scrape.GetAttribute("href", a))
		t := base.ResolveReference(h)
		t.Fragment = ""
		link := MakeLink(
			t.String(),
			scrape.GetText(a),
			scrape.GetAttribute("rel", n) == "nofollow", // FIXME: Trim whitespace
		)
		links = append(links, link)
	}
	return links
}
