package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TicketPriority represents support ticket priority levels
type TicketPriority string

const (
	PriorityP0 TicketPriority = "P0" // Critical - System down
	PriorityP1 TicketPriority = "P1" // High - Major feature broken
	PriorityP2 TicketPriority = "P2" // Medium - Feature impaired
	PriorityP3 TicketPriority = "P3" // Low - General questions
)

// TicketStatus represents support ticket status
type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "open"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusWaiting    TicketStatus = "waiting_customer"
	TicketStatusResolved   TicketStatus = "resolved"
	TicketStatusClosed     TicketStatus = "closed"
)

// SupportTicket represents a customer support ticket
type SupportTicket struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TicketNumber   string             `bson:"ticket_number" json:"ticket_number"` // e.g., PULSE-12345
	OrgID          primitive.ObjectID `bson:"org_id" json:"org_id" binding:"required"`
	ProjectID      *primitive.ObjectID `bson:"project_id,omitempty" json:"project_id,omitempty"`
	
	// Ticket details
	Subject        string          `bson:"subject" json:"subject" binding:"required"`
	Description    string          `bson:"description" json:"description" binding:"required"`
	Priority       TicketPriority  `bson:"priority" json:"priority" binding:"required"`
	Status         TicketStatus    `bson:"status" json:"status"`
	Category       string          `bson:"category" json:"category"` // technical, billing, general
	
	// People involved
	CreatedBy      primitive.ObjectID  `bson:"created_by" json:"created_by"`
	AssignedTo     *primitive.ObjectID `bson:"assigned_to,omitempty" json:"assigned_to,omitempty"`
	CreatorEmail   string              `bson:"creator_email" json:"creator_email"`
	
	// Time tracking
	CreatedAt      time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at" json:"updated_at"`
	FirstResponse  *time.Time `bson:"first_response,omitempty" json:"first_response,omitempty"`
	ResolvedAt     *time.Time `bson:"resolved_at,omitempty" json:"resolved_at,omitempty"`
	ClosedAt       *time.Time `bson:"closed_at,omitempty" json:"closed_at,omitempty"`
	
	// Response time tracking (in minutes)
	ResponseTime   int `bson:"response_time,omitempty" json:"response_time,omitempty"`
	ResolutionTime int `bson:"resolution_time,omitempty" json:"resolution_time,omitempty"`
	
	// SLA tracking
	SLABreached    bool `bson:"sla_breached" json:"sla_breached"`
	
	// Tags and metadata
	Tags           []string `bson:"tags,omitempty" json:"tags,omitempty"`
	Attachments    []string `bson:"attachments,omitempty" json:"attachments,omitempty"`
}

// TicketComment represents a comment on a support ticket
type TicketComment struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TicketID   primitive.ObjectID `bson:"ticket_id" json:"ticket_id" binding:"required"`
	AuthorID   primitive.ObjectID `bson:"author_id" json:"author_id"`
	AuthorName string             `bson:"author_name" json:"author_name"`
	AuthorType string             `bson:"author_type" json:"author_type"` // customer, agent, system
	Content    string             `bson:"content" json:"content" binding:"required"`
	IsInternal bool               `bson:"is_internal" json:"is_internal"` // Internal notes not visible to customer
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

// TicketStats represents support ticket statistics
type TicketStats struct {
	TotalTickets       int     `json:"total_tickets"`
	OpenTickets        int     `json:"open_tickets"`
	InProgressTickets  int     `json:"in_progress_tickets"`
	ResolvedTickets    int     `json:"resolved_tickets"`
	ClosedTickets      int     `json:"closed_tickets"`
	AvgResponseTime    float64 `json:"avg_response_time"` // in minutes
	AvgResolutionTime  float64 `json:"avg_resolution_time"` // in minutes
	SLAComplianceRate  float64 `json:"sla_compliance_rate"` // percentage
	CriticalTickets    int     `json:"critical_tickets"`
}
