package types

import "time"

type PaymentRequest struct {
	CardNumber string    `json:"card_number"`
	CVV        string    `json:"cvv"`
	Expiry     string    `json:"expiry"` // Format: "MM/YY"
	Name       string    `json:"name"`
	Timestamp  time.Time `json:"-"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
