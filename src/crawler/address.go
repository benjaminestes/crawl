package crawler

import "net/url"

type Address struct {
	Address string
	*url.URL
}

func MakeAddress(u string) (a *Address) {
	a = &Address{}
	a.SetURL(u)
	return
}

func (l *Address) SetURL(u string) {
	url, err := url.Parse(u)
	if err != nil {
		// FIXME: Handle error condition
		return
	}
	if url.Path == "" {
		url.Path = "/"
	}
	l.URL = url
	l.Address = l.URL.String()
}
