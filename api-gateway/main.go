package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var (
	shortenServiceURL  string
	redirectServiceURL string
)

func init() {
	shortenServiceURL = os.Getenv("SHORTEN_SERVICE_URL")
	if shortenServiceURL == "" {
		shortenServiceURL = "http://localhost:8081"
	}

	redirectServiceURL = os.Getenv("REDIRECT_SERVICE_URL")
	if redirectServiceURL == "" {
		redirectServiceURL = "http://localhost:8082"
	}
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

	// Redirect URL endpoint
	r.GET("/:code", handleRedirect)
	r.GET("/info/:code", handleURLInfo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API Gateway starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// handleShorten forwards the request to the shorten service
func handleShorten(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Create a new request to the shorten service
	req, err := http.NewRequest("POST", shortenServiceURL+"/shorten", bytes.NewBuffer(data))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to shorten service"})
		return
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// Set response headers and status code
	for k, v := range resp.Header {
		c.Writer.Header()[k] = v
	}
	c.Status(resp.StatusCode)
	c.Writer.Write(respBody)
}

// handleRedirect forwards the request to the redirect service
func handleRedirect(c *gin.Context) {
	code := c.Param("code")

	// Forward to redirect service
	resp, err := http.Get(redirectServiceURL + "/" + code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to redirect service"})
		return
	}
	defer resp.Body.Close()

	// If it's a redirect
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		location := resp.Header.Get("Location")
		if location != "" {
			c.Redirect(resp.StatusCode, location)
			return
		}
	}

	// Otherwise, copy the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// Set response headers
	for k, v := range resp.Header {
		c.Writer.Header()[k] = v
	}
	c.Status(resp.StatusCode)
	c.Writer.Write(respBody)
}

// handleURLInfo gets URL information without redirecting
func handleURLInfo(c *gin.Context) {
	code := c.Param("code")

	// Forward to redirect service info endpoint
	resp, err := http.Get(redirectServiceURL + "/info/" + code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to redirect service"})
		return
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// Set response headers
	for k, v := range resp.Header {
		c.Writer.Header()[k] = v
	}
	c.Status(resp.StatusCode)
	c.Writer.Write(respBody)
}
