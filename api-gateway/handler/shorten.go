package handler

import (
	"bytes"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type ShortenRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	CustomAlias string `json:"custom_alias,omitempty"`
}

var (
	shortenServiceURL string
)

func init() {
	shortenServiceURL = os.Getenv("SHORTEN_SERVICE_URL")
	if shortenServiceURL == "" {
		shortenServiceURL = "http://localhost:8081"
	}
}

// handleShorten forwards the request to the shorten service
func HandleShorten(c *gin.Context) {
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
