package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuditLog represents a log entry for auditing actions
type AuditLog struct {
	ID           primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	OrgID        primitive.ObjectID     `bson:"org_id" json:"org_id"`
	UserID       primitive.ObjectID     `bson:"user_id,omitempty" json:"user_id,omitempty"`
	UserEmail    string                 `bson:"user_email" json:"user_email"`
	Action       string                 `bson:"action" json:"action"` // e.g., "project.created", "api_key.regenerated"
	Resource     string                 `bson:"resource" json:"resource"` // e.g., "project", "team_member"
	ResourceID   string                 `bson:"resource_id" json:"resource_id"`
	ResourceName string                 `bson:"resource_name,omitempty" json:"resource_name,omitempty"`
	IPAddress    string                 `bson:"ip_address" json:"ip_address"`
	UserAgent    string                 `bson:"user_agent" json:"user_agent"`
	Status       string                 `bson:"status" json:"status"` // Success, Failed
	Details      map[string]interface{} `bson:"details,omitempty" json:"details,omitempty"`
	Timestamp    time.Time              `bson:"timestamp" json:"timestamp"`
	CreatedAt    time.Time              `bson:"created_at" json:"created_at"`
}

// AuditLogFilter represents filters for querying audit logs
type AuditLogFilter struct {
	OrgID      string    `form:"org_id"`
	UserEmail  string    `form:"user_email"`
	Action     string    `form:"action"`
	Resource   string    `form:"resource"`
	ResourceID string    `form:"resource_id"`
	Status     string    `form:"status"`
	StartDate  time.Time `form:"start_date"`
	EndDate    time.Time `form:"end_date"`
	Page       int       `form:"page"`
	Limit      int       `form:"limit"`
}

// AuditLogResponse represents an audit log in API responses
type AuditLogResponse struct {
	ID           string                 `json:"id"`
	UserEmail    string                 `json:"user_email"`
	Action       string                 `json:"action"`
	Resource     string                 `json:"resource"`
	ResourceID   string                 `json:"resource_id"`
	ResourceName string                 `json:"resource_name,omitempty"`
	IPAddress    string                 `json:"ip_address"`
	Status       string                 `json:"status"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// ToResponse converts AuditLog to AuditLogResponse
func (al *AuditLog) ToResponse() AuditLogResponse {
	return AuditLogResponse{
		ID:           al.ID.Hex(),
		UserEmail:    al.UserEmail,
		Action:       al.Action,
		Resource:     al.Resource,
		ResourceID:   al.ResourceID,
		ResourceName: al.ResourceName,
		IPAddress:    al.IPAddress,
		Status:       al.Status,
		Details:      al.Details,
		Timestamp:    al.Timestamp,
	}
}

// AuditActions defines common audit actions
var AuditActions = struct {
	// Project actions
	ProjectCreated string
	ProjectUpdated string
	ProjectDeleted string

	// API Key actions
	APIKeyRegenerated string

	// Team actions
	TeamMemberInvited string
	TeamMemberAdded   string
	TeamMemberRemoved string
	TeamMemberUpdated string

	// Organization actions
	OrganizationCreated string
	OrganizationUpdated string
	OrganizationDeleted string

	// Settings actions
	SettingsUpdated string

	// Webhook actions
	WebhookConfigured string
	WebhookUpdated    string
	WebhookDeleted    string

	// Billing actions
	BillingUpdated     string
	InvoiceGenerated   string
	PaymentMethodAdded string
}{
	ProjectCreated:     "project.created",
	ProjectUpdated:     "project.updated",
	ProjectDeleted:     "project.deleted",
	APIKeyRegenerated:  "api_key.regenerated",
	TeamMemberInvited:  "team_member.invited",
	TeamMemberAdded:    "team_member.added",
	TeamMemberRemoved:  "team_member.removed",
	TeamMemberUpdated:  "team_member.updated",
	OrganizationCreated: "organization.created",
	OrganizationUpdated: "organization.updated",
	OrganizationDeleted: "organization.deleted",
	SettingsUpdated:     "settings.updated",
	WebhookConfigured:   "webhook.configured",
	WebhookUpdated:      "webhook.updated",
	WebhookDeleted:      "webhook.deleted",
	BillingUpdated:      "billing.updated",
	InvoiceGenerated:    "invoice.generated",
	PaymentMethodAdded:  "payment_method.added",
}

// TableName returns the collection name
func (AuditLog) TableName() string {
	return "audit_logs"
}
