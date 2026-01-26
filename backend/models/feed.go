package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FeedType represents the type of feed
type FeedType string

const (
	FeedTypeUser       FeedType = "user"       // Personal user feed
	FeedTypeTimeline   FeedType = "timeline"   // Aggregated timeline
	FeedTypeActivity   FeedType = "activity"   // Activity-specific feed
	FeedTypeNotification FeedType = "notification" // Notification feed
)

// Feed represents a feed configuration
type Feed struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID   primitive.ObjectID `bson:"project_id" json:"project_id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	FeedType    FeedType           `bson:"feed_type" json:"feed_type"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Settings    FeedSettings       `bson:"settings" json:"settings"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// FeedSettings contains feed configuration
type FeedSettings struct {
	MaxItems        int    `bson:"max_items" json:"max_items"`                 // Max items in feed
	Ranking         string `bson:"ranking" json:"ranking"`                     // chronological, popularity
	Aggregation     bool   `bson:"aggregation" json:"aggregation"`             // Enable aggregation
	FanOutOnWrite   bool   `bson:"fan_out_on_write" json:"fan_out_on_write"`   // Fan-out strategy
}

// ActivityType represents types of activities
type ActivityType string

const (
	ActivityTypePost     ActivityType = "post"
	ActivityTypeLike     ActivityType = "like"
	ActivityTypeComment  ActivityType = "comment"
	ActivityTypeShare    ActivityType = "share"
	ActivityTypeFollow   ActivityType = "follow"
	ActivityTypeReaction ActivityType = "reaction"
)

// Activity represents an activity in the feed
type Activity struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID     `bson:"project_id" json:"project_id"`
	Actor     string                 `bson:"actor" json:"actor"`           // User who performed the activity
	Verb      ActivityType           `bson:"verb" json:"verb"`             // Type of activity
	Object    string                 `bson:"object" json:"object"`         // What the activity is about
	Target    string                 `bson:"target" json:"target"`         // Optional target
	ForeignID string                 `bson:"foreign_id" json:"foreign_id"` // External ID
	Time      time.Time              `bson:"time" json:"time"`
	Metadata  map[string]interface{} `bson:"metadata" json:"metadata"`
	CreatedAt time.Time              `bson:"created_at" json:"created_at"`
}

// FeedItem represents a denormalized item in a user's feed
type FeedItem struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FeedID     primitive.ObjectID `bson:"feed_id" json:"feed_id"`
	ActivityID primitive.ObjectID `bson:"activity_id" json:"activity_id"`
	ProjectID  primitive.ObjectID `bson:"project_id" json:"project_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Activity   Activity           `bson:"activity" json:"activity"`
	Score      float64            `bson:"score" json:"score"` // For ranking
	Seen       bool               `bson:"seen" json:"seen"`
	Read       bool               `bson:"read" json:"read"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

// Follow represents a follow relationship
type Follow struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID  primitive.ObjectID `bson:"project_id" json:"project_id"`
	Follower   string             `bson:"follower" json:"follower"`     // User ID who is following
	Following  string             `bson:"following" json:"following"`   // User ID being followed
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

// FollowStats represents follower/following statistics
type FollowStats struct {
	UserID         string `json:"user_id"`
	FollowersCount int64  `json:"followers_count"`
	FollowingCount int64  `json:"following_count"`
}

// AggregatedActivity represents grouped activities
type AggregatedActivity struct {
	ID          string     `json:"id"`
	Verb        ActivityType `json:"verb"`
	Actors      []string   `json:"actors"`
	ActivityCount int      `json:"activity_count"`
	FirstActivity Activity `json:"first_activity"`
	LatestActivity Activity `json:"latest_activity"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// FeedResponse represents the API response for feed queries
type FeedResponse struct {
	Items      []FeedItem `json:"items"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	HasMore    bool       `json:"has_more"`
	NextCursor string     `json:"next_cursor,omitempty"`
}

// ActivityRequest represents the request to create an activity
type ActivityRequest struct {
	Actor     string                 `json:"actor" binding:"required"`
	Verb      ActivityType           `json:"verb" binding:"required"`
	Object    string                 `json:"object" binding:"required"`
	Target    string                 `json:"target"`
	ForeignID string                 `json:"foreign_id"`
	Metadata  map[string]interface{} `json:"metadata"`
}
