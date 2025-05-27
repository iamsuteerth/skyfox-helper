package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iamsuteerth/skyfox-helper/tree/main/movie_service/internal/services"
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
	port := getEnvWithDefault("PORT", "4567")
	apiKey := getEnvWithDefault("API_KEY", "")
	dataPath := getEnvWithDefault("MOVIES_DATA_PATH", "data/movies.json")

	logFields := logrus.Fields{
		"port": port,
	}

	if apiKey != "" {
		logFields["api_key_protected"] = true
	} else {
		logFields["api_key_protected"] = false
		log.Warn("No API_KEY set. API endpoints are unprotected!")
	}

	log.WithFields(logFields).Info("Starting movie service")

	movieService, err := services.NewMovieService(dataPath)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize movie service")
	}

	router := gin.Default()
	router.Use(gin.Recovery())

	router.GET("/mshealth", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"version":   getEnvWithDefault("APP_VERSION", "dev"),
			"timestamp": time.Now().Unix(),
		})
	})

	router.GET("/movie-service/mshealth", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"version":   getEnvWithDefault("APP_VERSION", "dev"),
			"timestamp": time.Now().Unix(),
		})
	})

	protected := router.Group("/")
	protected.Use(apiKeyAuthMiddleware())
	{
		protected.GET("/movies", func(c *gin.Context) {
			requestLogger := log.WithFields(logrus.Fields{
				"client_ip": c.ClientIP(),
				"method":    c.Request.Method,
				"path":      c.Request.URL.Path,
			})
			requestLogger.Info("Received request for all movies")

			c.JSON(http.StatusOK, movieService.GetAllMovies())
		})

		protected.GET("/movies/:id", func(c *gin.Context) {
			id := c.Param("id")

			requestLogger := log.WithFields(logrus.Fields{
				"client_ip": c.ClientIP(),
				"method":    c.Request.Method,
				"path":      c.Request.URL.Path,
				"movie_id":  id,
			})
			requestLogger.Info("Received request for specific movie")

			movie, found := movieService.GetMovieByID(id)
			if !found {
				requestLogger.Warn("Movie not found")
				c.JSON(http.StatusNotFound, gin.H{
					"status": "NOT_FOUND",
					"error":  "Movie with requested ID not found",
				})
				return
			}

			c.JSON(http.StatusOK, movie)
		})
	}

	protectedProd := router.Group("/movie-service")
	protectedProd.Use(apiKeyAuthMiddleware())
	{
		protectedProd.GET("/movies", func(c *gin.Context) {
			requestLogger := log.WithFields(logrus.Fields{
				"client_ip": c.ClientIP(),
				"method":    c.Request.Method,
				"path":      c.Request.URL.Path,
			})
			requestLogger.Info("Received request for all movies")

			c.JSON(http.StatusOK, movieService.GetAllMovies())
		})

		protectedProd.GET("/movies/:id", func(c *gin.Context) {
			id := c.Param("id")

			requestLogger := log.WithFields(logrus.Fields{
				"client_ip": c.ClientIP(),
				"method":    c.Request.Method,
				"path":      c.Request.URL.Path,
				"movie_id":  id,
			})
			requestLogger.Info("Received request for specific movie")

			movie, found := movieService.GetMovieByID(id)
			if !found {
				requestLogger.Warn("Movie not found")
				c.JSON(http.StatusNotFound, gin.H{
					"status": "NOT_FOUND",
					"error":  "Movie with requested ID not found",
				})
				return
			}

			c.JSON(http.StatusOK, movie)
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
