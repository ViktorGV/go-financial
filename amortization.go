package gofinancial

import (
	"time"

	"github.com/shopspring/decimal"
)

// Amortization struct holds the configuration and financial details.
type Amortization struct {
	Config    *Config
	Financial Financial
}

// NewAmortization return a new amortisation object with config and financial fields initialised.
func NewAmortization(c *Config) (*Amortization, error) {
	a := Amortization{Config: c}
	if err := a.Config.setPeriodsAndDates(); err != nil {
		return nil, err
	}
	switch a.Config.InterestType {
	case REDUCING:
		a.Financial = &Reducing{}
	case FLAT:
		a.Financial = &Flat{}
	}
	return &a, nil
}

// Row represents a single row in an amortization schedule.
type Row struct {
	Period    int64
	StartDate time.Time
	EndDate   time.Time
	Payment   decimal.Decimal
	Interest  decimal.Decimal
	Principal decimal.Decimal
}

// GenerateTable constructs the amortization table based on the configuration.
func (a Amortization) GenerateTable() ([]Row, error) {
	var result []Row
	for i := int64(1); i <= a.Config.periods; i++ {
		var row Row
		row.Period = i
		row.StartDate = a.Config.startDates[i-1]
		row.EndDate = a.Config.endDates[i-1]

		payment := a.Financial.GetPayment(*a.Config)
		principalPayment := a.Financial.GetPrincipal(*a.Config, i)
		interestPayment := a.Financial.GetInterest(*a.Config, i)
		if a.Config.EnableRounding {
			row.Payment = payment.Round(a.Config.RoundingPlaces)
			row.Principal = principalPayment.Round(a.Config.RoundingPlaces)
			// to avoid rounding errors.
			row.Interest = row.Payment.Sub(row.Principal)
		} else {
			row.Payment = payment
			row.Principal = principalPayment
			row.Interest = interestPayment
		}
		if i == a.Config.periods {
			DoPrincipalAdjustmentDueToRounding(&row, result, a.Config.AmountBorrowed, a.Config.EnableRounding, a.Config.RoundingPlaces)
		}
		if err := sanityCheckUpdate(&row, a.Config.RoundingErrorTolerance); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, nil
}

// DoPrincipalAdjustmentDueToRounding takes care of errors in total principal to be collected and adjusts it against the
// the final principal and payment amount.
func DoPrincipalAdjustmentDueToRounding(finalRow *Row, rows []Row, principal decimal.Decimal, round bool, places int32) {
	principalCollected := finalRow.Principal
	for _, row := range rows {
		principalCollected = principalCollected.Add(row.Principal)
	}
	diff := principal.Abs().Sub(principalCollected.Abs())
	if round {
		// subtracting diff coz payment, principal and interest are -ve.
		finalRow.Payment = finalRow.Payment.Sub(diff).Round(places)
		finalRow.Principal = finalRow.Principal.Sub(diff).Round(places)
	} else {
		finalRow.Payment = finalRow.Payment.Sub(diff)
		finalRow.Principal = finalRow.Principal.Sub(diff)
	}
}

// sanityCheckUpdate verifies the equation,
// payment = principal + interest for every row.
// If there is a mismatch due to rounding error and it is withing the tolerance,
// the difference is adjusted against the interest.
func sanityCheckUpdate(row *Row, tolerance decimal.Decimal) error {
	if !row.Payment.Equal(row.Principal.Add(row.Interest)) {
		diff := row.Payment.Abs().Sub(row.Principal.Add(row.Interest).Abs())
		if diff.LessThanOrEqual(tolerance) {
			row.Interest = row.Interest.Sub(diff)
		} else {
			return ErrPayment
		}
	}
	return nil
}
