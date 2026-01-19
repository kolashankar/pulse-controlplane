package services

import (
	"context"
	"fmt"
	"time"

	"pulse-control-plane/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BillingService handles billing operations
type BillingService struct {
	db           *mongo.Database
	usageService *UsageService
}

// NewBillingService creates a new billing service
func NewBillingService(db *mongo.Database, usageService *UsageService) *BillingService {
	return &BillingService{
		db:           db,
		usageService: usageService,
	}
}

// CalculateCost calculates the cost based on usage and plan
func (s *BillingService) CalculateCost(ctx context.Context, summary *models.UsageSummary, plan string) float64 {
	var pricing models.PricingModel

	switch plan {
	case "Free":
		pricing = models.FreePricing
	case "Pro":
		pricing = models.ProPricing
	case "Enterprise":
		pricing = models.EnterprisePricing
	default:
		pricing = models.FreePricing
	}

	totalCost := 0.0

	// Participant minutes cost
	totalCost += summary.ParticipantMinutes * pricing.ParticipantMinutePrice

	// Egress minutes cost
	totalCost += summary.EgressMinutes * pricing.EgressMinutePrice

	// Storage cost
	totalCost += summary.StorageGB * pricing.StorageGBPrice

	// Bandwidth cost
	totalCost += summary.BandwidthGB * pricing.BandwidthGBPrice

	// API requests cost (per 1000)
	totalCost += (float64(summary.APIRequests) / 1000.0) * pricing.APIRequestPrice

	// Add monthly base price (prorated for the period)
	if pricing.MonthlyBasePrice > 0 {
		// Calculate days in period
		periodDays := summary.EndDate.Sub(summary.StartDate).Hours() / 24
		monthlyProration := periodDays / 30.0 // Assume 30 days per month
		totalCost += pricing.MonthlyBasePrice * monthlyProration
	}

	return totalCost
}

// GenerateInvoice generates an invoice for a project
func (s *BillingService) GenerateInvoice(ctx context.Context, projectID primitive.ObjectID, periodStart, periodEnd time.Time) (*models.Invoice, error) {
	// Get project details
	projectCollection := s.db.Collection(models.Project{}.TableName())
	var project models.Project
	err := projectCollection.FindOne(ctx, bson.M{"_id": projectID}).Decode(&project)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// Get organization details
	orgCollection := s.db.Collection(models.Organization{}.TableName())
	var org models.Organization
	err = orgCollection.FindOne(ctx, bson.M{"_id": project.OrgID}).Decode(&org)
	if err != nil {
		return nil, fmt.Errorf("failed to find organization: %w", err)
	}

	// Get usage summary
	summary, err := s.usageService.GetUsageSummary(ctx, projectID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage summary: %w", err)
	}

	// Get pricing model
	var pricing models.PricingModel
	switch org.Plan {
	case "Free":
		pricing = models.FreePricing
	case "Pro":
		pricing = models.ProPricing
	case "Enterprise":
		pricing = models.EnterprisePricing
	default:
		pricing = models.FreePricing
	}

	// Create line items
	lineItems := []models.InvoiceLineItem{}

	// Participant minutes
	if summary.ParticipantMinutes > 0 {
		lineItems = append(lineItems, models.InvoiceLineItem{
			Description: "Participant Minutes",
			Quantity:    summary.ParticipantMinutes,
			Unit:        "minutes",
			UnitPrice:   pricing.ParticipantMinutePrice,
			Amount:      summary.ParticipantMinutes * pricing.ParticipantMinutePrice,
		})
	}

	// Egress minutes
	if summary.EgressMinutes > 0 {
		lineItems = append(lineItems, models.InvoiceLineItem{
			Description: "Egress/Streaming Minutes",
			Quantity:    summary.EgressMinutes,
			Unit:        "minutes",
			UnitPrice:   pricing.EgressMinutePrice,
			Amount:      summary.EgressMinutes * pricing.EgressMinutePrice,
		})
	}

	// Storage
	if summary.StorageGB > 0 {
		lineItems = append(lineItems, models.InvoiceLineItem{
			Description: "Storage",
			Quantity:    summary.StorageGB,
			Unit:        "GB",
			UnitPrice:   pricing.StorageGBPrice,
			Amount:      summary.StorageGB * pricing.StorageGBPrice,
		})
	}

	// Bandwidth
	if summary.BandwidthGB > 0 {
		lineItems = append(lineItems, models.InvoiceLineItem{
			Description: "Bandwidth",
			Quantity:    summary.BandwidthGB,
			Unit:        "GB",
			UnitPrice:   pricing.BandwidthGBPrice,
			Amount:      summary.BandwidthGB * pricing.BandwidthGBPrice,
		})
	}

	// API requests
	if summary.APIRequests > 0 {
		quantity := float64(summary.APIRequests) / 1000.0
		lineItems = append(lineItems, models.InvoiceLineItem{
			Description: "API Requests",
			Quantity:    quantity,
			Unit:        "per 1000",
			UnitPrice:   pricing.APIRequestPrice,
			Amount:      quantity * pricing.APIRequestPrice,
		})
	}

	// Monthly base price (prorated)
	if pricing.MonthlyBasePrice > 0 {
		periodDays := periodEnd.Sub(periodStart).Hours() / 24
		monthlyProration := periodDays / 30.0
		proratedAmount := pricing.MonthlyBasePrice * monthlyProration
		lineItems = append(lineItems, models.InvoiceLineItem{
			Description: fmt.Sprintf("Monthly Subscription (%s Plan)", org.Plan),
			Quantity:    monthlyProration,
			Unit:        "month",
			UnitPrice:   pricing.MonthlyBasePrice,
			Amount:      proratedAmount,
		})
	}

	// Calculate totals
	subtotal := 0.0
	for _, item := range lineItems {
		subtotal += item.Amount
	}

	tax := 0.0 // Tax calculation can be added here
	total := subtotal + tax

	// Generate invoice number
	invoiceNumber := fmt.Sprintf("INV-%s-%s", projectID.Hex()[:8], time.Now().Format("200601"))

	// Create invoice
	invoice := models.Invoice{
		ProjectID:     projectID,
		OrgID:         project.OrgID,
		InvoiceNumber: invoiceNumber,
		BillingPeriod: periodStart.Format("2006-01"),
		PeriodStart:   periodStart,
		PeriodEnd:     periodEnd,
		Status:        models.InvoiceStatusDraft,
		LineItems:     lineItems,
		Subtotal:      subtotal,
		Tax:           tax,
		Total:         total,
		Currency:      pricing.Currency,
		DueDate:       time.Now().Add(30 * 24 * time.Hour), // 30 days
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Insert invoice
	invoiceCollection := s.db.Collection(models.Invoice{}.TableName())
	result, err := invoiceCollection.InsertOne(ctx, invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	invoice.ID = result.InsertedID.(primitive.ObjectID)

	return &invoice, nil
}

// GetInvoice retrieves an invoice by ID
func (s *BillingService) GetInvoice(ctx context.Context, invoiceID primitive.ObjectID) (*models.Invoice, error) {
	collection := s.db.Collection(models.Invoice{}.TableName())

	var invoice models.Invoice
	err := collection.FindOne(ctx, bson.M{"_id": invoiceID}).Decode(&invoice)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return &invoice, nil
}

// ListInvoices retrieves invoices for a project
func (s *BillingService) ListInvoices(ctx context.Context, projectID primitive.ObjectID, page, limit int) ([]models.Invoice, int64, error) {
	collection := s.db.Collection(models.Invoice{}.TableName())

	filter := bson.M{"project_id": projectID}

	// Count total
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Find with pagination
	opts := options.Find()
	opts.SetSkip(int64((page - 1) * limit))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find invoices: %w", err)
	}
	defer cursor.Close(ctx)

	var invoices []models.Invoice
	if err := cursor.All(ctx, &invoices); err != nil {
		return nil, 0, fmt.Errorf("failed to decode invoices: %w", err)
	}

	return invoices, total, nil
}

// UpdateInvoiceStatus updates the status of an invoice
func (s *BillingService) UpdateInvoiceStatus(ctx context.Context, invoiceID primitive.ObjectID, status string) error {
	collection := s.db.Collection(models.Invoice{}.TableName())

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	if status == models.InvoiceStatusPaid {
		now := time.Now()
		update["$set"].(bson.M)["paid_at"] = &now
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": invoiceID}, update)
	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

// GetBillingDashboard retrieves billing dashboard data for a project
func (s *BillingService) GetBillingDashboard(ctx context.Context, projectID primitive.ObjectID) (*models.BillingDashboardResponse, error) {
	// Get project details
	projectCollection := s.db.Collection(models.Project{}.TableName())
	var project models.Project
	err := projectCollection.FindOne(ctx, bson.M{"_id": projectID}).Decode(&project)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// Get organization details
	orgCollection := s.db.Collection(models.Organization{}.TableName())
	var org models.Organization
	err = orgCollection.FindOne(ctx, bson.M{"_id": project.OrgID}).Decode(&org)
	if err != nil {
		return nil, fmt.Errorf("failed to find organization: %w", err)
	}

	// Get current month usage
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	periodEnd := now

	usageSummary, err := s.usageService.GetUsageSummary(ctx, projectID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage summary: %w", err)
	}

	// Calculate current charges
	currentCharges := s.CalculateCost(ctx, usageSummary, org.Plan)

	// Project total for end of month
	daysInMonth := 30.0
	daysPassed := float64(now.Day())
	projectionMultiplier := daysInMonth / daysPassed
	projectedTotal := currentCharges * projectionMultiplier

	// Get plan limits
	var planLimits models.PlanLimits
	switch org.Plan {
	case "Free":
		planLimits = models.FreePlanLimits
	case "Pro":
		planLimits = models.ProPlanLimits
	case "Enterprise":
		planLimits = models.EnterprisePlanLimits
	default:
		planLimits = models.FreePlanLimits
	}

	// Get alerts
	alerts, _ := s.usageService.GetAlerts(ctx, projectID)
	if alerts == nil {
		alerts = []models.UsageAlert{}
	}

	// Get recent invoices
	recentInvoices, _, _ := s.ListInvoices(ctx, projectID, 1, 5)
	if recentInvoices == nil {
		recentInvoices = []models.Invoice{}
	}

	dashboard := &models.BillingDashboardResponse{
		ProjectID:      projectID.Hex(),
		CurrentPlan:    org.Plan,
		BillingPeriod:  periodStart.Format("2006-01"),
		UsageSummary:   *usageSummary,
		CurrentCharges: currentCharges,
		ProjectedTotal: projectedTotal,
		PlanLimits:     planLimits,
		Alerts:         alerts,
		RecentInvoices: recentInvoices,
	}

	return dashboard, nil
}

// IntegrateStripe is a placeholder for Stripe integration
func (s *BillingService) IntegrateStripe(ctx context.Context, invoiceID primitive.ObjectID) error {
	// TODO: Implement Stripe integration
	// 1. Create Stripe invoice
	// 2. Send to customer email
	// 3. Update invoice with Stripe invoice ID
	// 4. Listen for payment webhooks
	// 5. Update invoice status when paid

	return fmt.Errorf("stripe integration not implemented yet")
}

// CreateStripeCustomer is a placeholder for creating a Stripe customer
func (s *BillingService) CreateStripeCustomer(ctx context.Context, orgID primitive.ObjectID) (string, error) {
	// TODO: Implement Stripe customer creation
	// 1. Get organization details
	// 2. Create Stripe customer with email
	// 3. Store Stripe customer ID in organization
	// 4. Return customer ID

	return "", fmt.Errorf("stripe customer creation not implemented yet")
}

// AttachPaymentMethod is a placeholder for attaching payment method
func (s *BillingService) AttachPaymentMethod(ctx context.Context, orgID primitive.ObjectID, paymentMethodID string) error {
	// TODO: Implement payment method attachment
	// 1. Get organization's Stripe customer ID
	// 2. Attach payment method to customer
	// 3. Set as default payment method

	return fmt.Errorf("payment method attachment not implemented yet")
}
