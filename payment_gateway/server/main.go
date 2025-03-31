package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/database"
	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/types"
	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/validator"
)

func main() {
	region := getEnvWithDefault("AWS_REGION", "ap-south-1")
	tableName := getEnvWithDefault("DYNAMODB_TABLE", "PendingTransactions")

	router := gin.Default()
	validator := validator.NewStrictValidator()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}

	dynamoDBClient := dynamodb.New(sess)
	if err := verifyDynamoDBConnection(dynamoDBClient); err != nil {
		log.Fatalf("Failed to connect to DynamoDB: %v", err)
	}
	dynamoDBManager := database.NewDynamoDBManager(tableName, dynamoDBClient)

	router.POST("/payment", func(c *gin.Context) {
		var req types.PaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		req.Timestamp = time.Now()
		errors := validator.Validate(req)
		if len(errors) > 0 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": "REJECT",
				"errors": errors,
			})
			return
		}

		transactionID := uuid.New().String()
		cardHash := generateCardHash(req.CardNumber, req.CVV)
		transaction := types.Transaction{
			TransactionID: transactionID,
			CardHash:      cardHash,
			Timestamp:     time.Now().Unix(),
			ExpiryTime:    time.Now().Unix() + 300, // 5 minutes TTL
		}

		status, err := dynamoDBManager.ProcessTransaction(transaction)

		if err != nil {
			log.Printf("Transaction error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":      status,
				"message":     err.Error(),
				"transaction": transactionID,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":      status,
			"message":     "Transaction processed successfully",
			"transaction": transactionID,
		})
	})

	port := getEnvWithDefault("PORT", "8080")
	log.Printf("Starting server on port %s...", port)
	router.Run(":" + port)
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func verifyDynamoDBConnection(dynamoDBClient *dynamodb.DynamoDB) error {
	input := &dynamodb.ListTablesInput{}
	result, err := dynamoDBClient.ListTables(input)
	if err != nil {
		return err
	}

	log.Println("Connected to DynamoDB successfully")
	log.Println("Available tables:")
	for _, tableName := range result.TableNames {
		log.Println("- " + *tableName)
	}
	return nil
}

func generateCardHash(cardNumber, cvv string) string {
	data := fmt.Sprintf("%s:%s", cardNumber, cvv)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}
