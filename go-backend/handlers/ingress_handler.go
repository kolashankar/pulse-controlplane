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

// IngressHandler handles ingress-related HTTP requests
type IngressHandler struct {
	ingressService *services.IngressService
}

// NewIngressHandler creates a new ingress handler
func NewIngressHandler() *IngressHandler {
	return &IngressHandler{
		ingressService: services.NewIngressService(),
	}
}

// CreateIngress handles POST /v1/media/ingress/create
func (h *IngressHandler) CreateIngress(c *gin.Context) {
	// Get project from context (set by auth middleware)
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	proj := project.(*models.Project)
	
	// Parse request
	var req models.IngressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid ingress request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validate ingress type
	if req.IngressType != models.IngressTypeRTMP && 
	   req.IngressType != models.IngressTypeWHIP && 
	   req.IngressType != models.IngressTypeURL {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ingress type"})
		return
	}
	
	// Create ingress
	ingress, err := h.ingressService.CreateIngress(c.Request.Context(), proj.ID, proj, &req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create ingress")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Convert to response
	response := h.ingressService.ToResponse(ingress)
	
	log.Info().Str("ingress_id", ingress.ID.Hex()).Str("room", req.RoomName).Msg("Ingress created")
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Ingress created successfully",
		"ingress": response,
		"note": "Save the RTMP stream key or WHIP URL - they won't be shown again for security reasons",
	})
}

// GetIngress handles GET /v1/media/ingress/:id
func (h *IngressHandler) GetIngress(c *gin.Context) {
	// Get ingress ID from URL
	ingressIDStr := c.Param("id")
	
	// Parse ingress ID
	ingressID, err := primitive.ObjectIDFromHex(ingressIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ingress ID"})
		return
	}
	
	// Get ingress
	ingress, err := h.ingressService.GetIngress(c.Request.Context(), ingressID)
	if err != nil {
		log.Error().Err(err).Str("ingress_id", ingressIDStr).Msg("Failed to get ingress")
		c.JSON(http.StatusNotFound, gin.H{"error": "Ingress not found"})
		return
	}
	
	// Convert to response
	response := h.ingressService.ToResponse(ingress)
	
	c.JSON(http.StatusOK, response)
}

// ListIngresses handles GET /v1/media/ingress
func (h *IngressHandler) ListIngresses(c *gin.Context) {
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
	
	// List ingresses
	ingresses, total, err := h.ingressService.ListIngresses(c.Request.Context(), proj.ID, page, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list ingresses")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Convert to responses
	responses := make([]*models.IngressResponse, len(ingresses))
	for i, ingress := range ingresses {
		responses[i] = h.ingressService.ToResponse(&ingress)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"ingresses": responses,
		"total": total,
		"page": page,
		"limit": limit,
	})
}

// DeleteIngress handles DELETE /v1/media/ingress/:id
func (h *IngressHandler) DeleteIngress(c *gin.Context) {
	// Get ingress ID from URL
	ingressIDStr := c.Param("id")
	
	// Parse ingress ID
	ingressID, err := primitive.ObjectIDFromHex(ingressIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ingress ID"})
		return
	}
	
	// Delete ingress
	err = h.ingressService.DeleteIngress(c.Request.Context(), ingressID)
	if err != nil {
		log.Error().Err(err).Str("ingress_id", ingressIDStr).Msg("Failed to delete ingress")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	log.Info().Str("ingress_id", ingressIDStr).Msg("Ingress deleted")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Ingress deleted successfully",
	})
}
