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

// SupportHandler handles support ticket endpoints
type SupportHandler struct {
	service *services.SupportService
}

// NewSupportHandler creates a new support handler
func NewSupportHandler() *SupportHandler {
	db := database.GetDB()
	return &SupportHandler{
		service: services.NewSupportService(db),
	}
}

// CreateTicket creates a new support ticket
// @Summary Create support ticket
// @Description Create a new support ticket
// @Tags Support
// @Accept json
// @Produce json
// @Param ticket body models.SupportTicket true "Support Ticket"
// @Success 201 {object} models.SupportTicket
// @Router /api/v1/support/tickets [post]
func (h *SupportHandler) CreateTicket(c *gin.Context) {
	var ticket models.SupportTicket
	if err := c.ShouldBindJSON(&ticket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.service.CreateTicket(c.Request.Context(), &ticket); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, ticket)
}

// GetTicket retrieves a ticket by ID
// @Summary Get support ticket
// @Description Get a support ticket by ID
// @Tags Support
// @Produce json
// @Param id path string true "Ticket ID"
// @Success 200 {object} models.SupportTicket
// @Router /api/v1/support/tickets/{id} [get]
func (h *SupportHandler) GetTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket ID"})
		return
	}
	
	ticket, err := h.service.GetTicket(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, ticket)
}

// ListTickets lists tickets with filters
// @Summary List support tickets
// @Description List tickets with optional filters
// @Tags Support
// @Produce json
// @Param org_id query string false "Organization ID"
// @Param status query string false "Status filter"
// @Param priority query string false "Priority filter"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 20)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/support/tickets [get]
func (h *SupportHandler) ListTickets(c *gin.Context) {
	var orgID *primitive.ObjectID
	if orgIDStr := c.Query("org_id"); orgIDStr != "" {
		id, err := primitive.ObjectIDFromHex(orgIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
			return
		}
		orgID = &id
	}
	
	status := models.TicketStatus(c.Query("status"))
	priority := models.TicketPriority(c.Query("priority"))
	
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
	
	tickets, totalCount, err := h.service.ListTickets(c.Request.Context(), orgID, status, priority, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"tickets":     tickets,
		"total_count": totalCount,
		"page":        page,
		"limit":       limit,
	})
}

// UpdateTicket updates ticket fields
// @Summary Update support ticket
// @Description Update ticket status and other fields
// @Tags Support
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param updates body map[string]interface{} true "Update fields"
// @Success 200 {object} map[string]string
// @Router /api/v1/support/tickets/{id} [put]
func (h *SupportHandler) UpdateTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket ID"})
		return
	}
	
	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.service.UpdateTicket(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "ticket updated successfully"})
}

// AssignTicket assigns a ticket to an agent
// @Summary Assign ticket
// @Description Assign a ticket to a support agent
// @Tags Support
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param body body map[string]string true "Assignment (agent_id)"
// @Success 200 {object} map[string]string
// @Router /api/v1/support/tickets/{id}/assign [post]
func (h *SupportHandler) AssignTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket ID"})
		return
	}
	
	var body struct {
		AgentID string `json:"agent_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	agentID, err := primitive.ObjectIDFromHex(body.AgentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agent ID"})
		return
	}
	
	if err := h.service.AssignTicket(c.Request.Context(), id, agentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "ticket assigned successfully"})
}

// AddComment adds a comment to a ticket
// @Summary Add comment
// @Description Add a comment to a support ticket
// @Tags Support
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param comment body models.TicketComment true "Comment"
// @Success 201 {object} models.TicketComment
// @Router /api/v1/support/tickets/{id}/comments [post]
func (h *SupportHandler) AddComment(c *gin.Context) {
	idStr := c.Param("id")
	ticketID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket ID"})
		return
	}
	
	var comment models.TicketComment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	comment.TicketID = ticketID
	
	if err := h.service.AddComment(c.Request.Context(), &comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, comment)
}

// GetTicketComments retrieves all comments for a ticket
// @Summary Get ticket comments
// @Description Get all comments for a support ticket
// @Tags Support
// @Produce json
// @Param id path string true "Ticket ID"
// @Success 200 {array} models.TicketComment
// @Router /api/v1/support/tickets/{id}/comments [get]
func (h *SupportHandler) GetTicketComments(c *gin.Context) {
	idStr := c.Param("id")
	ticketID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket ID"})
		return
	}
	
	comments, err := h.service.GetTicketComments(c.Request.Context(), ticketID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"comments": comments, "count": len(comments)})
}

// GetTicketStats returns support ticket statistics
// @Summary Get ticket statistics
// @Description Get aggregate statistics for support tickets
// @Tags Support
// @Produce json
// @Param org_id query string false "Organization ID"
// @Success 200 {object} models.TicketStats
// @Router /api/v1/support/stats [get]
func (h *SupportHandler) GetTicketStats(c *gin.Context) {
	var orgID *primitive.ObjectID
	if orgIDStr := c.Query("org_id"); orgIDStr != "" {
		id, err := primitive.ObjectIDFromHex(orgIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
			return
		}
		orgID = &id
	}
	
	stats, err := h.service.GetTicketStats(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, stats)
}
