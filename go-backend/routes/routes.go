package routes

import (
        "net/http"

        "pulse-control-plane/config"
        "pulse-control-plane/handlers"
        "pulse-control-plane/middleware"

        "github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, cfg *config.Config) {
        // Apply CORS middleware
        router.Use(middleware.CORSMiddleware(cfg))

        // Apply rate limiting middleware
        router.Use(middleware.GlobalRateLimiter())

        // Initialize handlers
        organizationHandler := handlers.NewOrganizationHandler()
        projectHandler := handlers.NewProjectHandler()
        tokenHandler := handlers.NewTokenHandler(cfg)
        egressHandler := handlers.NewEgressHandler()
        ingressHandler := handlers.NewIngressHandler()
        webhookHandler := handlers.NewWebhookHandler()
        usageHandler := handlers.NewUsageHandler(cfg.UsageService, cfg.AggregatorService)
        billingHandler := handlers.NewBillingHandler(cfg.BillingService)

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

                // ======= Phase 2: Core Control Plane APIs =======

                // Organization routes
                organizations := v1.Group("/organizations")
                {
                        organizations.POST("", organizationHandler.CreateOrganization)
                        organizations.GET("", organizationHandler.ListOrganizations)
                        organizations.GET("/:id", organizationHandler.GetOrganization)
                        organizations.PUT("/:id", organizationHandler.UpdateOrganization)
                        organizations.DELETE("/:id", organizationHandler.DeleteOrganization)
                }

                // Project routes
                projects := v1.Group("/projects")
                {
                        projects.POST("", projectHandler.CreateProject)
                        projects.GET("", projectHandler.ListProjects)
                        projects.GET("/:id", projectHandler.GetProject)
                        projects.PUT("/:id", projectHandler.UpdateProject)
                        projects.DELETE("/:id", projectHandler.DeleteProject)
                        projects.POST("/:id/regenerate-keys", projectHandler.RegenerateAPIKeys)
                }

                // Token routes (requires API key authentication)
                tokens := v1.Group("/tokens")
                tokens.Use(middleware.AuthenticateProject())
                tokens.Use(middleware.ProjectRateLimiter())
                {
                        tokens.POST("/create", tokenHandler.CreateToken)
                        tokens.POST("/validate", tokenHandler.ValidateToken)
                }

                // ======= Phase 3: Media Control & Scaling =======

                // Media routes (requires API key authentication)
                media := v1.Group("/media")
                media.Use(middleware.AuthenticateProject())
                media.Use(middleware.ProjectRateLimiter())
                {
                        // Egress routes
                        egress := media.Group("/egress")
                        {
                                egress.POST("/start", egressHandler.StartEgress)
                                egress.POST("/stop", egressHandler.StopEgress)
                                egress.GET("/:id", egressHandler.GetEgress)
                                egress.GET("", egressHandler.ListEgresses)
                        }

                        // Ingress routes
                        ingress := media.Group("/ingress")
                        {
                                ingress.POST("/create", ingressHandler.CreateIngress)
                                ingress.GET("/:id", ingressHandler.GetIngress)
                                ingress.GET("", ingressHandler.ListIngresses)
                                ingress.DELETE("/:id", ingressHandler.DeleteIngress)
                        }
                }

                // Webhook routes
                webhooks := v1.Group("/webhooks")
                {
                        // Internal webhook endpoint (receives webhooks from LiveKit)
                        webhooks.POST("/livekit", webhookHandler.HandleLiveKitWebhook)
                        
                        // Webhook logs (requires authentication)
                        webhooks.GET("/logs", middleware.AuthenticateProject(), webhookHandler.GetWebhookLogs)
                }

                // ======= Phase 4: Usage Tracking (Placeholder) =======
                // usage := v1.Group("/usage")
                // usage.Use(middleware.AuthenticateProject())
                // {
                //      usage.GET("/:project_id", handlers.GetUsageMetrics)
                //      usage.GET("/:project_id/summary", handlers.GetUsageSummary)
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
