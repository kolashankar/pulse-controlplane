package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UsageMetric represents a usage event for billing
type UsageMetric struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID     `bson:"project_id" json:"project_id"`
	EventType string                 `bson:"event_type" json:"event_type"` // participant_joined, egress_started, etc.
	Value     float64                `bson:"value" json:"value"`           // duration in minutes, size in GB, etc.
	Timestamp time.Time              `bson:"timestamp" json:"timestamp"`
	Metadata  map[string]interface{} `bson:"metadata" json:"metadata"`
}

// UsageMetricCreate represents the input for creating a usage metric
type UsageMetricCreate struct {
	ProjectID string                 `json:"project_id" binding:"required"`
	EventType string                 `json:"event_type" binding:"required"`
	Value     float64                `json:"value" binding:"required"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// UsageSummary represents aggregated usage for a project
type UsageSummary struct {
	ProjectID         string    `json:"project_id"`
	ParticipantMinutes float64   `json:"participant_minutes"`
	EgressMinutes     float64   `json:"egress_minutes"`
	StorageGB         float64   `json:"storage_gb"`
	BandwidthGB       float64   `json:"bandwidth_gb"`
	APIRequests       int64     `json:"api_requests"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	TotalCost         float64   `json:"total_cost"`
}

// TableName returns the collection name
func (UsageMetric) TableName() string {
	return "usage_metrics"
}

// Event types for usage tracking
const (
	EventParticipantJoined = "participant_joined"
	EventParticipantLeft   = "participant_left"
	EventRoomStarted       = "room_started"
	EventRoomEnded         = "room_ended"
	EventEgressStarted     = "egress_started"
	EventEgressEnded       = "egress_ended"
	EventRecordingStarted  = "recording_started"
	EventRecordingEnded    = "recording_ended"
	EventStorageUsed       = "storage_used"
	EventBandwidthUsed     = "bandwidth_used"
	EventAPIRequest        = "api_request"
)
