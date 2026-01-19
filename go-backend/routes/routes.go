package routes

import (
	"net/http"

	"pulse-control-plane/config"
	"pulse-control-plane/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, cfg *config.Config) {
	// Apply CORS middleware
	router.Use(middleware.CORSMiddleware(cfg))

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "pulse-control-plane",
			"version": "1.0.0",
		})
	})

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Public routes (no authentication)
		v1.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "operational",
				"message": "Pulse Control Plane is running",
			})
		})

		// Organization routes (will be implemented in Phase 2)
		// organizations := v1.Group("/organizations")
		// {
		// 	organizations.POST("", handlers.CreateOrganization)
		// 	organizations.GET("", handlers.ListOrganizations)
		// 	organizations.GET("/:id", handlers.GetOrganization)
		// 	organizations.PUT("/:id", handlers.UpdateOrganization)
		// 	organizations.DELETE("/:id", handlers.DeleteOrganization)
		// }

		// Project routes (will be implemented in Phase 2)
		// projects := v1.Group("/projects")
		// {
		// 	projects.POST("", handlers.CreateProject)
		// 	projects.GET("", handlers.ListProjects)
		// 	projects.GET("/:id", handlers.GetProject)
		// 	projects.PUT("/:id", handlers.UpdateProject)
		// 	projects.DELETE("/:id", handlers.DeleteProject)
		// 	projects.POST("/:id/regenerate-keys", handlers.RegenerateAPIKeys)
		// }

		// Token routes (will be implemented in Phase 2)
		// Requires API key authentication
		// tokens := v1.Group("/tokens")
		// tokens.Use(middleware.AuthenticateProject())
		// {
		// 	tokens.POST("/create", handlers.CreateToken)
		// 	tokens.POST("/validate", handlers.ValidateToken)
		// }

		// Media routes (will be implemented in Phase 3)
		// media := v1.Group("/media")
		// media.Use(middleware.AuthenticateProject())
		// {
		// 	egress := media.Group("/egress")
		// 	{
		// 		egress.POST("/start", handlers.StartEgress)
		// 		egress.POST("/stop", handlers.StopEgress)
		// 		egress.GET("/:id", handlers.GetEgressStatus)
		// 	}

		// 	ingress := media.Group("/ingress")
		// 	{
		// 		ingress.POST("/create", handlers.CreateIngress)
		// 		ingress.DELETE("/:id", handlers.DeleteIngress)
		// 	}
		// }

		// Webhook routes (will be implemented in Phase 3)
		// webhooks := v1.Group("/webhooks")
		// {
		// 	webhooks.POST("/internal", handlers.HandleSystemWebhook)
		// }

		// Usage routes (will be implemented in Phase 4)
		// usage := v1.Group("/usage")
		// usage.Use(middleware.AuthenticateProject())
		// {
		// 	usage.GET("/:project_id", handlers.GetUsageMetrics)
		// 	usage.GET("/:project_id/summary", handlers.GetUsageSummary)
		// }
	}

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Route not found",
			"path":  c.Request.URL.Path,
		})
	})
}
