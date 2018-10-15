package data

import "net/url"

type Address struct {
	Full   string
	Scheme string
	Opaque string
	Host   string
	Path   string
	Query  string
}

func (a *Address) String() string {
	return a.Full
}

func (a *Address) toURL() *url.URL {
	u, _ := url.Parse(a.Full) // FIXME: use error
	return u
}

func MakeAddress(addr string) *Address {
	u, err := url.Parse(addr)
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

func MakeAddressFromRelative(base *Address, addr string) *Address {
	u, err := url.Parse(addr)
	if err != nil {
		return nil
	}
	t := base.toURL()
	if t != nil {
		return addressFromURL(t.ResolveReference(u))
	}
	return nil
}
