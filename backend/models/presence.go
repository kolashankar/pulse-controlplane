package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PresenceStatus represents user presence status
type PresenceStatus string

const (
	PresenceOnline  PresenceStatus = "online"
	PresenceOffline PresenceStatus = "offline"
	PresenceAway    PresenceStatus = "away"
	PresenceBusy    PresenceStatus = "busy"
)

// UserPresence represents user online/offline status
type UserPresence struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID    primitive.ObjectID `bson:"project_id" json:"project_id"`
	UserID       string             `bson:"user_id" json:"user_id"`
	Status       PresenceStatus     `bson:"status" json:"status"`
	StatusMessage string            `bson:"status_message" json:"status_message"`
	LastSeen     time.Time          `bson:"last_seen" json:"last_seen"`
	CurrentRoom  string             `bson:"current_room" json:"current_room"`
	Device       string             `bson:"device" json:"device"`       // web, mobile, desktop
	IPAddress    string             `bson:"ip_address" json:"ip_address"`
	UserAgent    string             `bson:"user_agent" json:"user_agent"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// TypingIndicator represents a user typing indicator
type TypingIndicator struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID `bson:"project_id" json:"project_id"`
	RoomID    string             `bson:"room_id" json:"room_id"`
	UserID    string             `bson:"user_id" json:"user_id"`
	IsTyping  bool               `bson:"is_typing" json:"is_typing"`
	ExpiresAt time.Time          `bson:"expires_at" json:"expires_at"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// UserActivity represents user's current activity
type UserActivity struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID    primitive.ObjectID `bson:"project_id" json:"project_id"`
	UserID       string             `bson:"user_id" json:"user_id"`
	ActivityType string             `bson:"activity_type" json:"activity_type"` // viewing, editing, streaming
	ResourceID   string             `bson:"resource_id" json:"resource_id"`     // Room ID, document ID, etc.
	ResourceType string             `bson:"resource_type" json:"resource_type"` // room, document, page
	Metadata     map[string]interface{} `bson:"metadata" json:"metadata"`
	StartedAt    time.Time          `bson:"started_at" json:"started_at"`
	LastActivity time.Time          `bson:"last_activity" json:"last_activity"`
}

// RoomPresence represents presence information for a room
type RoomPresence struct {
	RoomID       string           `json:"room_id"`
	Participants []UserPresence   `json:"participants"`
	Count        int              `json:"count"`
	TypingUsers  []string         `json:"typing_users"`
}

// BulkPresenceRequest represents a request for multiple user statuses
type BulkPresenceRequest struct {
	UserIDs []string `json:"user_ids" binding:"required"`
}

// BulkPresenceResponse represents presence status for multiple users
type BulkPresenceResponse struct {
	Presences map[string]PresenceInfo `json:"presences"`
}

// PresenceInfo represents condensed presence information
type PresenceInfo struct {
	UserID        string         `json:"user_id"`
	Status        PresenceStatus `json:"status"`
	StatusMessage string         `json:"status_message,omitempty"`
	LastSeen      time.Time      `json:"last_seen"`
	CurrentRoom   string         `json:"current_room,omitempty"`
}

// PresenceUpdateRequest represents a presence update request
type PresenceUpdateRequest struct {
	UserID        string         `json:"user_id" binding:"required"`
	Status        PresenceStatus `json:"status" binding:"required"`
	StatusMessage string         `json:"status_message"`
	CurrentRoom   string         `json:"current_room"`
	Device        string         `json:"device"`
}

// TypingRequest represents a typing indicator request
type TypingRequest struct {
	RoomID   string `json:"room_id" binding:"required"`
	UserID   string `json:"user_id" binding:"required"`
	IsTyping bool   `json:"is_typing"`
}

// ActivityUpdateRequest represents an activity update
type ActivityUpdateRequest struct {
	UserID       string                 `json:"user_id" binding:"required"`
	ActivityType string                 `json:"activity_type" binding:"required"`
	ResourceID   string                 `json:"resource_id" binding:"required"`
	ResourceType string                 `json:"resource_type" binding:"required"`
	Metadata     map[string]interface{} `json:"metadata"`
}
