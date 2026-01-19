package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// VerifyWebhookSignature validates webhook signatures from external services
func VerifyWebhookSignature(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get signature from header
		signature := c.GetHeader("X-Webhook-Signature")
		if signature == "" {
			// Try alternative header names
			signature = c.GetHeader("X-Hub-Signature-256")
		}
		
		if signature == "" {
			log.Warn().Msg("Webhook signature missing")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Webhook signature is required",
			})
			c.Abort()
			return
		}
		
		// Read request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read webhook body")
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read request body",
			})
			c.Abort()
			return
		}
		
		// Verify signature
		if !verifySignature(body, signature, secret) {
			log.Warn().Str("signature", signature).Msg("Invalid webhook signature")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid webhook signature",
			})
			c.Abort()
			return
		}
		
		// Store body in context for handlers to use
		c.Set("webhook_body", body)
		
		log.Debug().Msg("Webhook signature verified successfully")
		c.Next()
	}
}

// verifySignature verifies HMAC SHA256 signature
func verifySignature(payload []byte, signature, secret string) bool {
	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	
	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// GenerateWebhookSignature generates a signature for outgoing webhooks
func GenerateWebhookSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
