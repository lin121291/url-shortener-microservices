package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	dbutils "github.com/urlshortener/common/db"
	"github.com/urlshortener/common/models"
	redisstore "github.com/urlshortener/common/redis"
)

// RedirectResponse is returned when getting URL info without redirecting
type RedirectResponse struct {
	LongURL   string    `json:"long_url"`
	Visits    int       `json:"visits"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

var (
	db          *gorm.DB
	redisClient *redis.Client
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

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "up",
		})
	})

	r.GET("/:code", handleRedirect)
	r.GET("/info/:code", handleURLInfo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Redirect service starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleRedirect(c *gin.Context) {
	code := c.Param("code")
	ctx := context.Background()

	longURL, err := redisClient.Get(ctx, code).Result()
	if err == redis.Nil || err != nil {
		var urlMapping models.URLMapping
		result := db.Where("code = ?", code).First(&urlMapping)

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}

		if !urlMapping.ExpiresAt.IsZero() && urlMapping.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusGone, gin.H{"error": "URL has expired"})
			return
		}

		longURL = urlMapping.LongURL

		db.Model(&urlMapping).Update("visits", urlMapping.Visits+1)

		err = redisClient.Set(ctx, code, longURL, time.Until(urlMapping.ExpiresAt)).Err()
		if err != nil {
			log.Printf("Warning: Failed to cache URL in Redis: %v", err)
		}
	} else {
		db.Model(&models.URLMapping{}).Where("code = ?", code).Update("visits", gorm.Expr("visits + 1"))
	}

	c.Redirect(http.StatusFound, longURL)
}

func handleURLInfo(c *gin.Context) {
	code := c.Param("code")

	var urlMapping models.URLMapping
	result := db.Where("code = ?", code).First(&urlMapping)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	if !urlMapping.ExpiresAt.IsZero() && urlMapping.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "URL has expired"})
		return
	}

	response := RedirectResponse{
		LongURL:   urlMapping.LongURL,
		Visits:    urlMapping.Visits,
		CreatedAt: urlMapping.CreatedAt,
		ExpiresAt: urlMapping.ExpiresAt,
	}

	c.JSON(http.StatusOK, response)
}
