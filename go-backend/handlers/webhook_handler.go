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

// WebhookHandler handles webhook-related HTTP requests
type WebhookHandler struct {
	webhookService *services.WebhookService
	egressService  *services.EgressService
	ingressService *services.IngressService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{
		webhookService: services.NewWebhookService(),
		egressService:  services.NewEgressService(),
		ingressService: services.NewIngressService(),
	}
}

// HandleLiveKitWebhook handles POST /v1/webhooks/livekit
// This endpoint receives webhooks from LiveKit server
func (h *WebhookHandler) HandleLiveKitWebhook(c *gin.Context) {
	// Parse webhook payload
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Error().Err(err).Msg("Invalid webhook payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}
	
	// Extract event type
	eventType, ok := payload["event"].(string)
	if !ok {
		log.Error().Msg("Missing event type in webhook")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing event type"})
		return
	}
	
	log.Info().Str("event", eventType).Msg("Received LiveKit webhook")
	
	// Process webhook based on event type
	switch eventType {
	case "egress_started":
		h.handleEgressStarted(c, payload)
	case "egress_ended":
		h.handleEgressEnded(c, payload)
	case "ingress_started":
		h.handleIngressStarted(c, payload)
	case "ingress_ended":
		h.handleIngressEnded(c, payload)
	case "participant_joined":
		h.handleParticipantJoined(c, payload)
	case "participant_left":
		h.handleParticipantLeft(c, payload)
	case "room_started":
		h.handleRoomStarted(c, payload)
	case "room_ended":
		h.handleRoomEnded(c, payload)
	default:
		log.Warn().Str("event", eventType).Msg("Unknown webhook event type")
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Webhook received"})
}

// handleEgressStarted processes egress started event
func (h *WebhookHandler) handleEgressStarted(c *gin.Context, payload map[string]interface{}) {
	egressID, ok := payload["egress_id"].(string)
	if !ok {
		return
	}
	
	// Update egress status
	err := h.egressService.UpdateEgressStatus(c.Request.Context(), egressID, models.EgressStatusActive, "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to update egress status")
	}
}

// handleEgressEnded processes egress ended event
func (h *WebhookHandler) handleEgressEnded(c *gin.Context, payload map[string]interface{}) {
	egressID, ok := payload["egress_id"].(string)
	if !ok {
		return
	}
	
	// Update egress status
	errorMsg := ""
	if errInterface, exists := payload["error"]; exists {
		if errStr, ok := errInterface.(string); ok {
			errorMsg = errStr
		}
	}
	
	status := models.EgressStatusEnded
	if errorMsg != "" {
		status = models.EgressStatusFailed
	}
	
	err := h.egressService.UpdateEgressStatus(c.Request.Context(), egressID, status, errorMsg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update egress status")
	}
}

// handleIngressStarted processes ingress started event
func (h *WebhookHandler) handleIngressStarted(c *gin.Context, payload map[string]interface{}) {
	ingressID, ok := payload["ingress_id"].(string)
	if !ok {
		return
	}
	
	// Update ingress status
	err := h.ingressService.UpdateIngressStatus(c.Request.Context(), ingressID, models.IngressStatusActive, "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to update ingress status")
	}
}

// handleIngressEnded processes ingress ended event
func (h *WebhookHandler) handleIngressEnded(c *gin.Context, payload map[string]interface{}) {
	ingressID, ok := payload["ingress_id"].(string)
	if !ok {
		return
	}
	
	// Update ingress status
	errorMsg := ""
	if errInterface, exists := payload["error"]; exists {
		if errStr, ok := errInterface.(string); ok {
			errorMsg = errStr
		}
	}
	
	status := models.IngressStatusInactive
	if errorMsg != "" {
		status = models.IngressStatusError
	}
	
	err := h.ingressService.UpdateIngressStatus(c.Request.Context(), ingressID, status, errorMsg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update ingress status")
	}
}

// handleParticipantJoined processes participant joined event
func (h *WebhookHandler) handleParticipantJoined(c *gin.Context, payload map[string]interface{}) {
	// Extract project ID and room name
	projectIDStr, _ := payload["project_id"].(string)
	roomName, _ := payload["room_name"].(string)
	
	// Create webhook payload
	webhookPayload := &models.WebhookPayload{
		Event:     models.WebhookEventParticipantJoined,
		Timestamp: time.Now().Unix(),
		ProjectID: projectIDStr,
		RoomName:  roomName,
		Metadata:  payload,
	}
	
	// Forward to customer webhook
	// Note: You would need to get the project from DB first
	log.Info().Str("event", "participant_joined").Str("room", roomName).Msg("Participant joined")
}

// handleParticipantLeft processes participant left event
func (h *WebhookHandler) handleParticipantLeft(c *gin.Context, payload map[string]interface{}) {
	log.Info().Str("event", "participant_left").Msg("Participant left")
}

// handleRoomStarted processes room started event
func (h *WebhookHandler) handleRoomStarted(c *gin.Context, payload map[string]interface{}) {
	log.Info().Str("event", "room_started").Msg("Room started")
}

// handleRoomEnded processes room ended event
func (h *WebhookHandler) handleRoomEnded(c *gin.Context, payload map[string]interface{}) {
	log.Info().Str("event", "room_ended").Msg("Room ended")
}

// GetWebhookLogs handles GET /v1/webhooks/logs
func (h *WebhookHandler) GetWebhookLogs(c *gin.Context) {
	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	proj := project.(*models.Project)
	
	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	
	// Get webhook logs
	logs, total, err := h.webhookService.GetWebhookLogs(c.Request.Context(), proj.ID, page, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get webhook logs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
