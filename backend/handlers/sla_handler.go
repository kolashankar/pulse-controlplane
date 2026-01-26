package handlers

import (
	"net/http"
	"strconv"
	"time"

	"pulse-control-plane/database"
	"pulse-control-plane/models"
	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SLAHandler handles SLA endpoints
type SLAHandler struct {
	service *services.SLAService
}

// NewSLAHandler creates a new SLA handler
func NewSLAHandler() *SLAHandler {
	db := database.GetDB()
	return &SLAHandler{
		service: services.NewSLAService(db),
	}
}

// CreateSLATemplate creates a new SLA template
// @Summary Create SLA template
// @Description Create a new SLA template
// @Tags SLA
// @Accept json
// @Produce json
// @Param template body models.SLATemplate true "SLA Template"
// @Success 201 {object} models.SLATemplate
// @Router /api/v1/sla/templates [post]
func (h *SLAHandler) CreateSLATemplate(c *gin.Context) {
	var template models.SLATemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.service.CreateSLATemplate(c.Request.Context(), &template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, template)
}

// GetSLATemplates lists all SLA templates
// @Summary List SLA templates
// @Description Get all SLA templates
// @Tags SLA
// @Produce json
// @Param active query boolean false "Filter active templates only"
// @Success 200 {array} models.SLATemplate
// @Router /api/v1/sla/templates [get]
func (h *SLAHandler) GetSLATemplates(c *gin.Context) {
	activeOnly := c.Query("active") == "true"
	
	templates, err := h.service.GetSLATemplates(c.Request.Context(), activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"templates": templates, "count": len(templates)})
}

// AssignSLAToOrg assigns an SLA to an organization
// @Summary Assign SLA to organization
// @Description Assign an SLA template to an organization
// @Tags SLA
// @Accept json
// @Produce json
// @Param assignment body models.OrganizationSLA true "Organization SLA"
// @Success 201 {object} models.OrganizationSLA
// @Router /api/v1/sla/assign [post]
func (h *SLAHandler) AssignSLAToOrg(c *gin.Context) {
	var orgSLA models.OrganizationSLA
	if err := c.ShouldBindJSON(&orgSLA); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.service.AssignSLAToOrg(c.Request.Context(), &orgSLA); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, orgSLA)
}

// GetOrganizationSLA gets active SLA for an organization
// @Summary Get organization SLA
// @Description Get active SLA for an organization
// @Tags SLA
// @Produce json
// @Param org_id path string true "Organization ID"
// @Success 200 {object} models.OrganizationSLA
// @Router /api/v1/sla/organization/{org_id} [get]
func (h *SLAHandler) GetOrganizationSLA(c *gin.Context) {
	orgIDStr := c.Param("org_id")
	orgID, err := primitive.ObjectIDFromHex(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}
	
	sla, err := h.service.GetOrganizationSLA(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, sla)
}

// GetSLAReport generates SLA report for a period
// @Summary Get SLA report
// @Description Generate SLA performance report for a time period
// @Tags SLA
// @Produce json
// @Param org_id path string true "Organization ID"
// @Param period_start query string true "Period start (YYYY-MM-DD)"
// @Param period_end query string true "Period end (YYYY-MM-DD)"
// @Success 200 {object} models.SLAMetric
// @Router /api/v1/sla/report/{org_id} [get]
func (h *SLAHandler) GetSLAReport(c *gin.Context) {
	orgIDStr := c.Param("org_id")
	orgID, err := primitive.ObjectIDFromHex(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}
	
	periodStartStr := c.Query("period_start")
	periodEndStr := c.Query("period_end")
	
	periodStart, err := time.Parse("2006-01-02", periodStartStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_start format (use YYYY-MM-DD)"})
		return
	}
	
	periodEnd, err := time.Parse("2006-01-02", periodEndStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_end format (use YYYY-MM-DD)"})
		return
	}
	
	report, err := h.service.GetSLAReport(c.Request.Context(), orgID, periodStart, periodEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, report)
}

// GetSLABreaches gets SLA breaches for an organization
// @Summary Get SLA breaches
// @Description Get recent SLA breach events
// @Tags SLA
// @Produce json
// @Param org_id path string true "Organization ID"
// @Param limit query int false "Limit (default 20)"
// @Success 200 {array} models.SLABreach
// @Router /api/v1/sla/breaches/{org_id} [get]
func (h *SLAHandler) GetSLABreaches(c *gin.Context) {
	orgIDStr := c.Param("org_id")
	orgID, err := primitive.ObjectIDFromHex(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}
	
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	breaches, err := h.service.GetSLABreaches(c.Request.Context(), orgID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"breaches": breaches, "count": len(breaches)})
}
