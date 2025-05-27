package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/iamsuteerth/skyfox-helper/tree/main/payment_gateway/processor"
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

func apiKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := getEnvWithDefault("API_KEY", "")
		if apiKey == "" {
			log.Warn("API_KEY environment variable not set")
			c.Next()
			return
		}
		requestApiKey := c.GetHeader("x-api-key")
		if requestApiKey == "" {
			log.WithField("client_ip", c.ClientIP()).Warn("Missing API key in request")
			c.JSON(http.StatusForbidden, gin.H{
				"status":  "FORBIDDEN",
				"message": "API key is required",
			})
			c.Abort()
			return
		}
		if requestApiKey != apiKey {
			log.WithField("client_ip", c.ClientIP()).Warn("Invalid API key provided")
			c.JSON(http.StatusForbidden, gin.H{
				"status":  "FORBIDDEN",
				"message": "Invalid API key",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func main() {
	port := getEnvWithDefault("PORT", "8082")
	apiKey := getEnvWithDefault("API_KEY", "")

	logFields := logrus.Fields{
		"port": port,
	}

	if apiKey != "" {
		logFields["api_key_protected"] = true
	} else {
		logFields["api_key_protected"] = false
		log.Warn("No API_KEY set. API endpoints are unprotected!")
	}

	log.WithFields(logFields).Info("Starting payment gateway service")

	router := gin.Default()
	validator := validator.NewStrictValidator()
	paymentProcessor := processor.NewPaymentProcessor()

	router.GET("/pshealth", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"version":   getEnvWithDefault("APP_VERSION", "dev"),
			"timestamp": time.Now().Unix(),
		})
	})

	// Added for production routes
	router.GET("/payment-service/pshealth", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"version":   getEnvWithDefault("APP_VERSION", "dev"),
			"timestamp": time.Now().Unix(),
		})
	})

	protected := router.Group("/")
	protected.Use(apiKeyAuthMiddleware())
	{
		protected.POST("/payment", func(c *gin.Context) {
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
			requestLogger = requestLogger.WithField("transaction_id", transactionID)

			status, err := paymentProcessor.ProcessPayment(req)

			processingTime := time.Since(startTime).Milliseconds()

			if err != nil {
				requestLogger.WithFields(logrus.Fields{
					"error":              err.Error(),
					"status":             status,
					"processing_time_ms": processingTime,
				}).Error("Transaction processing failed")

				c.JSON(http.StatusOK, gin.H{
					"status":         status,
					"message":        err.Error(),
					"transaction_id": transactionID,
					"request_id":     requestID,
				})
				return
			}

			requestLogger.WithFields(logrus.Fields{
				"status":             status,
				"processing_time_ms": processingTime,
			}).Info("Transaction completed successfully")

			c.JSON(http.StatusOK, gin.H{
				"status":         status,
				"message":        "Transaction processed successfully",
				"transaction_id": transactionID,
				"request_id":     requestID,
			})
		})
	}

	// Added for production routes
	protectedProd := router.Group("/payment-service")
	protectedProd.Use(apiKeyAuthMiddleware())
	{
		protectedProd.POST("/payment", func(c *gin.Context) {
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
			requestLogger = requestLogger.WithField("transaction_id", transactionID)

			status, err := paymentProcessor.ProcessPayment(req)

			processingTime := time.Since(startTime).Milliseconds()

			if err != nil {
				requestLogger.WithFields(logrus.Fields{
					"error":              err.Error(),
					"status":             status,
					"processing_time_ms": processingTime,
				}).Error("Transaction processing failed")

				c.JSON(http.StatusOK, gin.H{
					"status":         status,
					"message":        err.Error(),
					"transaction_id": transactionID,
					"request_id":     requestID,
				})
				return
			}

			requestLogger.WithFields(logrus.Fields{
				"status":             status,
				"processing_time_ms": processingTime,
			}).Info("Transaction completed successfully")

			c.JSON(http.StatusOK, gin.H{
				"status":         status,
				"message":        "Transaction processed successfully",
				"transaction_id": transactionID,
				"request_id":     requestID,
			})
		})
	}

	router.NoRoute(func(c *gin.Context) {
		log.WithFields(logrus.Fields{
			"client_ip": c.ClientIP(),
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
		}).Warn("Route not found")

		c.JSON(http.StatusNotFound, gin.H{
			"status": "NOT_FOUND",
			"error":  "There is nothing to do here! 404!",
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
