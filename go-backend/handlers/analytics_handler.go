package handlers

import (
	"net/http"
	"strconv"
	"time"

	"pulse-control-plane/models"
	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
}

func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: services.NewAnalyticsService(),
	}
}

// CreateCustomMetric creates a new custom metric
// @Summary Create custom metric
// @Description Define a new custom metric for tracking
// @Tags analytics
// @Accept json
// @Produce json
// @Param metric body models.CustomMetric true "Custom Metric"
// @Success 201 {object} models.CustomMetric
// @Router /v1/analytics/metrics/custom [post]
func (h *AnalyticsHandler) CreateCustomMetric(c *gin.Context) {
	var metric models.CustomMetric
	if err := c.ShouldBindJSON(&metric); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.analyticsService.CreateCustomMetric(c.Request.Context(), &metric); err != nil {
		log.Error().Err(err).Msg("Failed to create custom metric")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create custom metric"})
		return
	}

	c.JSON(http.StatusCreated, metric)
}

// GetCustomMetrics returns all custom metrics for a project
// @Summary List custom metrics
// @Description Get all custom metrics for a project
// @Tags analytics
// @Param project_id path string true "Project ID"
// @Produce json
// @Success 200 {array} models.CustomMetric
// @Router /v1/analytics/metrics/custom/{project_id} [get]
func (h *AnalyticsHandler) GetCustomMetrics(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	metrics, err := h.analyticsService.GetCustomMetrics(c.Request.Context(), projectID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get custom metrics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get custom metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
		"count":   len(metrics),
	})
}

// CreateAlert creates a new metric alert
// @Summary Create alert
// @Description Create a new alert for a metric
// @Tags analytics
// @Accept json
// @Produce json
// @Param alert body models.MetricAlert true "Metric Alert"
// @Success 201 {object} models.MetricAlert
// @Router /v1/analytics/alerts [post]
func (h *AnalyticsHandler) CreateAlert(c *gin.Context) {
	var alert models.MetricAlert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.analyticsService.CreateAlert(c.Request.Context(), &alert); err != nil {
		log.Error().Err(err).Msg("Failed to create alert")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create alert"})
		return
	}

	c.JSON(http.StatusCreated, alert)
}

// GetAlerts returns all alerts for a project
// @Summary List alerts
// @Description Get all alerts for a project
// @Tags analytics
// @Param project_id path string true "Project ID"
// @Produce json
// @Success 200 {array} models.MetricAlert
// @Router /v1/analytics/alerts/{project_id} [get]
func (h *AnalyticsHandler) GetAlerts(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	alerts, err := h.analyticsService.GetAlerts(c.Request.Context(), projectID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get alerts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// CheckAlerts checks all active alerts and triggers if needed
// @Summary Check alerts
// @Description Check all alerts and trigger if thresholds are met
// @Tags analytics
// @Param project_id path string true "Project ID"
// @Produce json
// @Success 200 {array} models.AlertTrigger
// @Router /v1/analytics/alerts/{project_id}/check [post]
func (h *AnalyticsHandler) CheckAlerts(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	triggers, err := h.analyticsService.CheckAlerts(c.Request.Context(), projectID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check alerts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"triggers":       triggers,
		"triggered_count": len(triggers),
		"checked_at":     time.Now(),
	})
}

// GetRealTimeDashboard returns real-time analytics dashboard data
// @Summary Get real-time dashboard
// @Description Get real-time analytics data for dashboard
// @Tags analytics
// @Param project_id path string true "Project ID"
// @Produce json
// @Success 200 {object} models.AnalyticsDashboard
// @Router /v1/analytics/realtime/{project_id} [get]
func (h *AnalyticsHandler) GetRealTimeDashboard(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	dashboard, err := h.analyticsService.GetRealTimeDashboard(c.Request.Context(), projectID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get real-time dashboard")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get real-time dashboard"})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// ExportAnalytics exports analytics data
// @Summary Export analytics
// @Description Export analytics data in specified format
// @Tags analytics
// @Param project_id path string true "Project ID"
// @Param export_type query string true "Export Type" Enums(csv, json)
// @Param date_from query string true "Date From" Format(2006-01-02)
// @Param date_to query string true "Date To" Format(2006-01-02)
// @Param metrics query []string false "Metric Types"
// @Produce json
// @Success 202 {object} models.AnalyticsExport
// @Router /v1/analytics/export/{project_id} [post]
func (h *AnalyticsHandler) ExportAnalytics(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	exportType := c.Query("export_type")
	if exportType != "csv" && exportType != "json" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid export type. Must be 'csv' or 'json'"})
		return
	}

	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")

	dateFrom, err := time.Parse("2006-01-02", dateFromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_from format. Use YYYY-MM-DD"})
		return
	}

	dateTo, err := time.Parse("2006-01-02", dateToStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_to format. Use YYYY-MM-DD"})
		return
	}

	// Get metrics array from query (can be multiple)
	metrics := c.QueryArray("metrics")

	export, err := h.analyticsService.ExportAnalytics(c.Request.Context(), projectID, exportType, dateFrom, dateTo, metrics)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initiate export")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate export"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"export": export,
		"message": "Export initiated. Check back later for the file.",
	})
}

// GetExportStatus returns the status of an export
// @Summary Get export status
// @Description Get the status of an analytics export
// @Tags analytics
// @Param export_id path string true "Export ID"
// @Produce json
// @Success 200 {object} models.AnalyticsExport
// @Router /v1/analytics/export/status/{export_id} [get]
func (h *AnalyticsHandler) GetExportStatus(c *gin.Context) {
	exportID, err := primitive.ObjectIDFromHex(c.Param("export_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid export ID"})
		return
	}

	// Get export from database
	var export models.AnalyticsExport
	// This should be implemented in the service
	// For now, return a placeholder
	_ = exportID
	
	c.JSON(http.StatusOK, export)
}

// ForecastUsage generates usage forecast
// @Summary Forecast usage
// @Description Generate usage forecast for specified metric
// @Tags analytics
// @Param project_id path string true "Project ID"
// @Param metric_type query string true "Metric Type"
// @Param days query int false "Forecast Days" default(7)
// @Produce json
// @Success 200 {array} models.UsageForecast
// @Router /v1/analytics/forecast/{project_id} [get]
func (h *AnalyticsHandler) ForecastUsage(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	metricType := c.Query("metric_type")
	if metricType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metric_type is required"})
		return
	}

	days := 7 // default
	if daysStr := c.Query("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 && parsedDays <= 90 {
			days = parsedDays
		}
	}

	forecasts, err := h.analyticsService.ForecastUsage(c.Request.Context(), projectID, metricType, days)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate forecast")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"forecasts": forecasts,
		"count":     len(forecasts),
		"metric":    metricType,
		"days":      days,
	})
}

// GetRecentTriggers returns recent alert triggers
// @Summary Get recent triggers
// @Description Get recent alert triggers for a project
// @Tags analytics
// @Param project_id path string true "Project ID"
// @Param limit query int false "Limit" default(20)
// @Produce json
// @Success 200 {array} models.AlertTrigger
// @Router /v1/analytics/triggers/{project_id} [get]
func (h *AnalyticsHandler) GetRecentTriggers(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	triggers, err := h.analyticsService.GetRecentTriggers(c.Request.Context(), projectID, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get recent triggers")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent triggers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"triggers": triggers,
		"count":    len(triggers),
	})
}
