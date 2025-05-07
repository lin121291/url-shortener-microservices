package handler

import (
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var (
	redirectServiceURL string
)

func init() {
	redirectServiceURL = os.Getenv("REDIRECT_SERVICE_URL")
	if redirectServiceURL == "" {
		redirectServiceURL = "http://localhost:8082"
	}
}

// handleRedirect forwards the request to the redirect service
func HandleRedirect(c *gin.Context) {
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
func HandleURLInfo(c *gin.Context) {
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
