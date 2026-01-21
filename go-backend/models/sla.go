package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SLATemplate defines service level agreement parameters
type SLATemplate struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name             string             `bson:"name" json:"name" binding:"required"`
	Description      string             `bson:"description" json:"description"`
	
	// Uptime commitments
	UptimePercent    float64 `bson:"uptime_percent" json:"uptime_percent" binding:"required"` // e.g., 99.9
	
	// Response time commitments (in milliseconds)
	APIResponseTime  int `bson:"api_response_time" json:"api_response_time"` // p95 in ms
	
	// Support response time commitments (in minutes)
	SupportResponseP0 int `bson:"support_response_p0" json:"support_response_p0"` // Critical
	SupportResponseP1 int `bson:"support_response_p1" json:"support_response_p1"` // High
	SupportResponseP2 int `bson:"support_response_p2" json:"support_response_p2"` // Medium
	SupportResponseP3 int `bson:"support_response_p3" json:"support_response_p3"` // Low
	
	// Credits for SLA breaches
	CreditPercent    float64 `bson:"credit_percent" json:"credit_percent"` // % of monthly fee
	
	IsActive         bool      `bson:"is_active" json:"is_active"`
	CreatedAt        time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time `bson:"updated_at" json:"updated_at"`
}

// OrganizationSLA assigns an SLA to an organization
type OrganizationSLA struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID          primitive.ObjectID `bson:"org_id" json:"org_id" binding:"required"`
	SLATemplateID  primitive.ObjectID `bson:"sla_template_id" json:"sla_template_id" binding:"required"`
	StartDate      time.Time          `bson:"start_date" json:"start_date"`
	EndDate        *time.Time         `bson:"end_date,omitempty" json:"end_date,omitempty"`
	IsActive       bool               `bson:"is_active" json:"is_active"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

// SLAMetric tracks actual performance against SLA
type SLAMetric struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID            primitive.ObjectID `bson:"org_id" json:"org_id"`
	SLAID            primitive.ObjectID `bson:"sla_id" json:"sla_id"`
	PeriodStart      time.Time          `bson:"period_start" json:"period_start"`
	PeriodEnd        time.Time          `bson:"period_end" json:"period_end"`
	
	// Actual metrics
	ActualUptime     float64 `bson:"actual_uptime" json:"actual_uptime"`
	ActualResponseP95 int    `bson:"actual_response_p95" json:"actual_response_p95"`
	
	// SLA compliance
	UptimeMet        bool    `bson:"uptime_met" json:"uptime_met"`
	ResponseTimeMet  bool    `bson:"response_time_met" json:"response_time_met"`
	OverallCompliance bool   `bson:"overall_compliance" json:"overall_compliance"`
	
	// Breach details
	BreachCount      int     `bson:"breach_count" json:"breach_count"`
	DowntimeMinutes  int     `bson:"downtime_minutes" json:"downtime_minutes"`
	CreditEarned     float64 `bson:"credit_earned" json:"credit_earned"`
	
	CreatedAt        time.Time `bson:"created_at" json:"created_at"`
}

// SLABreach records SLA violation events
type SLABreach struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID       primitive.ObjectID `bson:"org_id" json:"org_id"`
	SLAID       primitive.ObjectID `bson:"sla_id" json:"sla_id"`
	BreachType  string             `bson:"breach_type" json:"breach_type"` // uptime, response_time, support_response
	Severity    string             `bson:"severity" json:"severity"` // minor, major, critical
	Description string             `bson:"description" json:"description"`
	StartTime   time.Time          `bson:"start_time" json:"start_time"`
	EndTime     *time.Time         `bson:"end_time,omitempty" json:"end_time,omitempty"`
	Duration    int                `bson:"duration" json:"duration"` // in minutes
	Resolved    bool               `bson:"resolved" json:"resolved"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}
