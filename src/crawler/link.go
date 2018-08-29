package crawler

type Link struct {
	Address  *Address
	Anchor   string
	Href     string
	Nofollow bool
}

func MakeLink(base *Address, href string, anchor string, nofollow bool) *Link {
	link := &Link{
		Href:     href,
		Anchor:   anchor,
		Nofollow: nofollow,
	}
	link.Address = MakeAddressFromRelative(base, href)
	return link
}

func MakeAbsoluteLink(href string, anchor string, nofollow bool) *Link {
	base := MakeAddressFromString(href)
	return MakeLink(base, href, anchor, nofollow)
}
