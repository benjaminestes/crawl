package crawler

type Canonical struct {
	*Address
	Href string
}

func MakeCanonical(base *Address, href string) *Canonical {
	c := new(Canonical)
	c.Href = href
	c.Address = MakeAddressFromRelative(base, href)
	return c
}
