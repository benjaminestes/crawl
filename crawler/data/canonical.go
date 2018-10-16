package data

type Canonical struct {
	Address *Address
	Href    string
}

func MakeCanonical(base *Address, href string) *Canonical {
	c := &Canonical{
		Href:    href,
		Address: MakeAddressResolved(base, href),
	}
	return c
}
