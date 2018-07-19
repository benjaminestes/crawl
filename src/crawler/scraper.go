package crawler

import (
	"net/url"

	"golang.org/x/net/html"
)

func scrape(n *Result, doc *html.Node) {
	n.Content.Title = firstTextChildOf(findNode("title", nil, doc))
	n.Content.H1 = firstTextChildOf(findNode("h1", nil, doc))
	n.Content.Description = attributeValue("content", (findNode("meta", map[string]string{
		"name": "description",
	}, doc)))
	n.Content.Robots = attributeValue("content", (findNode("meta", map[string]string{
		"name": "robots",
	}, doc)))
	n.Content.Canonical = attributeValue("href", (findNode("link", map[string]string{
		"rel": "canonical",
	}, doc)))
}

func attributeValue(k string, n *html.Node) string {
	if n == nil {
		return ""
	}
	for _, a := range n.Attr {
		if a.Key == k {
			return a.Val
		}
	}
	return ""
}

func matchAttributes(attrs map[string]string, n *html.Node) bool {
	for k, v := range attrs {
		seen := false
		for _, a := range n.Attr {
			if a.Key == k {
				seen = true
			}
			if a.Key == k && a.Val != v {
				return false
			}
		}
		if !seen {
			return false
		}
	}
	return true
}

func findNode(name string, attrs map[string]string, n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == name {
		// if attrs == nil {
		// 	return n
		//}
		if matchAttributes(attrs, n) {
			return n
		}
	}

	for m := n.FirstChild; m != nil; m = m.NextSibling {
		if o := findNode(name, attrs, m); o != nil {
			return o
		}
	}

	return nil
}

func getHreflang(base *url.URL, n *html.Node) (hreflang []*Hreflang) {
	nodes := findNodes("link", map[string]string{
		"rel": "alternate",
	}, n)

	for _, n := range nodes {
		lang := attributeValue("hreflang", n)
		href := attributeValue("href", n)
		if href != "" {
			hreflang = append(hreflang, MakeHreflang(href, lang))
		}
	}

	return
}

func findNodes(name string, attrs map[string]string, n *html.Node) (nodes []*html.Node) {
	if n.Type == html.ElementNode && n.Data == name && matchAttributes(attrs, n) {
		nodes = append(nodes, n)
	}

	for m := n.FirstChild; m != nil; m = m.NextSibling {
		nodes = append(nodes, findNodes(name, attrs, m)...)
	}

	return nodes
}

func firstTextChildOf(n *html.Node) string {
	switch {
	case n == nil:
		return ""
	case n.FirstChild == nil:
		return ""
	case n.FirstChild.Type == html.TextNode:
		return n.FirstChild.Data
	default:
		return ""
	}
}

func getLinks(base *url.URL, n *html.Node) (links []*Link) {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				url, err := url.Parse(attr.Val)
				if err != nil {
					break
				}

				newurl := base.ResolveReference(url)
				newurl.Fragment = "" // Ignore fragments
				link := MakeLink(
					newurl.String(),
					firstTextChildOf(n),
					attributeValue("rel") == "nofollow", // FIXME: Trim whitespace
				) // FIXME: get actual field values
				links = append(links, link)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		links = append(links, getLinks(base, c)...)
	}
	return
}
