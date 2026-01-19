package handlers

import (
	"net/http"
	"strconv"
	"time"

	"pulse-control-plane/models"
	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UsageHandler handles usage-related HTTP requests
type UsageHandler struct {
	usageService      *services.UsageService
	aggregatorService *services.AggregatorService
}

// NewUsageHandler creates a new usage handler
func NewUsageHandler(usageService *services.UsageService, aggregatorService *services.AggregatorService) *UsageHandler {
	return &UsageHandler{
		usageService:      usageService,
		aggregatorService: aggregatorService,
	}
}

// GetUsageMetrics retrieves usage metrics for a project
// GET /v1/usage/:project_id
func (h *UsageHandler) GetUsageMetrics(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get query parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Parse dates
	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (use YYYY-MM-DD)"})
			return
		}
	} else {
		// Default to current month
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (use YYYY-MM-DD)"})
			return
		}
	} else {
		endDate = time.Now()
	}

	// Get metrics
	metrics, total, err := h.usageService.GetUsageMetrics(c.Request.Context(), projectID, startDate, endDate, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve usage metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics":    metrics,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetUsageSummary retrieves aggregated usage summary for a project
// GET /v1/usage/:project_id/summary
func (h *UsageHandler) GetUsageSummary(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get query parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Parse dates
	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (use YYYY-MM-DD)"})
			return
		}
	} else {
		// Default to current month
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (use YYYY-MM-DD)"})
			return
		}
	} else {
		endDate = time.Now()
	}

	// Get summary
	summary, err := h.usageService.GetUsageSummary(c.Request.Context(), projectID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve usage summary"})
		return
	}

	// Get project to get organization plan
	project, _ := c.Get("project")
	if project != nil {
		proj := project.(models.Project)
		// Get organization to get plan
		// For now, we'll use a placeholder
		_ = proj
	}

	c.JSON(http.StatusOK, summary)
}

// GetAggregatedUsage retrieves pre-aggregated usage data
// GET /v1/usage/:project_id/aggregated
func (h *UsageHandler) GetAggregatedUsage(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get query parameters
	periodType := c.DefaultQuery("period_type", "daily") // hourly, daily, monthly
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Validate period type
	if periodType != models.PeriodHourly && periodType != models.PeriodDaily && periodType != models.PeriodMonthly {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period_type (use: hourly, daily, monthly)"})
		return
	}

	// Parse dates
	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (use YYYY-MM-DD)"})
			return
		}
	} else {
		// Default to last 30 days
		startDate = time.Now().Add(-30 * 24 * time.Hour)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (use YYYY-MM-DD)"})
			return
		}
	} else {
		endDate = time.Now()
	}

	// Get aggregated data
	aggregates, err := h.aggregatorService.GetAggregatedUsage(c.Request.Context(), projectID, periodType, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve aggregated usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period_type": periodType,
		"start_date":  startDate,
		"end_date":    endDate,
		"aggregates":  aggregates,
		"count":       len(aggregates),
	})
}

// GetAlerts retrieves usage alerts for a project
// GET /v1/usage/:project_id/alerts
func (h *UsageHandler) GetAlerts(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get alerts
	alerts, err := h.usageService.GetAlerts(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// CheckLimits checks if usage is approaching limits
// POST /v1/usage/:project_id/check-limits
func (h *UsageHandler) CheckLimits(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	_, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get project to determine plan
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Project not found in context"})
		return
	}
	proj := project.(models.Project)

	// Get organization to get plan (placeholder - would need actual org fetch)
	plan := "Free" // Default to Free

	// Check current month
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endDate := now

	// Check limits
	alerts, err := h.usageService.CheckLimits(c.Request.Context(), proj.ID, plan, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check limits"})
		return
	}

	if len(alerts) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"approaching_limits": true,
			"alerts":            alerts,
			"count":             len(alerts),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"approaching_limits": false,
			"alerts":            []models.UsageAlert{},
			"count":             0,
		})
	}
}
