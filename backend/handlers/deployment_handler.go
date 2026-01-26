package handlers

import (
	"net/http"
	"strconv"

	"pulse-control-plane/database"
	"pulse-control-plane/models"
	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeploymentHandler handles deployment configuration endpoints
type DeploymentHandler struct {
	service *services.DeploymentService
}

// NewDeploymentHandler creates a new deployment handler
func NewDeploymentHandler() *DeploymentHandler {
	db := database.GetDB()
	return &DeploymentHandler{
		service: services.NewDeploymentService(db),
	}
}

// CreateDeploymentConfig creates a new deployment configuration
// @Summary Create deployment config
// @Description Create deployment configuration for an organization
// @Tags Deployment
// @Accept json
// @Produce json
// @Param config body models.DeploymentConfig true "Deployment Config"
// @Success 201 {object} models.DeploymentConfig
// @Router /api/v1/deployment/config [post]
func (h *DeploymentHandler) CreateDeploymentConfig(c *gin.Context) {
	var config models.DeploymentConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.service.CreateDeploymentConfig(c.Request.Context(), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, config)
}

// GetDeploymentConfig retrieves deployment config for an organization
// @Summary Get deployment config
// @Description Get deployment configuration for an organization
// @Tags Deployment
// @Produce json
// @Param org_id path string true "Organization ID"
// @Success 200 {object} models.DeploymentConfig
// @Router /api/v1/deployment/config/{org_id} [get]
func (h *DeploymentHandler) GetDeploymentConfig(c *gin.Context) {
	orgIDStr := c.Param("org_id")
	orgID, err := primitive.ObjectIDFromHex(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}
	
	config, err := h.service.GetDeploymentConfig(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, config)
}

// UpdateDeploymentConfig updates deployment configuration
// @Summary Update deployment config
// @Description Update deployment configuration
// @Tags Deployment
// @Accept json
// @Produce json
// @Param id path string true "Config ID"
// @Param updates body map[string]interface{} true "Updates"
// @Success 200 {object} map[string]string
// @Router /api/v1/deployment/config/{id} [put]
func (h *DeploymentHandler) UpdateDeploymentConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config ID"})
		return
	}
	
	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.service.UpdateDeploymentConfig(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "deployment configuration updated successfully"})
}

// DeleteDeploymentConfig deletes deployment configuration
// @Summary Delete deployment config
// @Description Delete deployment configuration
// @Tags Deployment
// @Param id path string true "Config ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/deployment/config/{id} [delete]
func (h *DeploymentHandler) DeleteDeploymentConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config ID"})
		return
	}
	
	if err := h.service.DeleteDeploymentConfig(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "deployment configuration deleted successfully"})
}

// ValidateLicense validates a license key
// @Summary Validate license
// @Description Validate a self-hosted license key
// @Tags Deployment
// @Accept json
// @Produce json
// @Param body body map[string]string true "License key"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/deployment/validate-license [post]
func (h *DeploymentHandler) ValidateLicense(c *gin.Context) {
	var body struct {
		LicenseKey string `json:"license_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	valid, err := h.service.ValidateLicense(c.Request.Context(), body.LicenseKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "valid": false})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"valid": valid, "message": "license is valid"})
}

// GetDeploymentMetrics retrieves deployment metrics
// @Summary Get deployment metrics
// @Description Get recent health and resource metrics for a deployment
// @Tags Deployment
// @Produce json
// @Param deployment_id path string true "Deployment ID"
// @Param limit query int false "Limit (default 100)"
// @Success 200 {array} models.DeploymentMetrics
// @Router /api/v1/deployment/metrics/{deployment_id} [get]
func (h *DeploymentHandler) GetDeploymentMetrics(c *gin.Context) {
	deploymentIDStr := c.Param("deployment_id")
	deploymentID, err := primitive.ObjectIDFromHex(deploymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deployment ID"})
		return
	}
	
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	metrics, err := h.service.GetDeploymentMetrics(c.Request.Context(), deploymentID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"metrics": metrics, "count": len(metrics)})
}

// GetLatestDeploymentMetrics retrieves latest metrics
// @Summary Get latest deployment metrics
// @Description Get the most recent health metrics for a deployment
// @Tags Deployment
// @Produce json
// @Param deployment_id path string true "Deployment ID"
// @Success 200 {object} models.DeploymentMetrics
// @Router /api/v1/deployment/metrics/{deployment_id}/latest [get]
func (h *DeploymentHandler) GetLatestDeploymentMetrics(c *gin.Context) {
	deploymentIDStr := c.Param("deployment_id")
	deploymentID, err := primitive.ObjectIDFromHex(deploymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deployment ID"})
		return
	}
	
	metrics, err := h.service.GetLatestDeploymentMetrics(c.Request.Context(), deploymentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, metrics)
}

// ListDeploymentConfigs lists all deployment configurations
// @Summary List deployment configs
// @Description List all deployment configurations with filters
// @Tags Deployment
// @Produce json
// @Param deployment_type query string false "Filter by deployment type"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 20)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/deployment/configs [get]
func (h *DeploymentHandler) ListDeploymentConfigs(c *gin.Context) {
	deploymentType := models.DeploymentType(c.Query("deployment_type"))
	
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	configs, totalCount, err := h.service.ListDeploymentConfigs(c.Request.Context(), deploymentType, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"configs":     configs,
		"total_count": totalCount,
		"page":        page,
		"limit":       limit,
	})
}
