package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WebhookEventType defines the type of webhook event
type WebhookEventType string

const (
	WebhookEventParticipantJoined WebhookEventType = "participant_joined"
	WebhookEventParticipantLeft WebhookEventType = "participant_left"
	WebhookEventRoomStarted WebhookEventType = "room_started"
	WebhookEventRoomEnded WebhookEventType = "room_ended"
	WebhookEventEgressStarted WebhookEventType = "egress_started"
	WebhookEventEgressEnded WebhookEventType = "egress_ended"
	WebhookEventRecordingAvailable WebhookEventType = "recording_available"
	WebhookEventIngressStarted WebhookEventType = "ingress_started"
	WebhookEventIngressEnded WebhookEventType = "ingress_ended"
)

// WebhookDeliveryStatus represents the delivery status
type WebhookDeliveryStatus string

const (
	WebhookStatusPending WebhookDeliveryStatus = "pending"
	WebhookStatusDelivered WebhookDeliveryStatus = "delivered"
	WebhookStatusFailed WebhookDeliveryStatus = "failed"
	WebhookStatusRetrying WebhookDeliveryStatus = "retrying"
)

// WebhookLog represents a webhook delivery log
type WebhookLog struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID `bson:"project_id" json:"project_id"`
	
	// Event details
	EventType WebhookEventType `bson:"event_type" json:"event_type"`
	Payload map[string]interface{} `bson:"payload" json:"payload"`
	
	// Delivery details
	WebhookURL string `bson:"webhook_url" json:"webhook_url"`
	Status WebhookDeliveryStatus `bson:"status" json:"status"`
	Attempts int `bson:"attempts" json:"attempts"`
	MaxAttempts int `bson:"max_attempts" json:"max_attempts"`
	
	// Response details
	ResponseStatus int `bson:"response_status" json:"response_status"`
	ResponseBody string `bson:"response_body,omitempty" json:"response_body,omitempty"`
	Error string `bson:"error,omitempty" json:"error,omitempty"`
	
	// Retry details
	NextRetryAt *time.Time `bson:"next_retry_at,omitempty" json:"next_retry_at,omitempty"`
	LastAttemptAt *time.Time `bson:"last_attempt_at,omitempty" json:"last_attempt_at,omitempty"`
	
	// HMAC signature for verification
	Signature string `bson:"signature" json:"signature"`
	
	// Timestamps
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// WebhookPayload represents the structure of a webhook payload
type WebhookPayload struct {
	Event WebhookEventType `json:"event"`
	Timestamp int64 `json:"timestamp"`
	ProjectID string `json:"project_id"`
	RoomName string `json:"room_name,omitempty"`
	Participant *ParticipantInfo `json:"participant,omitempty"`
	Egress *EgressInfo `json:"egress,omitempty"`
	Ingress *IngressInfo `json:"ingress,omitempty"`
	Recording *RecordingInfo `json:"recording,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ParticipantInfo contains participant details for webhooks
type ParticipantInfo struct {
	ID string `json:"id"`
	Identity string `json:"identity"`
	Name string `json:"name"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// EgressInfo contains egress details for webhooks
type EgressInfo struct {
	ID string `json:"id"`
	RoomName string `json:"room_name"`
	Status string `json:"status"`
	OutputURL string `json:"output_url,omitempty"`
	Duration int64 `json:"duration,omitempty"`
	Error string `json:"error,omitempty"`
}

// IngressInfo contains ingress details for webhooks
type IngressInfo struct {
	ID string `json:"id"`
	RoomName string `json:"room_name"`
	Status string `json:"status"`
	Error string `json:"error,omitempty"`
}

// RecordingInfo contains recording details for webhooks
type RecordingInfo struct {
	ID string `json:"id"`
	RoomName string `json:"room_name"`
	URL string `json:"url"`
	Duration int64 `json:"duration"`
	FileSize int64 `json:"file_size"`
}
