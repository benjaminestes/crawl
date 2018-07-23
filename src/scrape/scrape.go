package scrape

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func QueryNodes(name string, attrs map[string]string, n *html.Node) (list []*html.Node) {
	list = GetNodesByTagName(name, n)
	list = FilterByAttributes(attrs, list)
	return
}

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
	if MatchAttribute("name", name, node) {
		list = append(list, node)
	}
	for next := node.FirstChild; next != nil; next = next.NextSibling {
		list = append(list, GetNodesByName(name, next)...)
	}
	return list
}

func GetNodeByID(id string, node *html.Node) *html.Node {
	if MatchAttribute("id", id, node) {
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
	if MatchClass(name, node) {
		list = append(list, node)
	}
	for next := node.FirstChild; next != nil; next = next.NextSibling {
		list = append(list, GetNodesByClassName(name, next)...)
	}
	return list
}

func MatchAttribute(k, v string, n *html.Node) bool {
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

func MatchAttributes(attrs map[string]string, n *html.Node) bool {
	for k, v := range attrs {
		if !MatchAttribute(k, v, n) {
			return false
		}
	}
	return true
}

func FilterByAttribute(k, v string, nodes []*html.Node) (filtered []*html.Node) {
	for _, n := range nodes {
		if MatchAttribute(k, v, n) {
			filtered = append(filtered, n)
		}
	}
	return
}

func FilterByAttributes(attrs map[string]string, nodes []*html.Node) (filtered []*html.Node) {
	filtered = nodes
	for k, v := range attrs {
		filtered = FilterByAttribute(k, v, filtered)
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

func MatchClass(class string, n *html.Node) bool {
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

func getTextChildren(node *html.Node) (list []*html.Node) {
	if node == nil {
		return nil
	}
	if node.Type == html.TextNode {
		list = append(list, node)
	}
	for next := node.FirstChild; next != nil; next = next.NextSibling {
		list = append(list, getTextChildren(next)...)
	}
	return
}

func GetText(n *html.Node) string {
	texts := getTextChildren(n)
	var strs []string
	for _, s := range texts {
		strs = append(strs, s.Data)
	}
	return strings.Join(strs, "")
}
