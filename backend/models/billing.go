package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Invoice represents a billing invoice
type Invoice struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID       primitive.ObjectID `bson:"project_id" json:"project_id"`
	OrgID           primitive.ObjectID `bson:"org_id" json:"org_id"`
	InvoiceNumber   string             `bson:"invoice_number" json:"invoice_number"`
	BillingPeriod   string             `bson:"billing_period" json:"billing_period"` // e.g., "2025-01"
	PeriodStart     time.Time          `bson:"period_start" json:"period_start"`
	PeriodEnd       time.Time          `bson:"period_end" json:"period_end"`
	Status          string             `bson:"status" json:"status"` // draft, issued, paid, overdue
	LineItems       []InvoiceLineItem  `bson:"line_items" json:"line_items"`
	Subtotal        float64            `bson:"subtotal" json:"subtotal"`
	Tax             float64            `bson:"tax" json:"tax"`
	Total           float64            `bson:"total" json:"total"`
	Currency        string             `bson:"currency" json:"currency"`
	DueDate         time.Time          `bson:"due_date" json:"due_date"`
	PaidAt          *time.Time         `bson:"paid_at,omitempty" json:"paid_at,omitempty"`
	StripeInvoiceID string             `bson:"stripe_invoice_id,omitempty" json:"stripe_invoice_id,omitempty"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

// InvoiceLineItem represents a single line item in an invoice
type InvoiceLineItem struct {
	Description string  `bson:"description" json:"description"`
	Quantity    float64 `bson:"quantity" json:"quantity"`
	Unit        string  `bson:"unit" json:"unit"` // minutes, GB, requests
	UnitPrice   float64 `bson:"unit_price" json:"unit_price"`
	Amount      float64 `bson:"amount" json:"amount"`
}

// TableName returns the collection name
func (Invoice) TableName() string {
	return "invoices"
}

// Invoice statuses
const (
	InvoiceStatusDraft   = "draft"
	InvoiceStatusIssued  = "issued"
	InvoiceStatusPaid    = "paid"
	InvoiceStatusOverdue = "overdue"
	InvoiceStatusVoid    = "void"
)

// PricingModel represents the pricing structure
type PricingModel struct {
	ID                      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PlanName                string             `bson:"plan_name" json:"plan_name"`
	ParticipantMinutePrice  float64            `bson:"participant_minute_price" json:"participant_minute_price"`   // per minute
	EgressMinutePrice       float64            `bson:"egress_minute_price" json:"egress_minute_price"`             // per minute
	StorageGBPrice          float64            `bson:"storage_gb_price" json:"storage_gb_price"`                   // per GB per month
	BandwidthGBPrice        float64            `bson:"bandwidth_gb_price" json:"bandwidth_gb_price"`               // per GB
	APIRequestPrice         float64            `bson:"api_request_price" json:"api_request_price"`                 // per 1000 requests
	MonthlyBasePrice        float64            `bson:"monthly_base_price" json:"monthly_base_price"`               // monthly subscription
	Currency                string             `bson:"currency" json:"currency"`
	CreatedAt               time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt               time.Time          `bson:"updated_at" json:"updated_at"`
}

// TableName returns the collection name
func (PricingModel) TableName() string {
	return "pricing_models"
}

// Default pricing (example rates)
var (
	FreePricing = PricingModel{
		PlanName:                "Free",
		ParticipantMinutePrice:  0,
		EgressMinutePrice:       0,
		StorageGBPrice:          0,
		BandwidthGBPrice:        0,
		APIRequestPrice:         0,
		MonthlyBasePrice:        0,
		Currency:                "USD",
	}

	ProPricing = PricingModel{
		PlanName:                "Pro",
		ParticipantMinutePrice:  0.004, // $0.004 per minute ($0.24/hour)
		EgressMinutePrice:       0.012, // $0.012 per minute ($0.72/hour)
		StorageGBPrice:          0.10,  // $0.10 per GB per month
		BandwidthGBPrice:        0.05,  // $0.05 per GB
		APIRequestPrice:         0.001, // $0.001 per 1000 requests
		MonthlyBasePrice:        49.00, // $49/month base
		Currency:                "USD",
	}

	EnterprisePricing = PricingModel{
		PlanName:                "Enterprise",
		ParticipantMinutePrice:  0.003, // $0.003 per minute (volume discount)
		EgressMinutePrice:       0.010, // $0.010 per minute
		StorageGBPrice:          0.08,  // $0.08 per GB per month
		BandwidthGBPrice:        0.04,  // $0.04 per GB
		APIRequestPrice:         0.0008, // $0.0008 per 1000 requests
		MonthlyBasePrice:        299.00, // $299/month base
		Currency:                "USD",
	}
)

// BillingDashboardResponse represents the billing dashboard data
type BillingDashboardResponse struct {
	ProjectID       string                `json:"project_id"`
	CurrentPlan     string                `json:"current_plan"`
	BillingPeriod   string                `json:"billing_period"`
	UsageSummary    UsageSummary          `json:"usage_summary"`
	CurrentCharges  float64               `json:"current_charges"`
	ProjectedTotal  float64               `json:"projected_total"`
	PlanLimits      PlanLimits            `json:"plan_limits"`
	Alerts          []UsageAlert          `json:"alerts"`
	RecentInvoices  []Invoice             `json:"recent_invoices"`
}
