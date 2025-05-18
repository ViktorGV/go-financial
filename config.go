package gofinancial

import (
	"time"

	"github.com/shopspring/decimal"
)

// Config is used to store details used in generation of amortization table.
type Config struct {
	StartDate              time.Time           // Starting day of the amortization schedule(inclusive)
	EndDate                time.Time           // Ending day of the amortization schedule(inclusive)
	Frequency              FREQUENCY_TYPE      // Frequency enum with DAILY, WEEKLY, MONTHLY or ANNUALLY
	AmountBorrowed         decimal.Decimal     // Amount Borrowed
	InterestType           INTEREST_TYPE       // InterestType enum with FLAT or REDUCING value.
	Interest               decimal.Decimal     // Interest in basis points
	PaymentPeriod          PAYMENT_PERIOD_TYPE // Payment period enum to know whether payment made at the BEGINNING or ENDING of a period
	EnableRounding         bool                // If enabled, the final values in amortization schedule are rounded
	RoundingPlaces         int32               // If specified, the final values in amortization schedule are rounded to these many places
	RoundingErrorTolerance decimal.Decimal     // Any difference in [payment-(principal+interest)] will be adjusted in interest component, upto the RoundingErrorTolerance value specified
	periods                int64               // derived
	startDates             []time.Time         // derived
	endDates               []time.Time         // derived
}

func (c *Config) setPeriodsAndDates() error {
	sy, sm, sd := c.StartDate.Date()
	startDate := time.Date(sy, sm, sd, 0, 0, 0, 0, c.StartDate.Location())

	ey, em, ed := c.EndDate.Date()
	endDate := time.Date(ey, em, ed, 0, 0, 0, 0, c.EndDate.Location())

	period, err := GetPeriodDifference(startDate, endDate, c.Frequency)
	if err != nil {
		return err
	}
	c.periods = int64(period)
	for i := 0; i < period; i++ {
		date, err := getStartDate(startDate, c.Frequency, i)
		if err != nil {
			return err
		}
		if i == 0 {
			c.startDates = append(c.startDates, c.StartDate)
		} else {
			c.startDates = append(c.startDates, date)
		}
		if endDate, err := getEndDates(date, c.Frequency); err != nil {
			return err
		} else {
			c.endDates = append(c.endDates, endDate)
		}
	}
	return nil
}

func GetPeriodDifference(from time.Time, to time.Time, freq FREQUENCY_TYPE) (int, error) {
	var periods int
	switch freq {
	case DAILY:
		periods = int(to.Sub(from).Hours()/24) + 1
	case WEEKLY:
		days := int(to.Sub(from).Hours()/24) + 1
		if days%7 != 0 {
			return -1, ErrUnevenEndDate
		}
		periods = days / 7
	case MONTHLY:
		months, err := getMonthsBetweenDates(from, to)
		if err != nil {
			return -1, err
		}
		periods = *months
	case ANNUALLY:
		years, err := getYearsBetweenDates(from, to)
		if err != nil {
			return -1, err
		}
		periods = *years
	default:
		return -1, ErrInvalidFrequency
	}
	return periods, nil
}

func getStartDate(date time.Time, freq FREQUENCY_TYPE, index int) (time.Time, error) {
	var startDate time.Time
	switch freq {
	case DAILY:
		startDate = date.AddDate(0, 0, index)
	case WEEKLY:
		startDate = date.AddDate(0, 0, 7*index)
	case MONTHLY:
		startDate = date.AddDate(0, index, 0)
	case ANNUALLY:
		startDate = date.AddDate(index, 0, 0)
	default:
		return time.Time{}, ErrInvalidFrequency
	}
	return startDate, nil
}

func getMonthsBetweenDates(start time.Time, end time.Time) (*int, error) {
	count := 0
	for start.Before(end) {
		start = start.AddDate(0, 1, 0)
		count++
	}
	finalDate := start.AddDate(0, 0, -1)
	if !finalDate.Equal(end) {
		return nil, ErrUnevenEndDate
	}
	return &count, nil
}

func getYearsBetweenDates(start time.Time, end time.Time) (*int, error) {
	count := 0
	for start.Before(end) {
		start = start.AddDate(1, 0, 0)
		count++
	}
	finalDate := start.AddDate(0, 0, -1)
	if !finalDate.Equal(end) {
		return nil, ErrUnevenEndDate
	}
	return &count, nil
}

func getEndDates(date time.Time, freq FREQUENCY_TYPE) (time.Time, error) {
	var nextDate time.Time
	switch freq {
	case DAILY:
		nextDate = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, date.Location())
	case WEEKLY:
		date = date.AddDate(0, 0, 6)
		nextDate = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, date.Location())
	case MONTHLY:
		date = date.AddDate(0, 1, 0).AddDate(0, 0, -1)
		nextDate = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, date.Location())
	case ANNUALLY:
		date = date.AddDate(1, 0, 0).AddDate(0, 0, -1)
		nextDate = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, date.Location())
	default:
		return time.Time{}, ErrInvalidFrequency
	}
	return nextDate, nil
}

func (c *Config) getInterestRatePerPeriodInDecimal() decimal.Decimal {
	hundred := decimal.NewFromInt(100)
	freq := decimal.NewFromInt(int64(c.Frequency.Value()))
	interestInPercent := c.Interest.Div(hundred)
	InterestInDecimal := interestInPercent.Div(hundred)
	InterestPerPeriod := InterestInDecimal.Div(freq)
	return InterestPerPeriod
}
