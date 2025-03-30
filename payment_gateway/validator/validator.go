package validator

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/types"
)

type PaymentValidator interface {
	Validate(req types.PaymentRequest) []types.ValidationError
}

type StrictValidator struct{}

func NewStrictValidator() *StrictValidator {
	return &StrictValidator{}
}

func (v *StrictValidator) Validate(req types.PaymentRequest) []types.ValidationError {
	var errors []types.ValidationError

	// Card number validation
	errors = append(errors, validateCardNumber(req.CardNumber)...)

	// CVV validation
	errors = append(errors, validateCVV(req.CVV)...)

	// Expiry validation
	errors = append(errors, validateExpiry(req.Expiry)...)

	// Other validations (to be implemented later)

	return errors
}

func validateCardNumber(cardNumber string) []types.ValidationError {
	var errs []types.ValidationError

	// Ensure card number is digits only
	if !regexp.MustCompile(`^\d{16}$`).MatchString(cardNumber) {
		errs = append(errs, types.ValidationError{
			Field:   "card_number",
			Message: "Card number must be exactly 16 digits",
		})
		return errs
	}

	// Luhn algorithm check
	if !isValidLuhn(cardNumber) {
		errs = append(errs, types.ValidationError{
			Field:   "card_number",
			Message: "Card number failed Luhn check",
		})
	}

	return errs
}

func isValidLuhn(cardNumber string) bool {
	var sum int
	reversed := reverseString(cardNumber)

	for i, digit := range reversed {
		n := int(digit - '0')
		if i%2 == 1 { // Double every second digit
			n *= 2
			if n > 9 { // Subtract 9 if the result is greater than 9
				n -= 9
			}
		}
		sum += n
	}

	return sum%10 == 0
}

func reverseString(s string) string {
	var reversed strings.Builder
	for i := len(s) - 1; i >= 0; i-- {
		reversed.WriteByte(s[i])
	}
	return reversed.String()
}

func validateCVV(cvv string) []types.ValidationError {
	var errs []types.ValidationError

	// CVV should be numeric
	if !regexp.MustCompile(`^\d+$`).MatchString(cvv) {
		errs = append(errs, types.ValidationError{
			Field:   "cvv",
			Message: "CVV must contain only numeric characters",
		})
		return errs
	}

	// CVV should be 3 digits
	if len(cvv) != 3 {
		errs = append(errs, types.ValidationError{
			Field:   "cvv",
			Message: "CVV must be exactly 3 digits",
		})
		return errs
	}

	// CVV should be between 001 and 999
	cvvInt := atoi(cvv)
	if cvvInt < 1 || cvvInt > 999 {
		errs = append(errs, types.ValidationError{
			Field:   "cvv",
			Message: "CVV must be between 001 and 999",
		})
	}

	return errs
}

func atoi(num string) int {
	val, _ := strconv.Atoi(num) // Conversion without crashing
	return val
}

func validateExpiry(expiry string) []types.ValidationError {
	var errs []types.ValidationError

	// Format validation
	if !regexp.MustCompile(`^(0[1-9]|1[0-2])/(\d{2})$`).MatchString(expiry) {
		errs = append(errs, types.ValidationError{
			Field:   "expiry",
			Message: "Expiry must be in MM/YY format with valid month (01-12)",
		})
		return errs
	}

	parts := strings.Split(expiry, "/")
	month, _ := strconv.Atoi(parts[0])
	year2Digit, _ := strconv.Atoi(parts[1])

	now := time.Now().UTC()
	currentYear := now.Year()

	// Year calculation with proper century handling
	fullYear := currentYear/100*100 + year2Digit
	if fullYear < currentYear {
		fullYear += 100
	}

	// 20-year validation
	if fullYear > currentYear+20 {
		errs = append(errs, types.ValidationError{
			Field:   "expiry",
			Message: "Expiry cannot exceed 20 years from now",
		})
		return errs // Early return for unrecoverable error
	}

	// Calculate expiry moment (last nanosecond of the month)
	expiryTime := time.Date(fullYear, time.Month(month), 1, 0, 0, 0, 0, time.UTC).
		AddDate(0, 1, 0).     // First day of next month
		Add(-time.Nanosecond) // Last moment of expiry month

	// Time comparison (UTC)
	if now.After(expiryTime) {
		errs = append(errs, types.ValidationError{
			Field:   "expiry",
			Message: "Card has expired",
		})
	}

	return errs
}
