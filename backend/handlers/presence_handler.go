package handlers

import (
	"net/http"

	"pulse-control-plane/models"
	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
)

// PresenceHandler handles presence-related requests
type PresenceHandler struct {
	presenceService *services.PresenceService
}

// NewPresenceHandler creates a new presence handler
func NewPresenceHandler(presenceService *services.PresenceService) *PresenceHandler {
	return &PresenceHandler{presenceService: presenceService}
}

// SetOnline marks a user as online
func (h *PresenceHandler) SetOnline(c *gin.Context) {
	var req models.PresenceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	presence := &models.UserPresence{
		ProjectID:     projectModel.ID,
		UserID:        req.UserID,
		Status:        req.Status,
		StatusMessage: req.StatusMessage,
		CurrentRoom:   req.CurrentRoom,
		Device:        req.Device,
		IPAddress:     c.ClientIP(),
		UserAgent:     c.Request.UserAgent(),
	}

	if err := h.presenceService.SetOnline(c.Request.Context(), presence); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "User marked as online",
		"presence": presence,
	})
}

// SetOffline marks a user as offline
func (h *PresenceHandler) SetOffline(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	if err := h.presenceService.SetOffline(c.Request.Context(), projectModel.ID, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User marked as offline",
	})
}

// SetStatus updates user status (away, busy, etc.)
func (h *PresenceHandler) SetStatus(c *gin.Context) {
	var req struct {
		UserID        string                 `json:"user_id" binding:"required"`
		Status        models.PresenceStatus  `json:"status" binding:"required"`
		StatusMessage string                 `json:"status_message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	if err := h.presenceService.SetStatus(c.Request.Context(), projectModel.ID, req.UserID, req.Status, req.StatusMessage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Status updated successfully",
	})
}

// GetUserStatus retrieves a user's presence status
func (h *PresenceHandler) GetUserStatus(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	presence, err := h.presenceService.GetUserStatus(c.Request.Context(), projectModel.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, presence)
}

// GetBulkStatus retrieves presence status for multiple users
func (h *PresenceHandler) GetBulkStatus(c *gin.Context) {
	var req models.BulkPresenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	presences, err := h.presenceService.GetBulkStatus(c.Request.Context(), projectModel.ID, req.UserIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.BulkPresenceResponse{
		Presences: presences,
	})
}

// SetTyping sets a typing indicator
func (h *PresenceHandler) SetTyping(c *gin.Context) {
	var req models.TypingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	typing := &models.TypingIndicator{
		ProjectID: projectModel.ID,
		RoomID:    req.RoomID,
		UserID:    req.UserID,
		IsTyping:  req.IsTyping,
	}

	if err := h.presenceService.SetTyping(c.Request.Context(), typing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Typing indicator updated",
	})
}

// GetRoomPresence retrieves presence information for a room
func (h *PresenceHandler) GetRoomPresence(c *gin.Context) {
	roomID := c.Param("room_id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id is required"})
		return
	}

	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	roomPresence, err := h.presenceService.GetRoomPresence(c.Request.Context(), projectModel.ID, roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, roomPresence)
}

// UpdateActivity updates user activity
func (h *PresenceHandler) UpdateActivity(c *gin.Context) {
	var req models.ActivityUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	activity := &models.UserActivity{
		ProjectID:    projectModel.ID,
		UserID:       req.UserID,
		ActivityType: req.ActivityType,
		ResourceID:   req.ResourceID,
		ResourceType: req.ResourceType,
		Metadata:     req.Metadata,
	}

	if err := h.presenceService.UpdateActivity(c.Request.Context(), activity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Activity updated",
		"activity": activity,
	})
}

// GetUserActivities retrieves recent activities for a user
func (h *PresenceHandler) GetUserActivities(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	activities, err := h.presenceService.GetUserActivities(c.Request.Context(), projectModel.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
		"count":      len(activities),
	})
}
