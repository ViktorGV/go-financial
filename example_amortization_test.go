package gofinancial_test

import (
	"time"

	gofinancial "github.com/ViktorGV/go-financial"
	"github.com/ViktorGV/go-financial/enums/frequency"
	"github.com/ViktorGV/go-financial/enums/interesttype"
	"github.com/ViktorGV/go-financial/enums/paymentperiod"
	"github.com/shopspring/decimal"
)

// This example generates amortization table for a loan of 20 lakhs over 15years at 12% per annum.
func ExampleAmortization_GenerateTable() {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		panic("location loading error")
	}
	currentDate := time.Date(2009, 11, 11, 4, 30, 0, 0, loc)
	config := gofinancial.Config{

		// start date is inclusive
		StartDate: currentDate,

		// end date is inclusive.
		EndDate:   currentDate.AddDate(15, 0, 0).AddDate(0, 0, -1),
		Frequency: frequency.ANNUALLY,

		// AmountBorrowed is in paisa
		AmountBorrowed: decimal.NewFromInt(200000000),

		// InterestType can be flat or reducing
		InterestType: interesttype.REDUCING,

		// interest is in basis points
		Interest: decimal.NewFromInt(1200),

		// amount is paid at the end of the period
		PaymentPeriod: paymentperiod.ENDING,

		// all values will be rounded
		EnableRounding: true,

		// it will be rounded to nearest int
		RoundingPlaces: 0,

		// no error is tolerated
		RoundingErrorTolerance: decimal.Zero,
	}
	amortization, err := gofinancial.NewAmortization(&config)
	if err != nil {
		panic(err)
	}

	_, err = amortization.GenerateTable()
	if err != nil {
		panic(err)
	}
}
