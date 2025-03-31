package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/database"
	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/types"
	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/validator"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)

	if os.Getenv("LOG_LEVEL") == "debug" {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
}

func main() {
	region := getEnvWithDefault("AWS_REGION", "ap-south-1")
	tableName := getEnvWithDefault("DYNAMODB_TABLE", "PendingTransactions")
	port := getEnvWithDefault("PORT", "8080")

	log.WithFields(logrus.Fields{
		"region": region,
		"table":  tableName,
		"port":   port,
	}).Info("Starting payment gateway service")

	router := gin.Default()
	validator := validator.NewStrictValidator()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to create AWS session")
	}

	dynamoDBClient := dynamodb.New(sess)

	if err := verifyDynamoDBConnection(dynamoDBClient); err != nil {
		log.WithError(err).Fatal("Failed to connect to DynamoDB")
	}

	dynamoDBManager := database.NewDynamoDBManager(tableName, dynamoDBClient)

	router.GET("/health", func(c *gin.Context) {
		err := verifyDynamoDBConnection(dynamoDBClient)
		if err != nil {
			log.WithError(err).Error("Health check failed: DynamoDB connection issue")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "unhealthy",
				"message": "Database connection issue",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"version":   getEnvWithDefault("APP_VERSION", "dev"),
			"timestamp": time.Now().Unix(),
		})
	})

	router.POST("/payment", func(c *gin.Context) {
		requestID := uuid.New().String()
		startTime := time.Now()

		requestLogger := log.WithFields(logrus.Fields{
			"request_id": requestID,
			"client_ip":  c.ClientIP(),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
		})

		requestLogger.Info("Received payment request")

		var req types.PaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			requestLogger.WithError(err).Warn("Invalid request format")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":      "Invalid request format",
				"request_id": requestID,
			})
			return
		}

		req.Timestamp = time.Now()
		errors := validator.Validate(req)
		if len(errors) > 0 {
			requestLogger.WithField("validation_errors", errors).Warn("Validation failed")
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status":     "REJECT",
				"errors":     errors,
				"request_id": requestID,
			})
			return
		}

		transactionID := uuid.New().String()
		cardHash := generateCardHash(req.CardNumber, req.CVV)

		requestLogger = requestLogger.WithFields(logrus.Fields{
			"transaction_id": transactionID,
			"card_hash":      cardHash[:8] + "****",
		})

		transaction := types.Transaction{
			TransactionID: transactionID,
			CardHash:      cardHash,
			Timestamp:     time.Now().Unix(),
			ExpiryTime:    time.Now().Unix() + 300,
		}

		status, err := dynamoDBManager.ProcessTransaction(transaction)

		processingTime := time.Since(startTime).Milliseconds()

		if err != nil {
			requestLogger.WithFields(logrus.Fields{
				"error":              err.Error(),
				"status":             status,
				"processing_time_ms": processingTime,
			}).Error("Transaction processing failed")

			c.JSON(http.StatusOK, gin.H{
				"status":      status,
				"message":     err.Error(),
				"transaction": transactionID,
				"request_id":  requestID,
			})
			return
		}

		requestLogger.WithFields(logrus.Fields{
			"status":             status,
			"processing_time_ms": processingTime,
		}).Info("Transaction completed successfully")

		// Transaction succeeded
		c.JSON(http.StatusOK, gin.H{
			"status":      status,
			"message":     "Transaction processed successfully",
			"transaction": transactionID,
			"request_id":  requestID,
		})
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.WithField("port", port).Info("Server is running")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("Server forced to shutdown")
	}

	log.Info("Server exited gracefully")
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

	tables := make([]string, 0, len(result.TableNames))
	for _, tableName := range result.TableNames {
		tables = append(tables, *tableName)
	}

	log.WithField("tables", tables).Info("Connected to DynamoDB successfully")
	return nil
}

func generateCardHash(cardNumber, cvv string) string {
	data := fmt.Sprintf("%s:%s", cardNumber, cvv)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}
