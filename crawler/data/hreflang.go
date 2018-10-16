package data

type Hreflang struct {
	Address  *Address
	Href     string
	Hreflang string
}

func MakeHreflang(base *Address, href, lang string) *Hreflang {
	hreflang := &Hreflang{
		Href:     href,
		Hreflang: lang,
		Address:  MakeAddressResolved(base, href),
	}
	return hreflang
}
