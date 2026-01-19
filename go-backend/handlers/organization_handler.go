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

type OrganizationHandler struct {
	service *services.OrganizationService
}

func NewOrganizationHandler() *OrganizationHandler {
	return &OrganizationHandler{
		service: services.NewOrganizationService(),
	}
}

// CreateOrganization creates a new organization
// @Summary Create organization
// @Description Create a new organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param organization body models.OrganizationCreate true "Organization data"
// @Success 201 {object} models.Organization
// @Router /v1/organizations [post]
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var input models.OrganizationCreate

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Error().Err(err).Msg("Invalid input for organization creation")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	org, err := h.service.CreateOrganization(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, org)
}

// GetOrganization retrieves an organization by ID
// @Summary Get organization
// @Description Get organization by ID
// @Tags organizations
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} models.Organization
// @Router /v1/organizations/{id} [get]
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	idParam := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid organization ID format",
		})
		return
	}

	org, err := h.service.GetOrganization(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, org)
}

// ListOrganizations retrieves all organizations with pagination
// @Summary List organizations
// @Description Get all organizations with pagination and search
// @Tags organizations
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} map[string]interface{}
// @Router /v1/organizations [get]
func (h *OrganizationHandler) ListOrganizations(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
	search := c.Query("search")

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	orgs, total, err := h.service.ListOrganizations(c.Request.Context(), page, limit, search)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list organizations")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve organizations",
		})
		return
	}

	// Calculate total pages
	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"data": orgs,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// UpdateOrganization updates an organization
// @Summary Update organization
// @Description Update organization details
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param organization body models.OrganizationUpdate true "Organization update data"
// @Success 200 {object} models.Organization
// @Router /v1/organizations/{id} [put]
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	idParam := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid organization ID format",
		})
		return
	}

	var input models.OrganizationUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	org, err := h.service.UpdateOrganization(c.Request.Context(), id, &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, org)
}

// DeleteOrganization soft deletes an organization
// @Summary Delete organization
// @Description Soft delete an organization
// @Tags organizations
// @Param id path string true "Organization ID"
// @Success 200 {object} map[string]interface{}
// @Router /v1/organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	idParam := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid organization ID format",
		})
		return
	}

	err = h.service.DeleteOrganization(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Organization deleted successfully",
	})
}
