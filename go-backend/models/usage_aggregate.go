package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UsageAggregate represents hourly/daily aggregated usage metrics
type UsageAggregate struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID          primitive.ObjectID `bson:"project_id" json:"project_id"`
	PeriodType         string             `bson:"period_type" json:"period_type"` // hourly, daily, monthly
	PeriodStart        time.Time          `bson:"period_start" json:"period_start"`
	PeriodEnd          time.Time          `bson:"period_end" json:"period_end"`
	ParticipantMinutes float64            `bson:"participant_minutes" json:"participant_minutes"`
	EgressMinutes      float64            `bson:"egress_minutes" json:"egress_minutes"`
	StorageGB          float64            `bson:"storage_gb" json:"storage_gb"`
	BandwidthGB        float64            `bson:"bandwidth_gb" json:"bandwidth_gb"`
	APIRequests        int64              `bson:"api_requests" json:"api_requests"`
	TotalCost          float64            `bson:"total_cost" json:"total_cost"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
}

// TableName returns the collection name
func (UsageAggregate) TableName() string {
	return "usage_aggregates"
}

// Period types
const (
	PeriodHourly  = "hourly"
	PeriodDaily   = "daily"
	PeriodMonthly = "monthly"
)

// PlanLimits represents usage limits for different plans
type PlanLimits struct {
	ID                       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PlanName                 string             `bson:"plan_name" json:"plan_name"`
	MaxParticipantMinutes    float64            `bson:"max_participant_minutes" json:"max_participant_minutes"`
	MaxEgressMinutes         float64            `bson:"max_egress_minutes" json:"max_egress_minutes"`
	MaxStorageGB             float64            `bson:"max_storage_gb" json:"max_storage_gb"`
	MaxBandwidthGB           float64            `bson:"max_bandwidth_gb" json:"max_bandwidth_gb"`
	MaxAPIRequests           int64              `bson:"max_api_requests" json:"max_api_requests"`
	AlertThresholdPercentage int                `bson:"alert_threshold_percentage" json:"alert_threshold_percentage"` // e.g., 80
	CreatedAt                time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt                time.Time          `bson:"updated_at" json:"updated_at"`
}

// TableName returns the collection name
func (PlanLimits) TableName() string {
	return "plan_limits"
}

// Default plan limits
var (
	FreePlanLimits = PlanLimits{
		PlanName:                 "Free",
		MaxParticipantMinutes:    1000,    // 1000 minutes
		MaxEgressMinutes:         100,     // 100 minutes
		MaxStorageGB:             1,       // 1 GB
		MaxBandwidthGB:           10,      // 10 GB
		MaxAPIRequests:           10000,   // 10k requests
		AlertThresholdPercentage: 80,      // Alert at 80%
	}

	ProPlanLimits = PlanLimits{
		PlanName:                 "Pro",
		MaxParticipantMinutes:    100000,  // 100k minutes
		MaxEgressMinutes:         10000,   // 10k minutes
		MaxStorageGB:             100,     // 100 GB
		MaxBandwidthGB:           1000,    // 1 TB
		MaxAPIRequests:           1000000, // 1M requests
		AlertThresholdPercentage: 80,
	}

	EnterprisePlanLimits = PlanLimits{
		PlanName:                 "Enterprise",
		MaxParticipantMinutes:    -1, // Unlimited
		MaxEgressMinutes:         -1, // Unlimited
		MaxStorageGB:             -1, // Unlimited
		MaxBandwidthGB:           -1, // Unlimited
		MaxAPIRequests:           -1, // Unlimited
		AlertThresholdPercentage: 90, // Alert at 90%
	}
)

// UsageAlert represents an alert when approaching limits
type UsageAlert struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID   primitive.ObjectID `bson:"project_id" json:"project_id"`
	MetricType  string             `bson:"metric_type" json:"metric_type"` // participant_minutes, egress_minutes, etc.
	CurrentUsage float64           `bson:"current_usage" json:"current_usage"`
	Limit       float64            `bson:"limit" json:"limit"`
	Percentage  float64            `bson:"percentage" json:"percentage"`
	Severity    string             `bson:"severity" json:"severity"` // warning, critical
	Message     string             `bson:"message" json:"message"`
	Notified    bool               `bson:"notified" json:"notified"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

// TableName returns the collection name
func (UsageAlert) TableName() string {
	return "usage_alerts"
}

// Alert severities
const (
	SeverityWarning  = "warning"  // 80% threshold
	SeverityCritical = "critical" // 95% threshold
)
