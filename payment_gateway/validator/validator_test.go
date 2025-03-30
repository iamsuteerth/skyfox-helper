package validator_test

import (
	"testing"
	"time"

	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/types"
	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/validator"

	"github.com/stretchr/testify/assert"
)

func TestEmptyCardNumber(t *testing.T) {
	v := validator.NewStrictValidator()
	req := types.PaymentRequest{
		CardNumber: "",
		CVV:        "123",
		Expiry:     "12/30",
		Name:       "John Doe",
		Timestamp:  time.Now(),
	}

	errors := v.Validate(req)
	assert.NotEmpty(t, errors, "Should fail on empty card number")
}

func TestCardNumberValidation(t *testing.T) {
	v := validator.NewStrictValidator()

	tests := []struct {
		name       string
		cardNumber string
		wantErrors int
	}{
		{name: "ValidCardNumber", cardNumber: "4242424242424242", wantErrors: 0},
		{name: "EmptyCardNumber", cardNumber: "", wantErrors: 1},
		{name: "ShortCardNumber", cardNumber: "123456", wantErrors: 1},
		{name: "NonDigitCardNumber", cardNumber: "4242abcd4242abcd", wantErrors: 1},
		{name: "InvalidLuhnCardNumber", cardNumber: "4242424242424241", wantErrors: 1},
		{name: "TriggersSubtractNine", cardNumber: "7992739871320001", wantErrors: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := types.PaymentRequest{
				CardNumber: tc.cardNumber,
				CVV:        "123",
				Expiry:     "12/30",
				Name:       "John Doe",
				Timestamp:  time.Now(),
			}

			errors := v.Validate(req)

			var foundCardErrors int
			for _, err := range errors {
				if err.Field == "card_number" {
					foundCardErrors++
				}
			}

			if foundCardErrors != tc.wantErrors {
				t.Errorf("Expected %d errors for card number '%s', got %d",
					tc.wantErrors, tc.cardNumber, foundCardErrors)
			}
		})
	}
}

func TestCVVValidation(t *testing.T) {
	v := validator.NewStrictValidator()

	tests := []struct {
		name       string
		cvv        string
		wantErrors int
	}{
		{name: "ValidCVV", cvv: "123", wantErrors: 0},
		{name: "EmptyCVV", cvv: "", wantErrors: 1},
		{name: "ShortCVV", cvv: "12", wantErrors: 1},
		{name: "LongCVV", cvv: "1234", wantErrors: 1},
		{name: "NonNumericCVV", cvv: "abc", wantErrors: 1},
		{name: "BoundaryCVVLow", cvv: "000", wantErrors: 1},
		{name: "BoundaryCVVHigh", cvv: "999", wantErrors: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := types.PaymentRequest{
				CardNumber: "4242424242424242",
				CVV:        tc.cvv,
				Expiry:     "12/30",
				Name:       "John Doe",
				Timestamp:  time.Now(),
			}

			errors := v.Validate(req)

			var foundCVVErrors int
			for _, err := range errors {
				if err.Field == "cvv" {
					foundCVVErrors++
				}
			}

			if foundCVVErrors != tc.wantErrors {
				t.Errorf("Expected %d errors for CVV '%s', got %d",
					tc.wantErrors, tc.cvv, foundCVVErrors)
			}
		})
	}
}

func TestExpiryDateValidation(t *testing.T) {
	v := validator.NewStrictValidator()
	now := time.Now()

	currentMonthYear := now.Format("01/06")
	nextMonth := now.AddDate(0, 1, 0).Format("01/06")
	lastMonth := now.AddDate(0, 0, -31).Format("01/06")
	twoMonthsAgo := now.AddDate(0, -2, 0).Format("01/06")

	tests := []struct {
		name       string
		expiry     string
		wantErrors int
	}{
		{name: "ValidExpiryFutureDate", expiry: "12/30", wantErrors: 0},
		{name: "EmptyExpiry", expiry: "", wantErrors: 1},
		{name: "InvalidFormatExpiry", expiry: "1230", wantErrors: 1},
		{name: "ExpiredDate", expiry: "01/20", wantErrors: 1},
		{name: "TooFutureDate", expiry: "01/50", wantErrors: 1},
		{name: "CurrentMonthYear", expiry: currentMonthYear, wantErrors: 0},
		{name: "NextMonthYear", expiry: nextMonth, wantErrors: 0},
		{name: "ExpiredYearsAgo", expiry: "12/19", wantErrors: 1},
		{name: "LastMonthExpiry", expiry: lastMonth, wantErrors: 1},
		{name: "ExpiredTwoMonthsAgo", expiry: twoMonthsAgo, wantErrors: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := types.PaymentRequest{
				CardNumber: "4242424242424242",
				CVV:        "123",
				Expiry:     tc.expiry,
				Name:       "John Doe",
				Timestamp:  time.Now(),
			}

			errors := v.Validate(req)
			var foundExpiryErrors int
			for _, err := range errors {
				if err.Field == "expiry" {
					foundExpiryErrors++
				}
			}

			if foundExpiryErrors != tc.wantErrors {
				t.Errorf("Expected %d errors for expiry '%s', got %d",
					tc.wantErrors, tc.expiry, foundExpiryErrors)
			}
		})
	}
}
