package crawler

import "net/url"

type Pair struct {
	Key string
	Val string
}

type Address struct {
	Full string
	*url.URL
}

type Node struct {
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
	Links []*Address
	*Address
}
