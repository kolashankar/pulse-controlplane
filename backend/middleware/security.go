package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds security-related HTTP headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Strict Transport Security (HSTS)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:;")
		
		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Permissions Policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		c.Next()
	}
}

// EnforceHTTPS redirects HTTP requests to HTTPS
func EnforceHTTPS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request is HTTPS
		if c.Request.Header.Get("X-Forwarded-Proto") == "http" {
			// Redirect to HTTPS
			httpsURL := "https://" + c.Request.Host + c.Request.RequestURI
			c.Redirect(http.StatusMovedPermanently, httpsURL)
			c.Abort()
			return
		}
		c.Next()
	}
}

// ValidateRequest performs input validation to prevent injection attacks
func ValidateRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request body for validation
		contentType := c.GetHeader("Content-Type")
		
		// Validate content type for POST/PUT requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if !strings.Contains(contentType, "application/json") && 
			   !strings.Contains(contentType, "multipart/form-data") {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid content type. Expected application/json",
				})
				c.Abort()
				return
			}
		}
		
		// Validate URL parameters for common injection patterns
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if containsSQLInjection(value) || containsNoSQLInjection(value) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":     "Invalid input detected",
						"parameter": key,
						"message":   "Input contains potentially malicious patterns",
					})
					c.Abort()
					return
				}
			}
		}
		
		c.Next()
	}
}

// containsSQLInjection checks for common SQL injection patterns
func containsSQLInjection(input string) bool {
	// Common SQL injection patterns
	sqlPatterns := []string{
		`(?i)(union.*select)`,
		`(?i)(insert.*into)`,
		`(?i)(delete.*from)`,
		`(?i)(drop.*table)`,
		`(?i)(update.*set)`,
		`(?i)(exec\s*\()`,
		`(?i)(execute\s*\()`,
		`--`,
		`/\*`,
		`\*/`,
		`xp_`,
		`sp_`,
	}
	
	for _, pattern := range sqlPatterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}
	return false
}

// containsNoSQLInjection checks for common NoSQL injection patterns
func containsNoSQLInjection(input string) bool {
	// Common NoSQL injection patterns for MongoDB
	noSQLPatterns := []string{
		`\$where`,
		`\$ne`,
		`\$gt`,
		`\$lt`,
		`\$gte`,
		`\$lte`,
		`\$regex`,
		`\$or`,
		`\$and`,
		`\$nin`,
		`\$in`,
		`\$exists`,
	}
	
	for _, pattern := range noSQLPatterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}
	return false
}

// SanitizeInput removes potentially dangerous characters from input
func SanitizeInput(input string) string {
	// Remove HTML tags
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	input = htmlRegex.ReplaceAllString(input, "")
	
	// Remove script tags
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	input = scriptRegex.ReplaceAllString(input, "")
	
	// Trim whitespace
	input = strings.TrimSpace(input)
	
	return input
}

// PreventXSS adds XSS protection
func PreventXSS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set content type to prevent XSS
		if c.Request.Method == "GET" {
			c.Header("X-Content-Type-Options", "nosniff")
		}
		c.Next()
	}
}
