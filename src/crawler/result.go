package crawler

import "net/url"

type Pair struct {
	Key string
	Val string
}

type Address struct {
	FullAddress string
	*url.URL
}

func (l *Address) SetURL(u string) {
	url, err := url.Parse(u)
	if err != nil {
		// FIXME: Handle error condition
		return
	}
	l.URL = url
	l.FullAddress = l.URL.String()
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
