package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ModerationAction represents the action to take
type ModerationAction string

const (
	ModerationActionAllow  ModerationAction = "allow"
	ModerationActionWarn   ModerationAction = "warn"
	ModerationActionFlag   ModerationAction = "flag"
	ModerationActionBlock  ModerationAction = "block"
	ModerationActionDelete ModerationAction = "delete"
)

// ModerationSeverity represents the severity level
type ModerationSeverity string

const (
	ModerationSeverityLow      ModerationSeverity = "low"
	ModerationSeverityMedium   ModerationSeverity = "medium"
	ModerationSeverityHigh     ModerationSeverity = "high"
	ModerationSeverityCritical ModerationSeverity = "critical"
)

// ContentType represents the type of content being moderated
type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
	ContentTypeVideo ContentType = "video"
	ContentTypeAudio ContentType = "audio"
)

// ModerationConfig represents project-specific moderation configuration
type ModerationConfig struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID          primitive.ObjectID `bson:"project_id" json:"project_id"`
	Enabled            bool               `bson:"enabled" json:"enabled"`
	AutoModeration     bool               `bson:"auto_moderation" json:"auto_moderation"`
	AIEnabled          bool               `bson:"ai_enabled" json:"ai_enabled"`
	ToxicityThreshold  float64            `bson:"toxicity_threshold" json:"toxicity_threshold"`     // 0.0 - 1.0
	ProfanityThreshold float64            `bson:"profanity_threshold" json:"profanity_threshold"`   // 0.0 - 1.0
	SpamThreshold      float64            `bson:"spam_threshold" json:"spam_threshold"`             // 0.0 - 1.0
	AutoAction         ModerationAction   `bson:"auto_action" json:"auto_action"`                   // Action to take automatically
	NotifyAdmins       bool               `bson:"notify_admins" json:"notify_admins"`               // Notify on flagged content
	WebhookURL         string             `bson:"webhook_url" json:"webhook_url"`                   // Webhook for moderation events
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
}

// ModerationRule represents a custom moderation rule
type ModerationRule struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID   primitive.ObjectID `bson:"project_id" json:"project_id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	RuleType    string             `bson:"rule_type" json:"rule_type"` // keyword, regex, pattern, ai
	Pattern     string             `bson:"pattern" json:"pattern"`     // The pattern to match
	Action      ModerationAction   `bson:"action" json:"action"`
	Severity    ModerationSeverity `bson:"severity" json:"severity"`
	Enabled     bool               `bson:"enabled" json:"enabled"`
	CaseSensitive bool             `bson:"case_sensitive" json:"case_sensitive"`
	WholeWord   bool               `bson:"whole_word" json:"whole_word"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// ContentAnalysis represents the result of AI content analysis
type ContentAnalysis struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	ProjectID       primitive.ObjectID     `bson:"project_id" json:"project_id"`
	ContentType     ContentType            `bson:"content_type" json:"content_type"`
	ContentID       string                 `bson:"content_id" json:"content_id"` // External content ID
	Content         string                 `bson:"content" json:"content"`       // Text content or URL
	UserID          string                 `bson:"user_id" json:"user_id"`
	AnalysisMethod  string                 `bson:"analysis_method" json:"analysis_method"` // gemini, rule-based, hybrid
	Toxicity        float64                `bson:"toxicity" json:"toxicity"`
	Profanity       float64                `bson:"profanity" json:"profanity"`
	Spam            float64                `bson:"spam" json:"spam"`
	Hate            float64                `bson:"hate" json:"hate"`
	Sexual          float64                `bson:"sexual" json:"sexual"`
	Violence        float64                `bson:"violence" json:"violence"`
	Severity        ModerationSeverity     `bson:"severity" json:"severity"`
	IsFlagged       bool                   `bson:"is_flagged" json:"is_flagged"`
	RecommendedAction ModerationAction     `bson:"recommended_action" json:"recommended_action"`
	Reason          string                 `bson:"reason" json:"reason"`
	MatchedRules    []string               `bson:"matched_rules" json:"matched_rules"` // Rule IDs that matched
	Metadata        map[string]interface{} `bson:"metadata" json:"metadata"`
	ProcessingTime  int64                  `bson:"processing_time" json:"processing_time"` // milliseconds
	CreatedAt       time.Time              `bson:"created_at" json:"created_at"`
}

// ModerationLog represents a moderation action log
type ModerationLog struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	ProjectID       primitive.ObjectID     `bson:"project_id" json:"project_id"`
	AnalysisID      primitive.ObjectID     `bson:"analysis_id" json:"analysis_id"`
	ContentType     ContentType            `bson:"content_type" json:"content_type"`
	ContentID       string                 `bson:"content_id" json:"content_id"`
	UserID          string                 `bson:"user_id" json:"user_id"`
	Action          ModerationAction       `bson:"action" json:"action"`
	Severity        ModerationSeverity     `bson:"severity" json:"severity"`
	Reason          string                 `bson:"reason" json:"reason"`
	Automatic       bool                   `bson:"automatic" json:"automatic"`         // Was it automatic or manual
	ModeratorID     string                 `bson:"moderator_id" json:"moderator_id"`   // Admin/moderator who took action
	Metadata        map[string]interface{} `bson:"metadata" json:"metadata"`
	CreatedAt       time.Time              `bson:"created_at" json:"created_at"`
}

// Whitelist represents whitelisted content/users
type Whitelist struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID   primitive.ObjectID `bson:"project_id" json:"project_id"`
	Type        string             `bson:"type" json:"type"` // user, keyword, domain, ip
	Value       string             `bson:"value" json:"value"`
	Description string             `bson:"description" json:"description"`
	CreatedBy   string             `bson:"created_by" json:"created_by"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

// Blacklist represents blacklisted content/users
type Blacklist struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID   primitive.ObjectID `bson:"project_id" json:"project_id"`
	Type        string             `bson:"type" json:"type"` // user, keyword, domain, ip
	Value       string             `bson:"value" json:"value"`
	Reason      string             `bson:"reason" json:"reason"`
	Severity    ModerationSeverity `bson:"severity" json:"severity"`
	CreatedBy   string             `bson:"created_by" json:"created_by"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	ExpiresAt   *time.Time         `bson:"expires_at" json:"expires_at"` // Optional expiration
}

// ModerationStats represents moderation statistics
type ModerationStats struct {
	ProjectID         primitive.ObjectID `json:"project_id"`
	TotalAnalyzed     int64              `json:"total_analyzed"`
	Flagged           int64              `json:"flagged"`
	Blocked           int64              `json:"blocked"`
	Warnings          int64              `json:"warnings"`
	Deleted           int64              `json:"deleted"`
	AverageProcessing float64            `json:"average_processing_ms"`
	TopViolations     []ViolationCount   `json:"top_violations"`
	Period            string             `json:"period"` // daily, weekly, monthly
}

// ViolationCount represents a count of violations
type ViolationCount struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

// AnalyzeTextRequest represents a text analysis request
type AnalyzeTextRequest struct {
	Content   string                 `json:"content" binding:"required"`
	ContentID string                 `json:"content_id"`
	UserID    string                 `json:"user_id" binding:"required"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// AnalyzeImageRequest represents an image analysis request
type AnalyzeImageRequest struct {
	ImageURL  string                 `json:"image_url" binding:"required"`
	ContentID string                 `json:"content_id"`
	UserID    string                 `json:"user_id" binding:"required"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// CreateRuleRequest represents a moderation rule creation request
type CreateRuleRequest struct {
	Name          string             `json:"name" binding:"required"`
	Description   string             `json:"description"`
	RuleType      string             `json:"rule_type" binding:"required"`
	Pattern       string             `json:"pattern" binding:"required"`
	Action        ModerationAction   `json:"action" binding:"required"`
	Severity      ModerationSeverity `json:"severity" binding:"required"`
	CaseSensitive bool               `json:"case_sensitive"`
	WholeWord     bool               `json:"whole_word"`
}

// WhitelistRequest represents a whitelist request
type WhitelistRequest struct {
	Type        string `json:"type" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Description string `json:"description"`
	CreatedBy   string `json:"created_by" binding:"required"`
}

// BlacklistRequest represents a blacklist request
type BlacklistRequest struct {
	Type      string             `json:"type" binding:"required"`
	Value     string             `json:"value" binding:"required"`
	Reason    string             `json:"reason" binding:"required"`
	Severity  ModerationSeverity `json:"severity" binding:"required"`
	CreatedBy string             `json:"created_by" binding:"required"`
	ExpiresAt *time.Time         `json:"expires_at"`
}
