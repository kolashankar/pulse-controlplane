package middleware

import (
	"context"
	"net/http"
	"strings"

	"pulse-control-plane/database"
	"pulse-control-plane/models"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// AuthenticateProject validates the Pulse API Key from request header
func AuthenticateProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from header
		apiKey := c.GetHeader("X-Pulse-Key")
		if apiKey == "" {
			apiKey = c.GetHeader("Authorization")
			if apiKey != "" {
				// Remove "Bearer " prefix if present
				apiKey = strings.TrimPrefix(apiKey, "Bearer ")
			}
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing API key. Please provide X-Pulse-Key header or Authorization Bearer token",
			})
			c.Abort()
			return
		}

		// Query database for project with this API key
		projectCollection := database.GetCollection("projects")
		var project models.Project

		err := projectCollection.FindOne(context.Background(), bson.M{
			"pulse_api_key": apiKey,
			"is_deleted":    false,
		}).Decode(&project)

		if err != nil {
			log.Warn().Str("api_key", apiKey).Msg("Invalid API key")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}

		// Store project information in context
		c.Set("project_id", project.ID.Hex())
		c.Set("project", project)
		c.Set("org_id", project.OrgID.Hex())

		log.Debug().Str("project_id", project.ID.Hex()).Str("project_name", project.Name).Msg("Project authenticated")

		c.Next()
	}
}

// AuthenticateAPISecret validates both API key and secret for sensitive operations
func AuthenticateAPISecret() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-Pulse-Key")
		apiSecret := c.GetHeader("X-Pulse-Secret")

		if apiKey == "" || apiSecret == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing API key or secret. Please provide both X-Pulse-Key and X-Pulse-Secret headers",
			})
			c.Abort()
			return
		}

		// Query database
		projectCollection := database.GetCollection("projects")
		var project models.Project

		err := projectCollection.FindOne(context.Background(), bson.M{
			"pulse_api_key": apiKey,
			"is_deleted":    false,
		}).Decode(&project)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}

		// Verify API secret (hashed)
		if err := bcrypt.CompareHashAndPassword([]byte(project.PulseAPISecret), []byte(apiSecret)); err != nil {
			log.Warn().Str("project_id", project.ID.Hex()).Msg("Invalid API secret")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API secret",
			})
			c.Abort()
			return
		}

		// Store project information in context
		c.Set("project_id", project.ID.Hex())
		c.Set("project", project)
		c.Set("org_id", project.OrgID.Hex())

		c.Next()
	}
}

// RequireOrganization ensures the user has access to the specified organization
func RequireOrganization() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgIDParam := c.Param("org_id")
		if orgIDParam == "" {
			orgIDParam = c.Param("id")
		}

		if orgIDParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Organization ID is required",
			})
			c.Abort()
			return
		}

		// Validate ObjectID format
		orgID, err := primitive.ObjectIDFromHex(orgIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid organization ID format",
			})
			c.Abort()
			return
		}

		// Check if organization exists
		orgCollection := database.GetCollection("organizations")
		var org models.Organization

		err = orgCollection.FindOne(context.Background(), bson.M{
			"_id":        orgID,
			"is_deleted": false,
		}).Decode(&org)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Organization not found",
			})
			c.Abort()
			return
		}

		// Store organization in context
		c.Set("organization", org)
		c.Set("org_id", orgID.Hex())

		c.Next()
	}
}
