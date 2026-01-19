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

type ProjectHandler struct {
	service *services.ProjectService
}

func NewProjectHandler() *ProjectHandler {
	return &ProjectHandler{
		service: services.NewProjectService(),
	}
}

// CreateProject creates a new project
// @Summary Create project
// @Description Create a new project with API keys
// @Tags projects
// @Accept json
// @Produce json
// @Param project body models.ProjectCreate true "Project data"
// @Success 201 {object} map[string]interface{}
// @Router /v1/projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var input models.ProjectCreate

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Error().Err(err).Msg("Invalid input for project creation")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	// Get org_id from query parameter or request body
	orgIDParam := c.Query("org_id")
	if orgIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "org_id is required",
		})
		return
	}

	orgID, err := primitive.ObjectIDFromHex(orgIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid organization ID format",
		})
		return
	}

	project, apiSecret, err := h.service.CreateProject(c.Request.Context(), orgID, &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Return project with API secret (only shown once)
	c.JSON(http.StatusCreated, gin.H{
		"project":    project.ToResponse(),
		"api_secret": apiSecret,
		"message":    "⚠️ IMPORTANT: Save your API secret now. It won't be shown again.",
	})
}

// GetProject retrieves a project by ID
// @Summary Get project
// @Description Get project by ID
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} models.ProjectResponse
// @Router /v1/projects/{id} [get]
func (h *ProjectHandler) GetProject(c *gin.Context) {
	idParam := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID format",
		})
		return
	}

	project, err := h.service.GetProject(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, project.ToResponse())
}

// ListProjects retrieves all projects with pagination
// @Summary List projects
// @Description Get all projects with pagination and search
// @Tags projects
// @Produce json
// @Param org_id query string false "Organization ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} map[string]interface{}
// @Router /v1/projects [get]
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
	search := c.Query("search")

	// Parse optional org_id filter
	var orgID *primitive.ObjectID
	orgIDParam := c.Query("org_id")
	if orgIDParam != "" {
		parsedOrgID, err := primitive.ObjectIDFromHex(orgIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid organization ID format",
			})
			return
		}
		orgID = &parsedOrgID
	}

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	projects, total, err := h.service.ListProjects(c.Request.Context(), orgID, page, limit, search)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list projects")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve projects",
		})
		return
	}

	// Convert to response format
	projectResponses := make([]*models.ProjectResponse, len(projects))
	for i, p := range projects {
		projectResponses[i] = p.ToResponse()
	}

	// Calculate total pages
	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"data": projectResponses,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// UpdateProject updates a project
// @Summary Update project
// @Description Update project details
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param project body models.ProjectUpdate true "Project update data"
// @Success 200 {object} models.ProjectResponse
// @Router /v1/projects/{id} [put]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	idParam := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID format",
		})
		return
	}

	var input models.ProjectUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	project, err := h.service.UpdateProject(c.Request.Context(), id, &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, project.ToResponse())
}

// DeleteProject soft deletes a project
// @Summary Delete project
// @Description Soft delete a project
// @Tags projects
// @Param id path string true "Project ID"
// @Success 200 {object} map[string]interface{}
// @Router /v1/projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	idParam := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID format",
		})
		return
	}

	err = h.service.DeleteProject(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project deleted successfully",
	})
}

// RegenerateAPIKeys regenerates API keys for a project
// @Summary Regenerate API keys
// @Description Generate new API key and secret for a project
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} map[string]interface{}
// @Router /v1/projects/{id}/regenerate-keys [post]
func (h *ProjectHandler) RegenerateAPIKeys(c *gin.Context) {
	idParam := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID format",
		})
		return
	}

	apiKey, apiSecret, err := h.service.RegenerateAPIKeys(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pulse_api_key":    apiKey,
		"pulse_api_secret": apiSecret,
		"message":          "⚠️ IMPORTANT: Save your new API secret now. It won't be shown again. Your old keys are now invalid.",
	})
}
