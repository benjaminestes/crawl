package scrape

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func QueryAll(name string, attrs map[string]string, n *html.Node) []*html.Node {
	list := NodesByTagName(name, n)
	list = filterByAttributes(attrs, list)
	return list
}

// FIXME: inefficient
func Query(name string, attrs map[string]string, n *html.Node) *html.Node {
	list := QueryAll(name, attrs, n)
	if list == nil {
		return nil
	}
	return list[0]
}

func NodesByTagName(name string, node *html.Node) []*html.Node {
	a := atom.Lookup([]byte(name))
	var find func(*html.Node) []*html.Node
	find = func(node *html.Node) (list []*html.Node) {
		if node.DataAtom == a {
			list = append(list, node)
		}
		for next := node.FirstChild; next != nil; next = next.NextSibling {
			list = append(list, find(next)...)
		}
		return
	}
	return find(node)
}

func NodesByName(name string, node *html.Node) []*html.Node {
	var list []*html.Node
	if matchAttribute("name", name, node) {
		list = append(list, node)
	}
	for next := node.FirstChild; next != nil; next = next.NextSibling {
		list = append(list, NodesByName(name, next)...)
	}
	return list
}

func NodeByID(id string, node *html.Node) *html.Node {
	if matchAttribute("id", id, node) {
		return node
	}
	for next := node.FirstChild; next != nil; next = next.NextSibling {
		if el := NodeByID(id, next); el != nil {
			return el
		}
	}
	return nil
}

func NodesByClassName(name string, node *html.Node) []*html.Node {
	var list []*html.Node
	if matchClass(name, node) {
		list = append(list, node)
	}
	for next := node.FirstChild; next != nil; next = next.NextSibling {
		list = append(list, NodesByClassName(name, next)...)
	}
	return list
}

func matchAttribute(k, v string, n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	for _, a := range n.Attr {
		if a.Key == k && a.Val == v {
			return true
		}
	}
	return false
}

func matchAttributes(attrs map[string]string, n *html.Node) bool {
	for k, v := range attrs {
		if !matchAttribute(k, v, n) {
			return false
		}
	}
	return true
}

func filterByAttribute(k, v string, nodes []*html.Node) []*html.Node {
	var filtered []*html.Node
	for _, n := range nodes {
		if matchAttribute(k, v, n) {
			filtered = append(filtered, n)
		}
	}
	return filtered
}

func filterByAttributes(attrs map[string]string, nodes []*html.Node) []*html.Node {
	filtered := nodes
	for k, v := range attrs {
		filtered = filterByAttribute(k, v, filtered)
	}
	return filtered
}

func Attribute(k string, n *html.Node) string {
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

func matchClass(class string, n *html.Node) bool {
	for _, c := range Classes(n) {
		if c == class {
			return true
		}
	}
	return false
}

func Classes(node *html.Node) []string {
	return strings.Fields(Attribute("class", node))
}

func Text(n *html.Node) string {
	var b strings.Builder
	var getTextHelp func(node *html.Node)
	getTextHelp = func(node *html.Node) {
		switch {
		case node == nil:
			// Do nothing.
		case node.Type == html.TextNode:
			b.WriteString(node.Data)
		default:
			for next := node.FirstChild; next != nil; next = next.NextSibling {
				getTextHelp(next)
			}
		}
	}
	getTextHelp(n)
	return b.String()
}
