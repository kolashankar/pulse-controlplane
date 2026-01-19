package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StorageConfig holds cloud storage configuration
type StorageConfig struct {
	Provider        string `bson:"provider" json:"provider"` // s3, r2, gcs
	Bucket          string `bson:"bucket" json:"bucket"`
	AccessKeyID     string `bson:"access_key_id" json:"access_key_id,omitempty"`
	SecretAccessKey string `bson:"secret_access_key" json:"-"` // Never expose in JSON
	Region          string `bson:"region" json:"region"`
}

// Project represents a customer project/application
type Project struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID          primitive.ObjectID `bson:"org_id" json:"org_id"`
	Name           string             `bson:"name" json:"name"`
	PulseAPIKey    string             `bson:"pulse_api_key" json:"pulse_api_key"`
	PulseAPISecret string             `bson:"pulse_api_secret" json:"-"` // Never expose
	WebhookURL     string             `bson:"webhook_url" json:"webhook_url"`
	StorageConfig  StorageConfig      `bson:"storage_config" json:"storage_config"`
	LiveKitURL     string             `bson:"livekit_url" json:"livekit_url"`
	Region         string             `bson:"region" json:"region"` // us-east, eu-west, asia-south

	// Feature flags
	ChatEnabled        bool `bson:"chat_enabled" json:"chat_enabled"`
	VideoEnabled       bool `bson:"video_enabled" json:"video_enabled"`
	ActivityFeedEnabled bool `bson:"activity_feed_enabled" json:"activity_feed_enabled"`
	ModerationEnabled  bool `bson:"moderation_enabled" json:"moderation_enabled"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	IsDeleted bool      `bson:"is_deleted" json:"-"`
}

// ProjectCreate represents the input for creating a project
type ProjectCreate struct {
	Name       string        `json:"name" binding:"required,min=3,max=100"`
	WebhookURL string        `json:"webhook_url" binding:"omitempty,url"`
	Region     string        `json:"region" binding:"required,oneof=us-east us-west eu-west eu-central asia-south asia-east"`
	Storage    StorageConfig `json:"storage_config" binding:"omitempty"`
}

// ProjectUpdate represents the input for updating a project
type ProjectUpdate struct {
	Name       string        `json:"name" binding:"omitempty,min=3,max=100"`
	WebhookURL string        `json:"webhook_url" binding:"omitempty,url"`
	Storage    StorageConfig `json:"storage_config" binding:"omitempty"`
}

// ProjectResponse is the safe response excluding secrets
type ProjectResponse struct {
	ID                  string        `json:"id"`
	OrgID               string        `json:"org_id"`
	Name                string        `json:"name"`
	PulseAPIKey         string        `json:"pulse_api_key"`
	WebhookURL          string        `json:"webhook_url"`
	StorageConfig       StorageConfig `json:"storage_config"`
	LiveKitURL          string        `json:"livekit_url"`
	Region              string        `json:"region"`
	ChatEnabled         bool          `json:"chat_enabled"`
	VideoEnabled        bool          `json:"video_enabled"`
	ActivityFeedEnabled bool          `json:"activity_feed_enabled"`
	ModerationEnabled   bool          `json:"moderation_enabled"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

// TableName returns the collection name
func (Project) TableName() string {
	return "projects"
}

// ToResponse converts Project to ProjectResponse (safe for API)
func (p *Project) ToResponse() *ProjectResponse {
	return &ProjectResponse{
		ID:                  p.ID.Hex(),
		OrgID:               p.OrgID.Hex(),
		Name:                p.Name,
		PulseAPIKey:         p.PulseAPIKey,
		WebhookURL:          p.WebhookURL,
		StorageConfig:       p.StorageConfig,
		LiveKitURL:          p.LiveKitURL,
		Region:              p.Region,
		ChatEnabled:         p.ChatEnabled,
		VideoEnabled:        p.VideoEnabled,
		ActivityFeedEnabled: p.ActivityFeedEnabled,
		ModerationEnabled:   p.ModerationEnabled,
		CreatedAt:           p.CreatedAt,
		UpdatedAt:           p.UpdatedAt,
	}
}
