package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RegionConfig represents a LiveKit server region
type RegionConfig struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Code             string             `bson:"code" json:"code"` // us-east, eu-west, etc.
	Name             string             `bson:"name" json:"name"`
	LiveKitURL       string             `bson:"livekit_url" json:"livekit_url"`
	LatencyEndpoint  string             `bson:"latency_endpoint" json:"latency_endpoint"`
	IsActive         bool               `bson:"is_active" json:"is_active"`
	Priority         int                `bson:"priority" json:"priority"` // Lower = higher priority
	MaxCapacity      int                `bson:"max_capacity" json:"max_capacity"`
	CurrentLoad      int                `bson:"current_load" json:"current_load"`
	HealthStatus     string             `bson:"health_status" json:"health_status"` // healthy, degraded, down
	LastHealthCheck  time.Time          `bson:"last_health_check" json:"last_health_check"`
	AverageLatency   float64            `bson:"average_latency" json:"average_latency"` // in ms
	FailoverRegions  []string           `bson:"failover_regions" json:"failover_regions"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

// RegionHealth represents real-time health status
type RegionHealth struct {
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	Status         string    `json:"status"`
	Latency        float64   `json:"latency"`
	Load           int       `json:"load"`
	Capacity       int       `json:"capacity"`
	LoadPercentage float64   `json:"load_percentage"`
	LastChecked    time.Time `json:"last_checked"`
}

// RegionLatency represents client-to-region latency measurement
type RegionLatency struct {
	RegionCode string  `json:"region_code"`
	Latency    float64 `json:"latency"` // in ms
	Timestamp  time.Time `json:"timestamp"`
}

// NearestRegionRequest represents request for finding nearest region
type NearestRegionRequest struct {
	ClientIP   string             `json:"client_ip"`
	Preference string             `json:"preference"` // Optional: user's preferred region
	Latencies  []RegionLatency    `json:"latencies"` // Client-measured latencies
}

// NearestRegionResponse represents the nearest region response
type NearestRegionResponse struct {
	PrimaryRegion   *RegionConfig `json:"primary_region"`
	FallbackRegions []*RegionConfig `json:"fallback_regions"`
	RecommendedURL  string        `json:"recommended_url"`
}

// TableName returns the collection name
func (RegionConfig) TableName() string {
	return "regions"
}
