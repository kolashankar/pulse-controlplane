package handlers

import (
	"net/http"
	"strconv"

	"pulse-control-plane/models"
	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EgressHandler handles egress-related HTTP requests
type EgressHandler struct {
	egressService *services.EgressService
}

// NewEgressHandler creates a new egress handler
func NewEgressHandler() *EgressHandler {
	return &EgressHandler{
		egressService: services.NewEgressService(),
	}
}

// StartEgress handles POST /v1/media/egress/start
func (h *EgressHandler) StartEgress(c *gin.Context) {
	// Get project from context (set by auth middleware)
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	proj := project.(*models.Project)
	
	// Parse request
	var req models.EgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid egress request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validate egress type
	if req.EgressType != models.EgressTypeRoomComposite && 
	   req.EgressType != models.EgressTypeTrackComposite && 
	   req.EgressType != models.EgressTypeTrack {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid egress type"})
		return
	}
	
	// Validate output type
	if req.OutputType != models.OutputTypeHLS && 
	   req.OutputType != models.OutputTypeRTMP && 
	   req.OutputType != models.OutputTypeFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid output type"})
		return
	}
	
	// Start egress
	egress, err := h.egressService.StartEgress(c.Request.Context(), proj.ID, proj, &req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start egress")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Convert to response
	response := h.egressService.ToResponse(egress)
	
	log.Info().Str("egress_id", egress.ID.Hex()).Str("room", req.RoomName).Msg("Egress started")
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Egress started successfully",
		"egress": response,
	})
}

// StopEgress handles POST /v1/media/egress/stop
func (h *EgressHandler) StopEgress(c *gin.Context) {
	// Get egress ID from request body
	var req struct {
		EgressID string `json:"egress_id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Parse egress ID
	egressID, err := primitive.ObjectIDFromHex(req.EgressID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid egress ID"})
		return
	}
	
	// Stop egress
	egress, err := h.egressService.StopEgress(c.Request.Context(), egressID)
	if err != nil {
		log.Error().Err(err).Str("egress_id", req.EgressID).Msg("Failed to stop egress")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Convert to response
	response := h.egressService.ToResponse(egress)
	
	log.Info().Str("egress_id", egress.ID.Hex()).Msg("Egress stopped")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Egress stopped successfully",
		"egress": response,
	})
}

// GetEgress handles GET /v1/media/egress/:id
func (h *EgressHandler) GetEgress(c *gin.Context) {
	// Get egress ID from URL
	egressIDStr := c.Param("id")
	
	// Parse egress ID
	egressID, err := primitive.ObjectIDFromHex(egressIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid egress ID"})
		return
	}
	
	// Get egress
	egress, err := h.egressService.GetEgress(c.Request.Context(), egressID)
	if err != nil {
		log.Error().Err(err).Str("egress_id", egressIDStr).Msg("Failed to get egress")
		c.JSON(http.StatusNotFound, gin.H{"error": "Egress not found"})
		return
	}
	
	// Convert to response
	response := h.egressService.ToResponse(egress)
	
	c.JSON(http.StatusOK, response)
}

// ListEgresses handles GET /v1/media/egress
func (h *EgressHandler) ListEgresses(c *gin.Context) {
	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	proj := project.(*models.Project)
	
	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	// List egresses
	egresses, total, err := h.egressService.ListEgresses(c.Request.Context(), proj.ID, page, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list egresses")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Convert to responses
	responses := make([]*models.EgressResponse, len(egresses))
	for i, egress := range egresses {
		responses[i] = h.egressService.ToResponse(&egress)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"egresses": responses,
		"total": total,
		"page": page,
		"limit": limit,
	})
}
