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
