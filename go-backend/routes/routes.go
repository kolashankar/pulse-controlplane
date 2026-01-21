package routes

import (
	"net/http"

	"pulse-control-plane/config"
	"pulse-control-plane/database"
	"pulse-control-plane/handlers"
	"pulse-control-plane/middleware"
	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, cfg *config.Config) {
	// Apply security headers middleware
	router.Use(middleware.SecurityHeaders())

	// Apply CORS middleware
	router.Use(middleware.CORSMiddleware(cfg))

	// Apply HTTPS enforcement (production only)
	if cfg.Environment == "production" {
		router.Use(middleware.EnforceHTTPS())
	}

	// Apply rate limiting middleware
	router.Use(middleware.GlobalRateLimiter())

	// Apply request validation middleware
	router.Use(middleware.ValidateRequest())

	// Apply XSS protection middleware
	router.Use(middleware.PreventXSS())

	// Apply audit logging middleware
	router.Use(middleware.AuditMiddleware())

	// Initialize services
	db := database.GetDB()
	usageService := services.NewUsageService(db)
	aggregatorService := services.NewAggregatorService(db)
	billingService := services.NewBillingService(db, usageService)

	// Initialize all services
	feedService := services.NewFeedService(db)
	presenceService := services.NewPresenceService(db)
	moderationService := services.NewModerationService(db, cfg.GeminiAPIKey)
	
	// Initialize handlers
	organizationHandler := handlers.NewOrganizationHandler()
	projectHandler := handlers.NewProjectHandler()
	tokenHandler := handlers.NewTokenHandler(cfg)
	egressHandler := handlers.NewEgressHandler()
	ingressHandler := handlers.NewIngressHandler()
	webhookHandler := handlers.NewWebhookHandler()
	usageHandler := handlers.NewUsageHandler(usageService, aggregatorService)
	billingHandler := handlers.NewBillingHandler(billingService)
	teamHandler := handlers.NewTeamHandler()
	auditHandler := handlers.NewAuditHandler()
	statusHandler := handlers.NewStatusHandler()
	regionHandler := handlers.NewRegionHandler()
	analyticsHandler := handlers.NewAnalyticsHandler()
	feedHandler := handlers.NewFeedHandler(feedService)
	presenceHandler := handlers.NewPresenceHandler(presenceService)
	moderationHandler := handlers.NewModerationHandler(moderationService)

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

			// ======= Phase 8: Advanced Features =======

			// Region Management routes (Phase 8.1: Multi-Region Support)
			regions := v1.Group("/regions")
			{
				regions.GET("", regionHandler.GetAllRegions)
				regions.GET("/health", regionHandler.GetAllRegionHealth)
				regions.GET("/stats", regionHandler.GetRegionStats)
				regions.POST("/nearest", regionHandler.FindNearestRegion)
				regions.GET("/:code", regionHandler.GetRegionByCode)
				regions.GET("/:code/health", regionHandler.GetRegionHealth)
			}

			// Analytics routes (Phase 8.2: Advanced Analytics)
			analytics := v1.Group("/analytics")
			analytics.Use(middleware.AuthenticateProject())
			analytics.Use(middleware.ProjectRateLimiter())
			{
				// Custom metrics
				analytics.POST("/metrics/custom", analyticsHandler.CreateCustomMetric)
				analytics.GET("/metrics/custom/:project_id", analyticsHandler.GetCustomMetrics)

				// Alerts
				analytics.POST("/alerts", analyticsHandler.CreateAlert)
				analytics.GET("/alerts/:project_id", analyticsHandler.GetAlerts)
				analytics.POST("/alerts/:project_id/check", analyticsHandler.CheckAlerts)
				analytics.GET("/triggers/:project_id", analyticsHandler.GetRecentTriggers)

				// Real-time dashboard
				analytics.GET("/realtime/:project_id", analyticsHandler.GetRealTimeDashboard)

				// Export
				analytics.POST("/export/:project_id", analyticsHandler.ExportAnalytics)
				analytics.GET("/export/status/:export_id", analyticsHandler.GetExportStatus)

				// Forecast
				analytics.GET("/forecast/:project_id", analyticsHandler.ForecastUsage)
			}

			// ======= Phase 9: Activity Feeds & Presence (Completion Plan) =======

			// Activity Feeds routes (Phase 1)
			feeds := v1.Group("/feeds")
			feeds.Use(middleware.AuthenticateProject())
			feeds.Use(middleware.ProjectRateLimiter())
			{
				// Activity management
				feeds.POST("/activities", feedHandler.CreateActivity)
				feeds.DELETE("/activities/:activity_id", feedHandler.DeleteActivity)
				
				// Feed retrieval
				feeds.GET("/:user_id", feedHandler.GetFeedItems)
				feeds.GET("/:user_id/aggregated", feedHandler.GetAggregatedFeed)
				
				// Follow/Unfollow
				feeds.POST("/:user_id/follow", feedHandler.Follow)
				feeds.DELETE("/:user_id/unfollow", feedHandler.Unfollow)
				
				// Followers and Following
				feeds.GET("/:user_id/followers", feedHandler.GetFollowers)
				feeds.GET("/:user_id/following", feedHandler.GetFollowing)
				feeds.GET("/:user_id/stats", feedHandler.GetFollowStats)
				
				// Feed item actions
				feeds.POST("/:user_id/mark-seen", feedHandler.MarkAsSeen)
				feeds.POST("/:user_id/mark-read", feedHandler.MarkAsRead)
			}

			// Presence routes (Phase 2)
			presence := v1.Group("/presence")
			presence.Use(middleware.AuthenticateProject())
			presence.Use(middleware.ProjectRateLimiter())
			{
				// Online/Offline status
				presence.POST("/online", presenceHandler.SetOnline)
				presence.POST("/offline", presenceHandler.SetOffline)
				presence.POST("/status", presenceHandler.SetStatus)
				
				// Status retrieval
				presence.GET("/status/:user_id", presenceHandler.GetUserStatus)
				presence.POST("/bulk", presenceHandler.GetBulkStatus)
				
				// Typing indicators
				presence.POST("/typing", presenceHandler.SetTyping)
				
				// Room presence
				presence.GET("/room/:room_id", presenceHandler.GetRoomPresence)
				
				// Activity tracking
				presence.POST("/activity", presenceHandler.UpdateActivity)
				presence.GET("/activity/:user_id", presenceHandler.GetUserActivities)
			}

			// Moderation routes (Phase 3)
			moderation := v1.Group("/moderation")
			moderation.Use(middleware.AuthenticateProject())
			moderation.Use(middleware.ProjectRateLimiter())
			{
				// Content analysis
				moderation.POST("/analyze/text", moderationHandler.AnalyzeText)
				moderation.POST("/analyze/image", moderationHandler.AnalyzeImage)
				
				// Rules management
				moderation.POST("/rules", moderationHandler.CreateRule)
				moderation.GET("/rules", moderationHandler.GetRules)
				
				// Logs and stats
				moderation.GET("/logs", moderationHandler.GetLogs)
				moderation.GET("/stats", moderationHandler.GetStats)
				
				// Whitelist/Blacklist
				moderation.POST("/whitelist", moderationHandler.AddToWhitelist)
				moderation.POST("/blacklist", moderationHandler.AddToBlacklist)
				
				// Configuration
				moderation.GET("/config", moderationHandler.GetConfig)
			}

			// ======= Phase 8.3: Developer Tools =======

			// Developer tools handlers
			developerToolsHandler := handlers.NewDeveloperToolsHandler()
			ssoHandler := handlers.NewSSOHandler()
			slaHandler := handlers.NewSLAHandler()
			supportHandler := handlers.NewSupportHandler()
			deploymentHandler := handlers.NewDeploymentHandler()

			// Developer Tools routes (Phase 8.3)
			developer := v1.Group("/developer")
			{
				developer.GET("/postman-collection", developerToolsHandler.GetPostmanCollection)
				developer.GET("/openapi-spec", developerToolsHandler.GetOpenAPISpec)
				developer.GET("/sdk/go", developerToolsHandler.DownloadGoSDK)
				developer.GET("/sdk/javascript", developerToolsHandler.DownloadJavaScriptSDK)
				developer.GET("/sdk/python", developerToolsHandler.DownloadPythonSDK)
			}

			// API Documentation (Swagger UI)
			router.GET("/api/docs", developerToolsHandler.GetAPIDocumentation)

			// ======= Phase 8.4: Enterprise Features =======

			// SSO Configuration routes
			sso := v1.Group("/sso")
			{
				sso.POST("/config", ssoHandler.CreateSSOConfig)
				sso.GET("/config/:org_id", ssoHandler.GetSSOConfig)
				sso.PUT("/config/:id", ssoHandler.UpdateSSOConfig)
				sso.DELETE("/config/:id", ssoHandler.DeleteSSOConfig)
				
				// OAuth callbacks
				sso.GET("/callback/:provider", ssoHandler.OAuthCallback)
				
				// SAML callbacks
				sso.POST("/saml", ssoHandler.SAMLCallback)
			}

			// SLA Management routes
			sla := v1.Group("/sla")
			{
				// SLA Templates
				sla.POST("/templates", slaHandler.CreateSLATemplate)
				sla.GET("/templates", slaHandler.GetSLATemplates)
				
				// Organization SLA
				sla.POST("/assign", slaHandler.AssignSLAToOrg)
				sla.GET("/organization/:org_id", slaHandler.GetOrganizationSLA)
				
				// SLA Reports and Breaches
				sla.GET("/report/:org_id", slaHandler.GetSLAReport)
				sla.GET("/breaches/:org_id", slaHandler.GetSLABreaches)
			}

			// Support Ticket System routes
			support := v1.Group("/support")
			{
				// Ticket CRUD
				support.POST("/tickets", supportHandler.CreateTicket)
				support.GET("/tickets", supportHandler.ListTickets)
				support.GET("/tickets/:id", supportHandler.GetTicket)
				support.PUT("/tickets/:id", supportHandler.UpdateTicket)
				
				// Ticket assignment and comments
				support.POST("/tickets/:id/assign", supportHandler.AssignTicket)
				support.POST("/tickets/:id/comments", supportHandler.AddComment)
				support.GET("/tickets/:id/comments", supportHandler.GetTicketComments)
				
				// Statistics
				support.GET("/stats", supportHandler.GetTicketStats)
			}

			// Deployment Configuration routes
			deployment := v1.Group("/deployment")
			{
				// Deployment config CRUD
				deployment.POST("/config", deploymentHandler.CreateDeploymentConfig)
				deployment.GET("/config/:org_id", deploymentHandler.GetDeploymentConfig)
				deployment.PUT("/config/:id", deploymentHandler.UpdateDeploymentConfig)
				deployment.DELETE("/config/:id", deploymentHandler.DeleteDeploymentConfig)
				deployment.GET("/configs", deploymentHandler.ListDeploymentConfigs)
				
				// License validation
				deployment.POST("/validate-license", deploymentHandler.ValidateLicense)
				
				// Metrics
				deployment.GET("/metrics/:deployment_id", deploymentHandler.GetDeploymentMetrics)
				deployment.GET("/metrics/:deployment_id/latest", deploymentHandler.GetLatestDeploymentMetrics)
			}
		}
	} // Close api group

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Route not found",
			"path":  c.Request.URL.Path,
		})
	})
}
