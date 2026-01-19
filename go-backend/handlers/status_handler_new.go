package handlers

import (
	"net/http"

	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StatusHandler handles system status and monitoring
type StatusHandler struct {
	statusService *services.StatusService
}

// NewStatusHandler creates a new status handler
func NewStatusHandler() *StatusHandler {
	return &StatusHandler{
		statusService: services.NewStatusService(),
	}
}

// GetSystemStatus returns the overall system status
// GET /v1/status
func (h *StatusHandler) GetSystemStatus(c *gin.Context) {
	status, err := h.statusService.GetSystemStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetProjectHealth returns the health status of a specific project
// GET /v1/status/projects/:id
func (h *StatusHandler) GetProjectHealth(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	health, err := h.statusService.GetProjectHealth(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, health)
}

// GetRegionAvailability returns the availability of all regions
// GET /v1/status/regions
func (h *StatusHandler) GetRegionAvailability(c *gin.Context) {
	regions, err := h.statusService.GetRegionAvailability(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"regions": regions,
		"total":   len(regions),
	})
}
