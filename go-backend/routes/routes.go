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

        // Apply audit logging middleware
        router.Use(middleware.AuditMiddleware())

        // Initialize handlers
        organizationHandler := handlers.NewOrganizationHandler()
        projectHandler := handlers.NewProjectHandler()
        tokenHandler := handlers.NewTokenHandler(cfg)
        egressHandler := handlers.NewEgressHandler()
        ingressHandler := handlers.NewIngressHandler()
        webhookHandler := handlers.NewWebhookHandler()
        usageHandler := handlers.NewUsageHandler(cfg.UsageService, cfg.AggregatorService)
        billingHandler := handlers.NewBillingHandler(cfg.BillingService)
        teamHandler := handlers.NewTeamHandler()
        auditHandler := handlers.NewAuditHandler()
        statusHandler := handlers.NewStatusHandler()

        // Health check endpoint (no auth required)
        router.GET("/health", func(c *gin.Context) {
                c.JSON(http.StatusOK, gin.H{
                        "status":  "healthy",
                        "service": "pulse-control-plane",
                        "version": "1.0.0",
                })
        })

        // API routes with /api prefix for Kubernetes ingress routing
        api := router.Group("/api")
        {
                // API v1 routes
                v1 := api.Group("/v1")
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

                // ======= Phase 4: Usage Tracking & Billing =======
                
                // Usage routes (requires API key authentication)
                usage := v1.Group("/usage")
                usage.Use(middleware.AuthenticateProject())
                usage.Use(middleware.ProjectRateLimiter())
                {
                        usage.GET("/:project_id", usageHandler.GetUsageMetrics)
                        usage.GET("/:project_id/summary", usageHandler.GetUsageSummary)
                        usage.GET("/:project_id/aggregated", usageHandler.GetAggregatedUsage)
                        usage.GET("/:project_id/alerts", usageHandler.GetAlerts)
                        usage.POST("/:project_id/check-limits", usageHandler.CheckLimits)
                }

                // Billing routes (requires API key authentication)
                billing := v1.Group("/billing")
                billing.Use(middleware.AuthenticateProject())
                billing.Use(middleware.ProjectRateLimiter())
                {
                        billing.GET("/:project_id/dashboard", billingHandler.GetBillingDashboard)
                        billing.POST("/:project_id/invoice", billingHandler.GenerateInvoice)
                        billing.GET("/invoice/:invoice_id", billingHandler.GetInvoice)
                        billing.GET("/:project_id/invoices", billingHandler.ListInvoices)
                        billing.PUT("/invoice/:invoice_id/status", billingHandler.UpdateInvoiceStatus)
                        
                        // Stripe integration (placeholder)
                        billing.POST("/:project_id/stripe/integrate", billingHandler.IntegrateStripe)
                        billing.POST("/stripe/customer", billingHandler.CreateStripeCustomer)
                        billing.POST("/stripe/payment-method", billingHandler.AttachPaymentMethod)
                }

                // ======= Phase 5: Admin Dashboard Features =======

                // Team management routes (organization context)
                orgs := v1.Group("/organizations/:id")
                {
                        orgs.GET("/members", teamHandler.ListTeamMembers)
                        orgs.POST("/members", teamHandler.InviteTeamMember)
                        orgs.GET("/members/:user_id", teamHandler.GetTeamMember)
                        orgs.DELETE("/members/:user_id", teamHandler.RemoveTeamMember)
                        orgs.PUT("/members/:user_id/role", teamHandler.UpdateTeamMemberRole)
                        orgs.GET("/invitations", teamHandler.ListPendingInvitations)
                        orgs.DELETE("/invitations/:invitation_id", teamHandler.RevokeInvitation)
                }

                // Invitation acceptance (public)
                invitations := v1.Group("/invitations")
                {
                        invitations.POST("/accept", teamHandler.AcceptInvitation)
                }

                // Audit log routes
                auditLogs := v1.Group("/audit-logs")
                {
                        auditLogs.GET("", auditHandler.GetAuditLogs)
                        auditLogs.GET("/export", auditHandler.ExportAuditLogs)
                        auditLogs.GET("/stats", auditHandler.GetAuditStats)
                        auditLogs.GET("/recent", auditHandler.GetRecentLogs)
                }

                // Status and monitoring routes (enhanced)
                v1.GET("/status", statusHandler.GetSystemStatus)
                v1.GET("/status/projects/:id", statusHandler.GetProjectHealth)
                v1.GET("/status/regions", statusHandler.GetRegionAvailability)
        }

        // 404 handler
        router.NoRoute(func(c *gin.Context) {
                c.JSON(http.StatusNotFound, gin.H{
                        "error": "Route not found",
                        "path":  c.Request.URL.Path,
                })
        })
}
