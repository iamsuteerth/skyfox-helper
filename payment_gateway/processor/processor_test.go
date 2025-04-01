package processor

import (
	"testing"
	"time"

	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/types"
)

func TestProcessPayment(t *testing.T) {
	processor := NewPaymentProcessor()

	req := types.PaymentRequest{
		CardNumber: "4111111111111111",
		CVV:        "123",
		Expiry:     "12/25",
		Name:       "Test User",
		Amount:     100.0,
		Timestamp:  time.Now(),
	}

	successCount := 0
	failureCount := 0
	totalCalls := 20

	for i := 0; i < totalCalls; i++ {
		status, err := processor.ProcessPayment(req)
		if err != nil {
			failureCount++
			if status != "FAILED" {
				t.Errorf("Expected FAILED status on error, got %s", status)
			}
		} else {
			successCount++
			if status != "SUCCESS" {
				t.Errorf("Expected SUCCESS status, got %s", status)
			}
		}
	}

	if successCount == 0 {
		t.Error("Expected at least one successful payment, but got none")
	}

	if failureCount == 0 {
		t.Error("Expected at least one failed payment, but got none")
	}

	t.Logf("Success count: %d, Failure count: %d", successCount, failureCount)
}
