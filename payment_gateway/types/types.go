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

type Transaction struct {
	TransactionID string `json:"transactionID" dynamodbav:"TransactionID"`
	CardHash      string `json:"cardHash" dynamodbav:"CardHash"`
	Timestamp     int64  `json:"timestamp" dynamodbav:"Timestamp"`
	ExpiryTime    int64  `json:"expiryTime" dynamodbav:"ExpiryTime"`
}
