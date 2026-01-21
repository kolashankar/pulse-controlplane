package handlers

import (
	"net/http"
	"strconv"

	"pulse-control-plane/models"
	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ModerationHandler handles moderation-related requests
type ModerationHandler struct {
	moderationService *services.ModerationService
}

// NewModerationHandler creates a new moderation handler
func NewModerationHandler(moderationService *services.ModerationService) *ModerationHandler {
	return &ModerationHandler{moderationService: moderationService}
}

// AnalyzeText analyzes text content for moderation
func (h *ModerationHandler) AnalyzeText(c *gin.Context) {
	var req models.AnalyzeTextRequest
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

	analysis, err := h.moderationService.AnalyzeText(
		c.Request.Context(),
		projectModel.ID,
		req.Content,
		req.ContentID,
		req.UserID,
		req.Metadata,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log the action if content is flagged
	if analysis.IsFlagged {
		log := &models.ModerationLog{
			ProjectID:   projectModel.ID,
			AnalysisID:  analysis.ID,
			ContentType: models.ContentTypeText,
			ContentID:   req.ContentID,
			UserID:      req.UserID,
			Action:      analysis.RecommendedAction,
			Severity:    analysis.Severity,
			Reason:      analysis.Reason,
			Automatic:   true,
		}
		_ = h.moderationService.LogAction(c.Request.Context(), log)
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis": analysis,
		"message":  "Content analyzed successfully",
	})
}

// AnalyzeImage analyzes image content for moderation
func (h *ModerationHandler) AnalyzeImage(c *gin.Context) {
	var req models.AnalyzeImageRequest
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

	analysis, err := h.moderationService.AnalyzeImage(
		c.Request.Context(),
		projectModel.ID,
		req.ImageURL,
		req.ContentID,
		req.UserID,
		req.Metadata,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log the action if content is flagged
	if analysis.IsFlagged {
		log := &models.ModerationLog{
			ProjectID:   projectModel.ID,
			AnalysisID:  analysis.ID,
			ContentType: models.ContentTypeImage,
			ContentID:   req.ContentID,
			UserID:      req.UserID,
			Action:      analysis.RecommendedAction,
			Severity:    analysis.Severity,
			Reason:      analysis.Reason,
			Automatic:   true,
		}
		_ = h.moderationService.LogAction(c.Request.Context(), log)
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis": analysis,
		"message":  "Image analyzed successfully",
	})
}

// CreateRule creates a new moderation rule
func (h *ModerationHandler) CreateRule(c *gin.Context) {
	var req models.CreateRuleRequest
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

	rule, err := h.moderationService.CreateRule(c.Request.Context(), projectModel.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Moderation rule created successfully",
		"rule":    rule,
	})
}

// GetRules retrieves moderation rules for a project
func (h *ModerationHandler) GetRules(c *gin.Context) {
	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	rules, err := h.moderationService.GetRules(c.Request.Context(), projectModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"count": len(rules),
	})
}

// GetLogs retrieves moderation logs for a project
func (h *ModerationHandler) GetLogs(c *gin.Context) {
	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	logs, total, err := h.moderationService.GetLogs(c.Request.Context(), projectModel.ID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":     logs,
		"total":    total,
		"page":     page,
		"limit":    limit,
		"has_more": total > int64(page*limit),
	})
}

// AddToWhitelist adds an entry to the whitelist
func (h *ModerationHandler) AddToWhitelist(c *gin.Context) {
	var req models.WhitelistRequest
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

	whitelist, err := h.moderationService.AddToWhitelist(c.Request.Context(), projectModel.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Added to whitelist successfully",
		"whitelist": whitelist,
	})
}

// AddToBlacklist adds an entry to the blacklist
func (h *ModerationHandler) AddToBlacklist(c *gin.Context) {
	var req models.BlacklistRequest
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

	blacklist, err := h.moderationService.AddToBlacklist(c.Request.Context(), projectModel.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Added to blacklist successfully",
		"blacklist": blacklist,
	})
}

// GetStats retrieves moderation statistics
func (h *ModerationHandler) GetStats(c *gin.Context) {
	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	period := c.DefaultQuery("period", "weekly")

	stats, err := h.moderationService.GetStats(c.Request.Context(), projectModel.ID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetConfig retrieves moderation configuration
func (h *ModerationHandler) GetConfig(c *gin.Context) {
	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	config, err := h.moderationService.GetOrCreateConfig(c.Request.Context(), projectModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}
