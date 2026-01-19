package handlers

import (
	"net/http"

	"pulse-control-plane/models"
	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TeamHandler handles team management operations
type TeamHandler struct {
	teamService *services.TeamService
}

// NewTeamHandler creates a new team handler
func NewTeamHandler() *TeamHandler {
	return &TeamHandler{
		teamService: services.NewTeamService(),
	}
}

// ListTeamMembers lists all team members for an organization
// GET /v1/organizations/:id/members
func (h *TeamHandler) ListTeamMembers(c *gin.Context) {
	orgID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	page := 1
	limit := 20

	if p, ok := c.GetQuery("page"); ok {
		if _, err := fmt.Sscanf(p, "%d", &page); err != nil {
			page = 1
		}
	}

	if l, ok := c.GetQuery("limit"); ok {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			limit = 20
		}
	}

	members, total, err := h.teamService.ListTeamMembers(c.Request.Context(), orgID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to responses
	responses := make([]models.TeamMemberResponse, len(members))
	for i, member := range members {
		responses[i] = member.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"members": responses,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// InviteTeamMember invites a new team member
// POST /v1/organizations/:id/members
func (h *TeamHandler) InviteTeamMember(c *gin.Context) {
	orgID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var invite models.TeamMemberInvite
	if err := c.ShouldBindJSON(&invite); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get inviter user ID from context (placeholder - should be from auth)
	// For now, use a dummy ObjectID
	invitedBy := primitive.NewObjectID()

	invitation, err := h.teamService.InviteTeamMember(c.Request.Context(), orgID, invitedBy, invite)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Team member invited successfully",
		"invitation": invitation.ToResponse(),
		"invite_url": fmt.Sprintf("https://pulse.io/invite?token=%s", invitation.Token),
	})
}

// RemoveTeamMember removes a team member from the organization
// DELETE /v1/organizations/:id/members/:user_id
func (h *TeamHandler) RemoveTeamMember(c *gin.Context) {
	orgID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.teamService.RemoveTeamMember(c.Request.Context(), orgID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team member removed successfully",
	})
}

// UpdateTeamMemberRole updates a team member's role
// PUT /v1/organizations/:id/members/:user_id/role
func (h *TeamHandler) UpdateTeamMemberRole(c *gin.Context) {
	orgID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var update models.TeamMemberUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.teamService.UpdateTeamMemberRole(c.Request.Context(), orgID, userID, update.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team member role updated successfully",
		"member":  member.ToResponse(),
	})
}

// GetTeamMember gets a team member by ID
// GET /v1/organizations/:id/members/:user_id
func (h *TeamHandler) GetTeamMember(c *gin.Context) {
	orgID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	member, err := h.teamService.GetTeamMember(c.Request.Context(), orgID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, member.ToResponse())
}

// ListPendingInvitations lists all pending invitations
// GET /v1/organizations/:id/invitations
func (h *TeamHandler) ListPendingInvitations(c *gin.Context) {
	orgID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	invitations, err := h.teamService.ListPendingInvitations(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to responses
	responses := make([]models.InvitationResponse, len(invitations))
	for i, inv := range invitations {
		responses[i] = inv.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": responses,
		"total":       len(responses),
	})
}

// RevokeInvitation revokes a pending invitation
// DELETE /v1/organizations/:id/invitations/:invitation_id
func (h *TeamHandler) RevokeInvitation(c *gin.Context) {
	orgID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	invitationID, err := primitive.ObjectIDFromHex(c.Param("invitation_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitation ID"})
		return
	}

	err = h.teamService.RevokeInvitation(c.Request.Context(), orgID, invitationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation revoked successfully",
	})
}

// AcceptInvitation accepts an invitation
// POST /v1/invitations/accept
func (h *TeamHandler) AcceptInvitation(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	member, err := h.teamService.AcceptInvitation(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation accepted successfully",
		"member":  member.ToResponse(),
	})
}
