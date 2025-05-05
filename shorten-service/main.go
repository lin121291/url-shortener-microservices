package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gorm.io/gorm"

	dbutils "github.com/urlshortener/common/db"
	"github.com/urlshortener/common/models"
	redisstore "github.com/urlshortener/common/redis"
)

// ShortenRequest defines the request body for shortening a URL
type ShortenRequest struct {
	LongURL    string     `json:"long_url" binding:"required,url"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CustomCode string     `json:"custom_code,omitempty"`
}

// ShortenResponse defines the response for a successful URL shortening
type ShortenResponse struct {
	Code      string    `json:"code"`
	ShortURL  string    `json:"short_url"`
	LongURL   string    `json:"long_url"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

var (
	db          *gorm.DB
	redisClient *redis.Client
	baseURL     string
)

func init() {
	var err error

	dbutils.Init()
	db = dbutils.DB

	err = db.AutoMigrate(&models.URLMapping{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	redisstore.Init()
}

func main() {
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "up",
		})
	})

	// Shorten URL endpoint
	r.POST("/shorten", handleShorten)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Shorten service starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// handleShorten creates a shortened URL
func handleShorten(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Normalize the URL if needed
	if req.LongURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL cannot be empty"})
		return
	}

	code := ""
	// If a custom code is provided, use it
	if req.CustomCode != "" {
		// Check if custom code already exists
		var existingMapping models.URLMapping
		result := db.Where("code = ?", req.CustomCode).First(&existingMapping)
		if result.RowsAffected > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Custom code already in use"})
			return
		}
		code = req.CustomCode
	} else {
		// Generate a unique short code
		code = generateShortCode()
	}

	// Set expiration time if provided
	expiresAt := time.Now().AddDate(1, 0, 0) // Default: 1 year
	if req.ExpiresAt != nil {
		expiresAt = *req.ExpiresAt
	}

	// Create URL mapping
	urlMapping := models.URLMapping{
		Code:      code,
		LongURL:   req.LongURL,
		ExpiresAt: expiresAt,
	}

	// Save to database
	if result := db.Create(&urlMapping); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create shortened URL"})
		return
	}

	// Cache in Redis
	ctx := context.Background()
	err := redisClient.Set(ctx, code, req.LongURL, time.Until(expiresAt)).Err()
	if err != nil {
		log.Printf("Warning: Failed to cache URL in Redis: %v", err)
	}

	// Create response
	shortURL := fmt.Sprintf("%s/%s", baseURL, code)
	response := ShortenResponse{
		Code:      code,
		ShortURL:  shortURL,
		LongURL:   req.LongURL,
		ExpiresAt: expiresAt,
	}

	c.JSON(http.StatusCreated, response)
}

// generateShortCode creates a unique short code for URLs
func generateShortCode() string {
	// Generate a UUID and take first 6 characters
	id := uuid.New().String()
	return id[:6]
}
