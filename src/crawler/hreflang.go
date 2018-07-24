package crawler

type Hreflang struct {
	*Address
	Href     string
	Hreflang string
}

func MakeHreflang(base *Address, href, lang string) *Hreflang {
	hreflang := &Hreflang{
		Href:     href,
		Hreflang: lang,
	}
	hreflang.Address = MakeAddressFromRelative(base, href)
	return hreflang
}
