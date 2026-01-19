package handlers

import (
	"net/http"
	"time"

	"pulse-control-plane/models"
	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type RegionHandler struct {
	regionService *services.RegionService
}

func NewRegionHandler() *RegionHandler {
	return &RegionHandler{
		regionService: services.NewRegionService(),
	}
}

// GetAllRegions returns all region configurations
// @Summary List all regions
// @Description Get all LiveKit server regions with their configuration
// @Tags regions
// @Produce json
// @Success 200 {array} models.RegionConfig
// @Router /v1/regions [get]
func (h *RegionHandler) GetAllRegions(c *gin.Context) {
	regions, err := h.regionService.GetAllRegions(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get regions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get regions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"regions": regions,
		"count":   len(regions),
	})
}

// GetRegionByCode returns a specific region by code
// @Summary Get region by code
// @Description Get region details by region code
// @Tags regions
// @Param code path string true "Region Code"
// @Produce json
// @Success 200 {object} models.RegionConfig
// @Router /v1/regions/{code} [get]
func (h *RegionHandler) GetRegionByCode(c *gin.Context) {
	code := c.Param("code")

	region, err := h.regionService.GetRegionByCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Region not found"})
		return
	}

	c.JSON(http.StatusOK, region)
}

// GetRegionHealth returns health status for a specific region
// @Summary Check region health
// @Description Get health status for a specific region
// @Tags regions
// @Param code path string true "Region Code"
// @Produce json
// @Success 200 {object} models.RegionHealth
// @Router /v1/regions/{code}/health [get]
func (h *RegionHandler) GetRegionHealth(c *gin.Context) {
	code := c.Param("code")

	health, err := h.regionService.CheckRegionHealth(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check region health"})
		return
	}

	c.JSON(http.StatusOK, health)
}

// GetAllRegionHealth returns health status for all regions
// @Summary Get all region health
// @Description Get health status for all regions
// @Tags regions
// @Produce json
// @Success 200 {array} models.RegionHealth
// @Router /v1/regions/health [get]
func (h *RegionHandler) GetAllRegionHealth(c *gin.Context) {
	healthStatuses, err := h.regionService.GetAllRegionHealth(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get region health")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get region health"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"regions": healthStatuses,
		"count":   len(healthStatuses),
		"timestamp": time.Now(),
	})
}

// FindNearestRegion finds the best region for a client
// @Summary Find nearest region
// @Description Find the best region based on client location and latency
// @Tags regions
// @Accept json
// @Produce json
// @Param request body models.NearestRegionRequest true "Nearest Region Request"
// @Success 200 {object} models.NearestRegionResponse
// @Router /v1/regions/nearest [post]
func (h *RegionHandler) FindNearestRegion(c *gin.Context) {
	var req models.NearestRegionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// If client IP not provided, use request IP
	if req.ClientIP == "" {
		req.ClientIP = c.ClientIP()
	}

	response, err := h.regionService.FindNearestRegion(c.Request.Context(), &req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to find nearest region")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find nearest region"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetRegionStats returns aggregated region statistics
// @Summary Get region statistics
// @Description Get aggregated statistics for all regions
// @Tags regions
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /v1/regions/stats [get]
func (h *RegionHandler) GetRegionStats(c *gin.Context) {
	stats, err := h.regionService.GetRegionStats(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get region stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get region stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
