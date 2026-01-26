package handlers

import (
	"net/http"
	"strconv"

	"pulse-control-plane/models"
	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FeedHandler handles feed-related requests
type FeedHandler struct {
	feedService *services.FeedService
}

// NewFeedHandler creates a new feed handler
func NewFeedHandler(feedService *services.FeedService) *FeedHandler {
	return &FeedHandler{feedService: feedService}
}

// CreateActivity creates a new activity and fans it out
func (h *FeedHandler) CreateActivity(c *gin.Context) {
	var req models.ActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get project from context (set by auth middleware)
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	activity := &models.Activity{
		ProjectID: projectModel.ID,
		Actor:     req.Actor,
		Verb:      req.Verb,
		Object:    req.Object,
		Target:    req.Target,
		ForeignID: req.ForeignID,
		Metadata:  req.Metadata,
	}

	if err := h.feedService.CreateActivity(c.Request.Context(), activity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Activity created successfully",
		"activity": activity,
	})
}

// GetFeedItems retrieves feed items for a user
func (h *FeedHandler) GetFeedItems(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))

	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	response, err := h.feedService.GetFeedItems(c.Request.Context(), projectModel.ID, userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetAggregatedFeed retrieves aggregated feed (grouped activities)
func (h *FeedHandler) GetAggregatedFeed(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))

	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	aggregated, err := h.feedService.GetAggregatedFeed(c.Request.Context(), projectModel.ID, userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": aggregated,
		"count":      len(aggregated),
	})
}

// Follow creates a follow relationship
func (h *FeedHandler) Follow(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var req struct {
		Follower string `json:"follower" binding:"required"`
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

	follow := &models.Follow{
		ProjectID: projectModel.ID,
		Follower:  req.Follower,
		Following: userID,
	}

	if err := h.feedService.Follow(c.Request.Context(), follow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Successfully followed user",
		"follow":  follow,
	})
}

// Unfollow removes a follow relationship
func (h *FeedHandler) Unfollow(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	follower := c.Query("follower")
	if follower == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "follower query param is required"})
		return
	}

	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	if err := h.feedService.Unfollow(c.Request.Context(), projectModel.ID, follower, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully unfollowed user",
	})
}

// GetFollowers retrieves followers of a user
func (h *FeedHandler) GetFollowers(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	followers, err := h.feedService.GetFollowers(c.Request.Context(), projectModel.ID, userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"followers": followers,
		"count":     len(followers),
	})
}

// GetFollowing retrieves users that a user is following
func (h *FeedHandler) GetFollowing(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	following, err := h.feedService.GetFollowing(c.Request.Context(), projectModel.ID, userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"following": following,
		"count":     len(following),
	})
}

// GetFollowStats retrieves follower/following statistics
func (h *FeedHandler) GetFollowStats(c *gin.Context) {
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

	stats, err := h.feedService.GetFollowStats(c.Request.Context(), projectModel.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// MarkAsSeen marks feed items as seen
func (h *FeedHandler) MarkAsSeen(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var req struct {
		ItemIDs []string `json:"item_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert string IDs to ObjectIDs
	itemIDs := make([]primitive.ObjectID, 0, len(req.ItemIDs))
	for _, idStr := range req.ItemIDs {
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			continue
		}
		itemIDs = append(itemIDs, id)
	}

	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	if err := h.feedService.MarkAsSeen(c.Request.Context(), projectModel.ID, userID, itemIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Items marked as seen"})
}

// MarkAsRead marks feed items as read
func (h *FeedHandler) MarkAsRead(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var req struct {
		ItemIDs []string `json:"item_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert string IDs to ObjectIDs
	itemIDs := make([]primitive.ObjectID, 0, len(req.ItemIDs))
	for _, idStr := range req.ItemIDs {
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			continue
		}
		itemIDs = append(itemIDs, id)
	}

	// Get project from context
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}
	projectModel := project.(*models.Project)

	if err := h.feedService.MarkAsRead(c.Request.Context(), projectModel.ID, userID, itemIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Items marked as read"})
}

// DeleteActivity deletes an activity
func (h *FeedHandler) DeleteActivity(c *gin.Context) {
	activityIDStr := c.Param("activity_id")
	activityID, err := primitive.ObjectIDFromHex(activityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid activity ID"})
		return
	}

	if err := h.feedService.DeleteActivity(c.Request.Context(), activityID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Activity deleted successfully"})
}
