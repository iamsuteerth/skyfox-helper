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

	errors = append(errors, validateCardNumber(req.CardNumber)...)
	errors = append(errors, validateCVV(req.CVV)...)
	errors = append(errors, validateExpiry(req.Expiry)...)
	errors = append(errors, validateName(req.Name)...)

	return errors
}

func validateCardNumber(cardNumber string) []types.ValidationError {
	var errs []types.ValidationError
	if !regexp.MustCompile(`^\d{16}$`).MatchString(cardNumber) {
		errs = append(errs, types.ValidationError{
			Field:   "card_number",
			Message: "Card number must be exactly 16 digits",
		})
		return errs
	}
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
		if i%2 == 1 {
			n *= 2
			if n > 9 {
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

	if !regexp.MustCompile(`^\d+$`).MatchString(cvv) {
		errs = append(errs, types.ValidationError{
			Field:   "cvv",
			Message: "CVV must contain only numeric characters",
		})
		return errs
	}
	if len(cvv) != 3 {
		errs = append(errs, types.ValidationError{
			Field:   "cvv",
			Message: "CVV must be exactly 3 digits",
		})
		return errs
	}
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
	val, _ := strconv.Atoi(num)
	return val
}

func validateExpiry(expiry string) []types.ValidationError {
	var errs []types.ValidationError

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

	fullYear := 2000 + year2Digit

	if fullYear > currentYear+20 {
		errs = append(errs, types.ValidationError{
			Field:   "expiry",
			Message: "Expiry cannot exceed 20 years from now",
		})
		return errs
	}

	expiryTime := time.Date(fullYear, time.Month(month+1), 0, 23, 59, 59, 999999999, time.UTC)

	if now.After(expiryTime) {
		errs = append(errs, types.ValidationError{
			Field:   "expiry",
			Message: "Card has expired",
		})
	}
	return errs
}

func validateName(name string) []types.ValidationError {
	var errs []types.ValidationError

	name = strings.TrimSpace(name)

	if len(name) == 0 {
		errs = append(errs, types.ValidationError{
			Field:   "name",
			Message: "Name cannot be empty",
		})
		return errs
	}
	if !regexp.MustCompile(`^[a-zA-Z\s'-]+$`).MatchString(name) {
		errs = append(errs, types.ValidationError{
			Field:   "name",
			Message: "Name must contain only letters, spaces, apostrophes, and hyphens",
		})
	}
	if len(name) < 2 || len(name) > 40 {
		errs = append(errs, types.ValidationError{
			Field:   "name",
			Message: "Name must be between 2-40 characters",
		})
	}
	if strings.Contains(name, "  ") {
		errs = append(errs, types.ValidationError{
			Field:   "name",
			Message: "Consecutive spaces are not allowed",
		})
	}
	return errs
}
