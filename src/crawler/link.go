package crawler

type Hreflang struct {
	*Address
	Lang string
}

func MakeHreflang(address, lang string) *Hreflang {
	l := &Hreflang{
		Address: new(Address),
		Lang:    lang,
	}
	l.SetURL(address)
	return l
}

type Link struct {
	*Address
	Anchor   string
	Nofollow bool
	Internal bool // FIXME: Smart set methods for this
}

func MakeLink(address string, anchor string, nofollow bool) *Link {
	l := &Link{
		Address:  new(Address),
		Anchor:   anchor,
		Nofollow: nofollow,
	}
	l.SetURL(address)
	return l
}
