package scrape

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func QueryNodes(name string, attrs map[string]string, n *html.Node) (list []*html.Node) {
	list = GetNodesByTagName(name, n)
	list = filterByAttributes(attrs, list)
	return
}

// FIXME: inefficient
func QueryNode(name string, attrs map[string]string, n *html.Node) *html.Node {
	list := QueryNodes(name, attrs, n)
	if list == nil {
		return nil
	}
	return list[0]
}

func GetNodesByTagName(name string, node *html.Node) []*html.Node {
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

func GetNodesByName(name string, node *html.Node) (list []*html.Node) {
	if matchAttribute("name", name, node) {
		list = append(list, node)
	}
	for next := node.FirstChild; next != nil; next = next.NextSibling {
		list = append(list, GetNodesByName(name, next)...)
	}
	return list
}

func GetNodeByID(id string, node *html.Node) *html.Node {
	if matchAttribute("id", id, node) {
		return node
	}
	for next := node.FirstChild; next != nil; next = next.NextSibling {
		if el := GetNodeByID(id, next); el != nil {
			return el
		}
	}
	return nil
}

func GetNodesByClassName(name string, node *html.Node) (list []*html.Node) {
	if matchClass(name, node) {
		list = append(list, node)
	}
	for next := node.FirstChild; next != nil; next = next.NextSibling {
		list = append(list, GetNodesByClassName(name, next)...)
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

func filterByAttribute(k, v string, nodes []*html.Node) (filtered []*html.Node) {
	for _, n := range nodes {
		if matchAttribute(k, v, n) {
			filtered = append(filtered, n)
		}
	}
	return
}

func filterByAttributes(attrs map[string]string, nodes []*html.Node) (filtered []*html.Node) {
	filtered = nodes
	for k, v := range attrs {
		filtered = filterByAttribute(k, v, filtered)
	}
	return
}

func GetAttribute(k string, n *html.Node) (v string) {
	if n == nil {
		return
	}
	for _, a := range n.Attr {
		if a.Key == k {
			v = a.Val
			return
		}
	}
	return
}

func matchClass(class string, n *html.Node) bool {
	for _, c := range GetClasses(n) {
		if c == class {
			return true
		}
	}
	return false
}

func GetClasses(node *html.Node) []string {
	return strings.Fields(GetAttribute("class", node))
}

func TextContent(n *html.Node) string {
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
