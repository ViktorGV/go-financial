package gofinancial

type INTEREST_TYPE uint8

const (
	FLAT INTEREST_TYPE = iota + 1
	REDUCING
)

var toString = map[INTEREST_TYPE]string{
	FLAT:     "flat",
	REDUCING: "reducing",
}

func (t INTEREST_TYPE) String() string {
	return toString[t]
}
