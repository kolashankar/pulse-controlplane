package handlers

import (
	"net/http"
	"strconv"
	"time"

	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BillingHandler handles billing-related HTTP requests
type BillingHandler struct {
	billingService *services.BillingService
}

// NewBillingHandler creates a new billing handler
func NewBillingHandler(billingService *services.BillingService) *BillingHandler {
	return &BillingHandler{
		billingService: billingService,
	}
}

// GetBillingDashboard retrieves billing dashboard data for a project
// GET /v1/billing/:project_id/dashboard
func (h *BillingHandler) GetBillingDashboard(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get dashboard data
	dashboard, err := h.billingService.GetBillingDashboard(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve billing dashboard"})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// GenerateInvoice generates an invoice for a project
// POST /v1/billing/:project_id/invoice
func (h *BillingHandler) GenerateInvoice(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req struct {
		PeriodStart string `json:"period_start" binding:"required"`
		PeriodEnd   string `json:"period_end" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse dates
	periodStart, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period_start format (use YYYY-MM-DD)"})
		return
	}

	periodEnd, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period_end format (use YYYY-MM-DD)"})
		return
	}

	// Generate invoice
	invoice, err := h.billingService.GenerateInvoice(c.Request.Context(), projectID, periodStart, periodEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate invoice"})
		return
	}

	c.JSON(http.StatusCreated, invoice)
}

// GetInvoice retrieves an invoice by ID
// GET /v1/billing/invoice/:invoice_id
func (h *BillingHandler) GetInvoice(c *gin.Context) {
	invoiceIDStr := c.Param("invoice_id")
	invoiceID, err := primitive.ObjectIDFromHex(invoiceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	// Get invoice
	invoice, err := h.billingService.GetInvoice(c.Request.Context(), invoiceID)
	if err != nil {
		if err.Error() == "invoice not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve invoice"})
		}
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// ListInvoices retrieves invoices for a project
// GET /v1/billing/:project_id/invoices
func (h *BillingHandler) ListInvoices(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// List invoices
	invoices, total, err := h.billingService.ListInvoices(c.Request.Context(), projectID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve invoices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invoices": invoices,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// UpdateInvoiceStatus updates the status of an invoice
// PUT /v1/billing/invoice/:invoice_id/status
func (h *BillingHandler) UpdateInvoiceStatus(c *gin.Context) {
	invoiceIDStr := c.Param("invoice_id")
	invoiceID, err := primitive.ObjectIDFromHex(invoiceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update status
	err = h.billingService.UpdateInvoiceStatus(c.Request.Context(), invoiceID, req.Status)
	if err != nil {
		if err.Error() == "invoice not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update invoice status"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invoice status updated successfully",
		"status":  req.Status,
	})
}

// IntegrateStripe is a placeholder for Stripe integration
// POST /v1/billing/:project_id/stripe/integrate
func (h *BillingHandler) IntegrateStripe(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "Stripe integration coming soon",
		"status":  "placeholder",
	})
}

// CreateStripeCustomer is a placeholder for creating a Stripe customer
// POST /v1/billing/stripe/customer
func (h *BillingHandler) CreateStripeCustomer(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "Stripe customer creation coming soon",
		"status":  "placeholder",
	})
}

// AttachPaymentMethod is a placeholder for attaching payment method
// POST /v1/billing/stripe/payment-method
func (h *BillingHandler) AttachPaymentMethod(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "Payment method attachment coming soon",
		"status":  "placeholder",
	})
}
