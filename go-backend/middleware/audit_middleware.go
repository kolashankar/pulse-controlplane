package middleware

import (
	"context"
	"time"

	"pulse-control-plane/models"
	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuditMiddleware creates middleware that logs all actions
func AuditMiddleware() gin.HandlerFunc {
	auditService := services.NewAuditService()

	return func(c *gin.Context) {
		// Record start time
		start := time.Now()

		// Process request
		c.Next()

		// Skip logging for health check and status endpoints
		path := c.Request.URL.Path
		if path == "/health" || path == "/v1/status" {
			return
		}

		// Determine if we should log this action
		shouldLog := false
		var action, resource, resourceID, resourceName string

		method := c.Request.Method

		// Determine action and resource based on route
		switch {
		// Organization actions
		case c.FullPath() == "/v1/organizations" && method == "POST":
			action = models.AuditActions.OrganizationCreated
			resource = "organization"
			shouldLog = true
		case c.FullPath() == "/v1/organizations/:id" && method == "PUT":
			action = models.AuditActions.OrganizationUpdated
			resource = "organization"
			resourceID = c.Param("id")
			shouldLog = true
		case c.FullPath() == "/v1/organizations/:id" && method == "DELETE":
			action = models.AuditActions.OrganizationDeleted
			resource = "organization"
			resourceID = c.Param("id")
			shouldLog = true

		// Project actions
		case c.FullPath() == "/v1/projects" && method == "POST":
			action = models.AuditActions.ProjectCreated
			resource = "project"
			shouldLog = true
		case c.FullPath() == "/v1/projects/:id" && method == "PUT":
			action = models.AuditActions.ProjectUpdated
			resource = "project"
			resourceID = c.Param("id")
			shouldLog = true
		case c.FullPath() == "/v1/projects/:id" && method == "DELETE":
			action = models.AuditActions.ProjectDeleted
			resource = "project"
			resourceID = c.Param("id")
			shouldLog = true
		case c.FullPath() == "/v1/projects/:id/regenerate-keys" && method == "POST":
			action = models.AuditActions.APIKeyRegenerated
			resource = "project"
			resourceID = c.Param("id")
			shouldLog = true

		// Team member actions
		case c.FullPath() == "/v1/organizations/:id/members" && method == "POST":
			action = models.AuditActions.TeamMemberInvited
			resource = "team_member"
			shouldLog = true
		case c.FullPath() == "/v1/organizations/:id/members/:user_id" && method == "DELETE":
			action = models.AuditActions.TeamMemberRemoved
			resource = "team_member"
			resourceID = c.Param("user_id")
			shouldLog = true
		case c.FullPath() == "/v1/organizations/:id/members/:user_id/role" && method == "PUT":
			action = models.AuditActions.TeamMemberUpdated
			resource = "team_member"
			resourceID = c.Param("user_id")
			shouldLog = true

		// Settings and webhook actions
		case c.FullPath() == "/v1/projects/:id" && method == "PUT" && c.Request.ContentLength > 0:
			action = models.AuditActions.SettingsUpdated
			resource = "settings"
			resourceID = c.Param("id")
			shouldLog = true
		}

		if !shouldLog {
			return
		}

		// Get user information from context (placeholder - would come from auth)
		userEmail := c.GetString("user_email")
		if userEmail == "" {
			userEmail = "system@pulse.io"
		}

		var userID primitive.ObjectID
		if userIDStr := c.GetString("user_id"); userIDStr != "" {
			if oid, err := primitive.ObjectIDFromHex(userIDStr); err == nil {
				userID = oid
			}
		}

		// Get organization ID from context or params
		var orgID primitive.ObjectID
		if orgIDStr := c.GetString("org_id"); orgIDStr != "" {
			if oid, err := primitive.ObjectIDFromHex(orgIDStr); err == nil {
				orgID = oid
			}
		} else if orgIDStr := c.Param("id"); orgIDStr != "" {
			if oid, err := primitive.ObjectIDFromHex(orgIDStr); err == nil {
				orgID = oid
			}
		}

		// Determine status
		status := "Success"
		if c.Writer.Status() >= 400 {
			status = "Failed"
		}

		// Create audit log
		auditLog := &models.AuditLog{
			OrgID:        orgID,
			UserID:       userID,
			UserEmail:    userEmail,
			Action:       action,
			Resource:     resource,
			ResourceID:   resourceID,
			ResourceName: resourceName,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Status:       status,
			Details: map[string]interface{}{
				"method":        method,
				"path":          path,
				"status_code":   c.Writer.Status(),
				"duration_ms":   time.Since(start).Milliseconds(),
			},
		}

		// Log asynchronously to not block the response
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := auditService.LogAction(ctx, auditLog); err != nil {
				// Log error but don't fail the request
				// In production, you might want to send this to a monitoring service
				// logger.Error().Err(err).Msg("Failed to create audit log")
			}
		}()
	}
}
