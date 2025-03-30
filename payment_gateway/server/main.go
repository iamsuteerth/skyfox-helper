package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/types"
	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/validator"
)

func main() {
	router := gin.Default()
	validator := validator.NewStrictValidator()

	router.POST("/payment", func(c *gin.Context) {
		var req types.PaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request format"})
			return
		}

		req.Timestamp = time.Now()
		errors := validator.Validate(req)

		if len(errors) > 0 {
			c.JSON(422, gin.H{"errors": errors})
			return
		}

		c.JSON(200, gin.H{"status": "validation_passed"})
	})

	router.Run(":8080")
}
