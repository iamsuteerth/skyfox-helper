package processor

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/types"
)

type PaymentProcessor struct{}

func NewPaymentProcessor() *PaymentProcessor {
	return &PaymentProcessor{}
}

func (p *PaymentProcessor) ProcessPayment(req types.PaymentRequest) (string, error) {
	processingDelay := 400 + rand.Intn(401)
	time.Sleep(time.Duration(processingDelay) * time.Millisecond)

	if rand.Float64() < 0.1 {
		return "FAILED", fmt.Errorf("payment declined by the issuing bank")
	}

	return "SUCCESS", nil
}
