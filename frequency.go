package gofinancial

type FREQUENCY_TYPE uint8

const (
	DAILY FREQUENCY_TYPE = iota + 1
	WEEKLY
	MONTHLY
	ANNUALLY
)

var toValue = map[FREQUENCY_TYPE]int{
	DAILY:    360,
	WEEKLY:   52,
	MONTHLY:  12,
	ANNUALLY: 1,
}

func (t *FREQUENCY_TYPE) Value() int {
	return toValue[*t]
}
