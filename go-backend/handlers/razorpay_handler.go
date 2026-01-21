package handlers

import (
	"io"
	"net/http"
	"strconv"

	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RazorpayHandler handles Razorpay-related HTTP requests
type RazorpayHandler struct {
	razorpayService *services.RazorpayService
}

// NewRazorpayHandler creates a new Razorpay handler
func NewRazorpayHandler(razorpayService *services.RazorpayService) *RazorpayHandler {
	return &RazorpayHandler{
		razorpayService: razorpayService,
	}
}

// CreateCustomer creates a new Razorpay customer
// POST /api/v1/billing/razorpay/customer
func (h *RazorpayHandler) CreateCustomer(c *gin.Context) {
	var req struct {
		OrgID     string `json:"org_id" binding:"required"`
		Name      string `json:"name" binding:"required"`
		Email     string `json:"email" binding:"required,email"`
		Contact   string `json:"contact"`
		GSTNumber string `json:"gst_number"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orgID, err := primitive.ObjectIDFromHex(req.OrgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	customer, err := h.razorpayService.CreateCustomer(
		c.Request.Context(),
		orgID,
		req.Name,
		req.Email,
		req.Contact,
		req.GSTNumber,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Customer created successfully",
		"customer": customer,
	})
}

// CreateSubscription creates a new Razorpay subscription
// POST /api/v1/billing/razorpay/subscription
func (h *RazorpayHandler) CreateSubscription(c *gin.Context) {
	var req struct {
		ProjectID  string `json:"project_id" binding:"required"`
		OrgID      string `json:"org_id" binding:"required"`
		PlanID     string `json:"plan_id" binding:"required"`
		CustomerID string `json:"customer_id" binding:"required"`
		TotalCount int    `json:"total_count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	projectID, err := primitive.ObjectIDFromHex(req.ProjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	orgID, err := primitive.ObjectIDFromHex(req.OrgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Default to 12 months if not specified
	totalCount := req.TotalCount
	if totalCount == 0 {
		totalCount = 12
	}

	subscription, err := h.razorpayService.CreateSubscription(
		c.Request.Context(),
		projectID,
		orgID,
		req.PlanID,
		req.CustomerID,
		totalCount,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Subscription created successfully",
		"subscription": subscription,
	})
}

// GeneratePaymentLink generates a payment link for an invoice
// POST /api/v1/billing/razorpay/payment-link
func (h *RazorpayHandler) GeneratePaymentLink(c *gin.Context) {
	var req struct {
		InvoiceID string `json:"invoice_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoiceID, err := primitive.ObjectIDFromHex(req.InvoiceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	paymentLink, err := h.razorpayService.GeneratePaymentLink(c.Request.Context(), invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate payment link", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Payment link generated successfully",
		"payment_link": paymentLink,
	})
}

// VerifyPayment verifies a Razorpay payment
// POST /api/v1/billing/razorpay/verify
func (h *RazorpayHandler) VerifyPayment(c *gin.Context) {
	var req struct {
		PaymentID string `json:"razorpay_payment_id" binding:"required"`
		OrderID   string `json:"razorpay_order_id" binding:"required"`
		Signature string `json:"razorpay_signature" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	verified, err := h.razorpayService.VerifyPayment(
		c.Request.Context(),
		req.PaymentID,
		req.OrderID,
		req.Signature,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment verification failed", "details": err.Error()})
		return
	}

	if !verified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment signature"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Payment verified successfully",
		"verified": true,
	})
}

// HandleWebhook handles Razorpay webhook events
// POST /api/v1/billing/razorpay/webhook
func (h *RazorpayHandler) HandleWebhook(c *gin.Context) {
	// Read webhook payload
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read webhook payload"})
		return
	}

	// Get signature from header
	signature := c.GetHeader("X-Razorpay-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing webhook signature"})
		return
	}

	// Process webhook
	err = h.razorpayService.ProcessWebhook(c.Request.Context(), payload, signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook processing failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook processed successfully"})
}

// GetInvoices retrieves Razorpay invoices for a project
// GET /api/v1/billing/razorpay/invoices
func (h *RazorpayHandler) GetInvoices(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing project_id parameter"})
		return
	}

	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	invoices, total, err := h.razorpayService.GetInvoices(c.Request.Context(), projectID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve invoices", "details": err.Error()})
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

// ProcessRefund processes a refund for a payment
// POST /api/v1/billing/razorpay/refund
func (h *RazorpayHandler) ProcessRefund(c *gin.Context) {
	var req struct {
		PaymentID string  `json:"payment_id" binding:"required"`
		Amount    float64 `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be greater than zero"})
		return
	}

	refundID, err := h.razorpayService.ProcessRefund(c.Request.Context(), req.PaymentID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process refund", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Refund processed successfully",
		"refund_id": refundID,
	})
}

// GetCustomer retrieves a customer by organization ID
// GET /api/v1/billing/razorpay/customer/:org_id
func (h *RazorpayHandler) GetCustomer(c *gin.Context) {
	orgIDStr := c.Param("org_id")
	orgID, err := primitive.ObjectIDFromHex(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	customer, err := h.razorpayService.GetCustomerByOrgID(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	c.JSON(http.StatusOK, customer)
}

// GetSubscription retrieves a subscription by project ID
// GET /api/v1/billing/razorpay/subscription/:project_id
func (h *RazorpayHandler) GetSubscription(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	subscription, err := h.razorpayService.GetSubscriptionByProjectID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}
