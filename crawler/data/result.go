package data

import (
	"crypto/sha512"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/benjaminestes/crawl/scrape"
	"golang.org/x/net/html"
)

type Pair struct {
	K string
	V string
}

type Result struct {
	// Crawler state
	Address *Address `json:",omitempty"`
	Depth   int      `mode:"REQUIRED"`

	// Meta
	BodyTextHash string `json:",omitempty"`

	// Content
	Description string
	Title       string
	H1          string
	Robots      string
	Canonical   *Canonical  `json:",omitempty"`
	Links       []*Link     `json:",omitempty"`
	Hreflang    []*Hreflang `json:",omitempty"`

	// Response
	Status     string   `json:",omitempty"`
	StatusCode int      `json:",omitempty"`
	Proto      string   `json:",omitempty"`
	ProtoMajor int      `json:",omitempty"`
	ProtoMinor int      `json:",omitempty"`
	Header     []*Pair  `json:",omitempty"`
	ResolvesTo *Address `json:",omitempty"` // In case of redirect
}

func MakeResult(rawurl string, depth int, resp *http.Response) *Result {
	// FIXME: Should this contructor return an error?
	addr := MakeAddress(rawurl)
	result := &Result{
		Address: addr,
		Depth:   depth,
	}

	if resp != nil {
		result.hydrate(resp)
	}
	return result
}

func (r *Result) hydrate(resp *http.Response) {
	hydrateHeader(r, resp)

	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		doc, err := html.Parse(resp.Body)
		if err != nil {
			return
		}
		hydrateHTMLContent(r, doc)
	}

	// If the result doesn't redirect, we say it resolves to itself.
	r.ResolvesTo = r.Address
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		loc := resp.Header.Get("Location")
		r.ResolvesTo = MakeAddressResolved(r.Address, loc)
	}
}

func hydrateHeader(r *Result, resp *http.Response) {
	for k := range resp.Header {
		r.Header = append(r.Header, &Pair{k, resp.Header.Get(k)})
	}
	r.Status = resp.Status
	r.StatusCode = resp.StatusCode
	r.Proto = resp.Proto
	r.ProtoMajor = resp.ProtoMajor
	r.ProtoMinor = resp.ProtoMinor
}

func hydrateHTMLContent(r *Result, doc *html.Node) {
	r.Title = scrape.Text(scrape.Query("title", nil, doc))
	r.H1 = scrape.Text(scrape.Query("h1", nil, doc))
	r.Description = scrape.Attribute(
		"content",
		scrape.Query(
			"meta",
			map[string]string{
				"name": "description",
			},
			doc,
		))
	r.Robots = scrape.Attribute(
		"content",
		scrape.Query("meta",
			map[string]string{
				"name": "robots",
			},
			doc,
		))
	r.Canonical = getCanonical(r.Address, doc)
	r.Hreflang = getHreflang(r.Address, doc)
	r.Links = getLinks(r.Address, doc)

	sum := sha512.Sum512([]byte(scrape.Text(scrape.Query("body", nil, doc))))
	r.BodyTextHash = base64.StdEncoding.EncodeToString(sum[:])
}

func getCanonical(base *Address, n *html.Node) (c *Canonical) {
	href := scrape.Attribute("href", scrape.Query("link", map[string]string{
		"rel": "canonical",
	}, n))
	return MakeCanonical(base, href)
}

// FIXME: Should get the same URL resolving treatment as links
func getHreflang(base *Address, n *html.Node) (hreflang []*Hreflang) {
	nodes := scrape.QueryAll("link", map[string]string{
		"rel": "alternate",
	}, n)

	for _, n := range nodes {
		lang := scrape.Attribute("hreflang", n)
		href := scrape.Attribute("href", n)
		if href != "" {
			hreflang = append(hreflang, MakeHreflang(base, href, lang))
		}
	}

	return
}

func getLinks(base *Address, n *html.Node) (links []*Link) {
	els := scrape.NodesByTagName("a", n)
	for _, a := range els {
		href := scrape.Attribute("href", a)
		link := MakeLink(
			base,
			href,
			scrape.Text(a),
			scrape.Attribute("rel", a) == "nofollow", // FIXME: Trim whitespace?
		)
		links = append(links, link)
	}
	return links
}
