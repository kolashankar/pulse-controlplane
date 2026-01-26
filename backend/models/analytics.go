package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CustomMetric represents a user-defined metric
type CustomMetric struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID   primitive.ObjectID `bson:"project_id" json:"project_id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	MetricType  string             `bson:"metric_type" json:"metric_type"` // counter, gauge, histogram
	Unit        string             `bson:"unit" json:"unit"`               // count, ms, bytes, etc.
	Aggregation string             `bson:"aggregation" json:"aggregation"` // sum, avg, min, max, count
	Query       map[string]interface{} `bson:"query" json:"query"` // MongoDB query for data
	IsActive    bool               `bson:"is_active" json:"is_active"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// MetricAlert represents an alert configuration
type MetricAlert struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID        primitive.ObjectID `bson:"project_id" json:"project_id"`
	MetricName       string             `bson:"metric_name" json:"metric_name"`
	AlertName        string             `bson:"alert_name" json:"alert_name"`
	Description      string             `bson:"description" json:"description"`
	Condition        string             `bson:"condition" json:"condition"` // gt, lt, gte, lte, eq
	Threshold        float64            `bson:"threshold" json:"threshold"`
	Duration         int                `bson:"duration" json:"duration"` // minutes to check
	Severity         string             `bson:"severity" json:"severity"` // low, medium, high, critical
	NotificationChannels []string       `bson:"notification_channels" json:"notification_channels"` // email, webhook, slack
	IsActive         bool               `bson:"is_active" json:"is_active"`
	LastTriggered    *time.Time         `bson:"last_triggered,omitempty" json:"last_triggered,omitempty"`
	TriggerCount     int                `bson:"trigger_count" json:"trigger_count"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

// AlertTrigger represents an alert trigger event
type AlertTrigger struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AlertID     primitive.ObjectID `bson:"alert_id" json:"alert_id"`
	ProjectID   primitive.ObjectID `bson:"project_id" json:"project_id"`
	MetricValue float64            `bson:"metric_value" json:"metric_value"`
	Threshold   float64            `bson:"threshold" json:"threshold"`
	Message     string             `bson:"message" json:"message"`
	Severity    string             `bson:"severity" json:"severity"`
	Status      string             `bson:"status" json:"status"` // triggered, resolved, acknowledged
	TriggeredAt time.Time          `bson:"triggered_at" json:"triggered_at"`
	ResolvedAt  *time.Time         `bson:"resolved_at,omitempty" json:"resolved_at,omitempty"`
}

// AnalyticsExport represents an analytics export request
type AnalyticsExport struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID  primitive.ObjectID `bson:"project_id" json:"project_id"`
	ExportType string             `bson:"export_type" json:"export_type"` // csv, json, excel
	DateFrom   time.Time          `bson:"date_from" json:"date_from"`
	DateTo     time.Time          `bson:"date_to" json:"date_to"`
	Metrics    []string           `bson:"metrics" json:"metrics"` // List of metrics to export
	Status     string             `bson:"status" json:"status"` // pending, processing, completed, failed
	FileURL    string             `bson:"file_url" json:"file_url,omitempty"`
	FileSize   int64              `bson:"file_size" json:"file_size,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	CompletedAt *time.Time        `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

// UsageForecast represents predicted usage
type UsageForecast struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID      primitive.ObjectID `bson:"project_id" json:"project_id"`
	MetricType     string             `bson:"metric_type" json:"metric_type"`
	ForecastDate   time.Time          `bson:"forecast_date" json:"forecast_date"`
	PredictedValue float64            `bson:"predicted_value" json:"predicted_value"`
	ConfidenceLow  float64            `bson:"confidence_low" json:"confidence_low"`
	ConfidenceHigh float64            `bson:"confidence_high" json:"confidence_high"`
	Model          string             `bson:"model" json:"model"` // linear, exponential, moving_average
	Accuracy       float64            `bson:"accuracy" json:"accuracy"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

// RealTimeMetric represents real-time metric data
type RealTimeMetric struct {
	MetricName  string    `json:"metric_name"`
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Trend       string    `json:"trend"` // up, down, stable
	ChangeRate  float64   `json:"change_rate"` // percentage change
}

// AnalyticsDashboard represents the real-time analytics dashboard data
type AnalyticsDashboard struct {
	ProjectID       string           `json:"project_id"`
	Timestamp       time.Time        `json:"timestamp"`
	Metrics         []RealTimeMetric `json:"metrics"`
	ActiveAlerts    int              `json:"active_alerts"`
	RecentTriggers  []AlertTrigger   `json:"recent_triggers"`
	TopEvents       []EventSummary   `json:"top_events"`
}

// EventSummary represents summary of events
type EventSummary struct {
	EventType string `json:"event_type"`
	Count     int    `json:"count"`
	LastOccurred time.Time `json:"last_occurred"`
}

// TableName implementations
func (CustomMetric) TableName() string {
	return "custom_metrics"
}

func (MetricAlert) TableName() string {
	return "metric_alerts"
}

func (AlertTrigger) TableName() string {
	return "alert_triggers"
}

func (AnalyticsExport) TableName() string {
	return "analytics_exports"
}

func (UsageForecast) TableName() string {
	return "usage_forecasts"
}
