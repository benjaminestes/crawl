package data

import "net/url"

// Address represents the useful parts of a URL we'd like to have
// available for analysis. It is the basic type which other
// address-related types embed.
type Address struct {
	Full   string
	Scheme string
	Opaque string
	Host   string
	Path   string
	Query  string
}

func MakeAddress(rawurl string) *Address {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil
	}
	return addressFromURL(u)
}

func addressFromURL(u *url.URL) *Address {
	if u.Path == "" {
		u.Path = "/"
	}
	u.Fragment = ""
	return &Address{
		Full:   u.String(),
		Scheme: u.Scheme,
		Opaque: u.Opaque,
		Host:   u.Host,
		Path:   u.EscapedPath(),
		Query:  u.RawQuery,
	}
}

func MakeAddressResolved(base *Address, rawurl string) *Address {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil
	}
	t, err := url.Parse(base.Full)
	if err != nil {
		return nil
	}
	if t != nil {
		return addressFromURL(t.ResolveReference(u))
	}
	return nil
}
