package types

import (
	"time"

	"github.com/govalues/decimal"
)

type PaymentRequest struct {
	CardNumber string          `json:"card_number" binding:"required"`
	CVV        string          `json:"cvv" binding:"required"`
	Expiry     string          `json:"expiry" binding:"required"`
	Name       string          `json:"name" binding:"required"`
	Amount     decimal.Decimal `json:"amount" binding:"required"`
	Timestamp  time.Time       `json:"-"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type PaymentResponse struct {
	Status        string `json:"status"`
	Message       string `json:"message"`
	TransactionID string `json:"transaction_id"`
	RequestID     string `json:"request_id"`
}
