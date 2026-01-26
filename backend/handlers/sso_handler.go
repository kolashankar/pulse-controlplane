package handlers

import (
	"net/http"

	"pulse-control-plane/database"
	"pulse-control-plane/models"
	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SSOHandler handles SSO endpoints
type SSOHandler struct {
	service *services.SSOService
}

// NewSSOHandler creates a new SSO handler
func NewSSOHandler() *SSOHandler {
	db := database.GetDB()
	return &SSOHandler{
		service: services.NewSSOService(db),
	}
}

// CreateSSOConfig creates a new SSO configuration
// @Summary Create SSO configuration
// @Description Create SSO configuration for an organization
// @Tags SSO
// @Accept json
// @Produce json
// @Param config body models.SSOConfig true "SSO Configuration"
// @Success 201 {object} models.SSOConfig
// @Router /api/v1/sso/config [post]
func (h *SSOHandler) CreateSSOConfig(c *gin.Context) {
	var config models.SSOConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.service.CreateSSOConfig(c.Request.Context(), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, config)
}

// GetSSOConfig retrieves SSO configuration for an organization
// @Summary Get SSO configuration
// @Description Get SSO configuration for an organization
// @Tags SSO
// @Produce json
// @Param org_id path string true "Organization ID"
// @Success 200 {object} models.SSOConfig
// @Router /api/v1/sso/config/{org_id} [get]
func (h *SSOHandler) GetSSOConfig(c *gin.Context) {
	orgIDStr := c.Param("org_id")
	orgID, err := primitive.ObjectIDFromHex(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}
	
	config, err := h.service.GetSSOConfig(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, config)
}

// UpdateSSOConfig updates SSO configuration
// @Summary Update SSO configuration
// @Description Update SSO configuration
// @Tags SSO
// @Accept json
// @Produce json
// @Param id path string true "Config ID"
// @Param updates body map[string]interface{} true "Updates"
// @Success 200 {object} map[string]string
// @Router /api/v1/sso/config/{id} [put]
func (h *SSOHandler) UpdateSSOConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config ID"})
		return
	}
	
	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.service.UpdateSSOConfig(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "SSO configuration updated successfully"})
}

// DeleteSSOConfig deletes SSO configuration
// @Summary Delete SSO configuration
// @Description Delete SSO configuration
// @Tags SSO
// @Param id path string true "Config ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/sso/config/{id} [delete]
func (h *SSOHandler) DeleteSSOConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config ID"})
		return
	}
	
	if err := h.service.DeleteSSOConfig(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "SSO configuration deleted successfully"})
}

// OAuthCallback handles OAuth provider callbacks
// @Summary OAuth callback
// @Description Handle OAuth provider callback
// @Tags SSO
// @Param provider path string true "Provider (google/microsoft/github)"
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 200 {object} models.SSOSession
// @Router /api/v1/sso/callback/{provider} [get]
func (h *SSOHandler) OAuthCallback(c *gin.Context) {
	provider := models.SSOProvider(c.Param("provider"))
	code := c.Query("code")
	state := c.Query("state") // state contains org_id
	
	orgID, err := primitive.ObjectIDFromHex(state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state parameter"})
		return
	}
	
	session, err := h.service.ValidateOAuthCallback(c.Request.Context(), orgID, provider, code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, session)
}

// SAMLCallback handles SAML assertions
// @Summary SAML callback
// @Description Handle SAML assertion
// @Tags SSO
// @Accept application/x-www-form-urlencoded
// @Param SAMLResponse formData string true "SAML Response"
// @Success 200 {object} models.SSOSession
// @Router /api/v1/sso/saml [post]
func (h *SSOHandler) SAMLCallback(c *gin.Context) {
	assertion := c.PostForm("SAMLResponse")
	orgIDStr := c.PostForm("RelayState") // RelayState contains org_id
	
	orgID, err := primitive.ObjectIDFromHex(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid relay state"})
		return
	}
	
	session, err := h.service.ValidateSAMLAssertion(c.Request.Context(), orgID, assertion)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, session)
}
