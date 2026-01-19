package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EgressType defines the type of egress
type EgressType string

const (
	EgressTypeRoomComposite EgressType = "room_composite"
	EgressTypeTrackComposite EgressType = "track_composite"
	EgressTypeTrack EgressType = "track"
)

// EgressStatus represents the current status of an egress
type EgressStatus string

const (
	EgressStatusPending EgressStatus = "pending"
	EgressStatusActive EgressStatus = "active"
	EgressStatusEnded EgressStatus = "ended"
	EgressStatusFailed EgressStatus = "failed"
)

// OutputType defines where the egress output goes
type OutputType string

const (
	OutputTypeHLS OutputType = "hls"
	OutputTypeRTMP OutputType = "rtmp"
	OutputTypeFile OutputType = "file"
)

// LayoutType for room composite
type LayoutType string

const (
	LayoutTypeSpeaker LayoutType = "speaker"
	LayoutTypeGrid LayoutType = "grid"
	LayoutTypeSingle LayoutType = "single"
)

// Egress represents a media egress configuration
type Egress struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID `bson:"project_id" json:"project_id"`
	RoomName string `bson:"room_name" json:"room_name"`
	
	// Egress configuration
	EgressType EgressType `bson:"egress_type" json:"egress_type"`
	OutputType OutputType `bson:"output_type" json:"output_type"`
	LayoutType LayoutType `bson:"layout_type" json:"layout_type"`
	
	// LiveKit egress ID
	LiveKitEgressID string `bson:"livekit_egress_id" json:"livekit_egress_id"`
	
	// Status and metadata
	Status EgressStatus `bson:"status" json:"status"`
	Error string `bson:"error,omitempty" json:"error,omitempty"`
	
	// Output URLs
	OutputURL string `bson:"output_url" json:"output_url"` // S3/R2 URL
	CDNPlaybackURL string `bson:"cdn_playback_url" json:"cdn_playback_url"` // CDN URL for playback
	RTMPURL string `bson:"rtmp_url,omitempty" json:"rtmp_url,omitempty"`
	
	// Storage configuration
	StorageBucket string `bson:"storage_bucket" json:"storage_bucket"`
	StorageRegion string `bson:"storage_region" json:"storage_region"`
	StorageAccessKey string `bson:"storage_access_key" json:"-"` // Never expose
	StorageSecretKey string `bson:"storage_secret_key" json:"-"` // Never expose
	
	// Metrics
	DurationSeconds int64 `bson:"duration_seconds" json:"duration_seconds"`
	FileSizeBytes int64 `bson:"file_size_bytes" json:"file_size_bytes"`
	
	// Timestamps
	StartedAt *time.Time `bson:"started_at,omitempty" json:"started_at,omitempty"`
	EndedAt *time.Time `bson:"ended_at,omitempty" json:"ended_at,omitempty"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// EgressRequest represents a request to start an egress
type EgressRequest struct {
	RoomName string `json:"room_name" binding:"required"`
	EgressType EgressType `json:"egress_type" binding:"required"`
	OutputType OutputType `json:"output_type" binding:"required"`
	LayoutType LayoutType `json:"layout_type"`
	RTMPURL string `json:"rtmp_url,omitempty"`
	Filename string `json:"filename,omitempty"`
}

// EgressResponse represents a safe egress response (without sensitive data)
type EgressResponse struct {
	ID string `json:"id"`
	ProjectID string `json:"project_id"`
	RoomName string `json:"room_name"`
	EgressType EgressType `json:"egress_type"`
	OutputType OutputType `json:"output_type"`
	LayoutType LayoutType `json:"layout_type"`
	Status EgressStatus `json:"status"`
	Error string `json:"error,omitempty"`
	CDNPlaybackURL string `json:"cdn_playback_url,omitempty"`
	DurationSeconds int64 `json:"duration_seconds"`
	FileSizeBytes int64 `json:"file_size_bytes"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt *time.Time `json:"ended_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
