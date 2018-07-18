package crawler

import "net/url"

type Address struct {
	FullAddress string
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
	l.URL = url
	l.FullAddress = l.URL.String()
}
