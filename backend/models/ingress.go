package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IngressType defines the type of ingress
type IngressType string

const (
	IngressTypeRTMP IngressType = "rtmp"
	IngressTypeWHIP IngressType = "whip"
	IngressTypeURL IngressType = "url"
)

// IngressStatus represents the current status of an ingress
type IngressStatus string

const (
	IngressStatusActive IngressStatus = "active"
	IngressStatusInactive IngressStatus = "inactive"
	IngressStatusError IngressStatus = "error"
)

// Ingress represents a media ingress endpoint
type Ingress struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID `bson:"project_id" json:"project_id"`
	RoomName string `bson:"room_name" json:"room_name"`
	ParticipantName string `bson:"participant_name" json:"participant_name"`
	
	// Ingress configuration
	IngressType IngressType `bson:"ingress_type" json:"ingress_type"`
	
	// LiveKit ingress ID
	LiveKitIngressID string `bson:"livekit_ingress_id" json:"livekit_ingress_id"`
	
	// Status
	Status IngressStatus `bson:"status" json:"status"`
	Error string `bson:"error,omitempty" json:"error,omitempty"`
	
	// Ingress URLs
	RTMPURL string `bson:"rtmp_url,omitempty" json:"rtmp_url,omitempty"`
	RTMPStreamKey string `bson:"rtmp_stream_key,omitempty" json:"rtmp_stream_key,omitempty"`
	WHIPURL string `bson:"whip_url,omitempty" json:"whip_url,omitempty"`
	SourceURL string `bson:"source_url,omitempty" json:"source_url,omitempty"`
	
	// Media settings
	AudioEnabled bool `bson:"audio_enabled" json:"audio_enabled"`
	VideoEnabled bool `bson:"video_enabled" json:"video_enabled"`
	
	// Timestamps
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// IngressRequest represents a request to create an ingress
type IngressRequest struct {
	RoomName string `json:"room_name" binding:"required"`
	ParticipantName string `json:"participant_name" binding:"required"`
	IngressType IngressType `json:"ingress_type" binding:"required"`
	SourceURL string `json:"source_url,omitempty"`
	AudioEnabled bool `json:"audio_enabled"`
	VideoEnabled bool `json:"video_enabled"`
}

// IngressResponse represents a safe ingress response
type IngressResponse struct {
	ID string `json:"id"`
	ProjectID string `json:"project_id"`
	RoomName string `json:"room_name"`
	ParticipantName string `json:"participant_name"`
	IngressType IngressType `json:"ingress_type"`
	Status IngressStatus `json:"status"`
	Error string `json:"error,omitempty"`
	RTMPURL string `json:"rtmp_url,omitempty"`
	RTMPStreamKey string `json:"rtmp_stream_key,omitempty"`
	WHIPURL string `json:"whip_url,omitempty"`
	AudioEnabled bool `json:"audio_enabled"`
	VideoEnabled bool `json:"video_enabled"`
	CreatedAt time.Time `json:"created_at"`
}
