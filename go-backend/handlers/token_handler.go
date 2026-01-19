package handlers

import (
	"net/http"

	"pulse-control-plane/config"
	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TokenHandler struct {
	service *services.TokenService
}

func NewTokenHandler(cfg *config.Config) *TokenHandler {
	return &TokenHandler{
		service: services.NewTokenService(cfg),
	}
}

// CreateToken exchanges Pulse API key for a LiveKit token
// @Summary Create LiveKit token
// @Description Exchange Pulse API Key for LiveKit JWT token
// @Tags tokens
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body services.TokenRequest true "Token request"
// @Success 200 {object} services.TokenResponse
// @Router /v1/tokens/create [post]
func (h *TokenHandler) CreateToken(c *gin.Context) {
	// Get project from context (set by AuthenticateProject middleware)
	projectIDStr, exists := c.Get("project_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Project authentication required",
		})
		return
	}

	projectID, err := primitive.ObjectIDFromHex(projectIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid project ID",
		})
		return
	}

	// Parse request body
	var req services.TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid token request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Set default permissions if not provided
	if !req.CanPublish && !req.CanSubscribe {
		req.CanPublish = true
		req.CanSubscribe = true
	}

	// Create token
	tokenResp, err := h.service.CreateToken(c.Request.Context(), projectID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tokenResp)
}

// ValidateToken validates an existing token
// @Summary Validate token
// @Description Validate an existing LiveKit JWT token
// @Tags tokens
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body map[string]string true "Token validation request" example({"token":"eyJhbGc..."})
// @Success 200 {object} map[string]interface{}
// @Router /v1/tokens/validate [post]
func (h *TokenHandler) ValidateToken(c *gin.Context) {
	// Parse request body
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: token is required",
		})
		return
	}

	// Validate token
	valid, info, err := h.service.ValidateToken(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid": false,
			"error": "Invalid or expired token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"info":  info,
	})
}
