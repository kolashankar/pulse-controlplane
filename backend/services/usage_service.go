package services

import (
	"context"
	"fmt"
	"time"

	"pulse-control-plane/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UsageService handles usage metrics operations
type UsageService struct {
	db *mongo.Database
}

// NewUsageService creates a new usage service
func NewUsageService(db *mongo.Database) *UsageService {
	return &UsageService{db: db}
}

// TrackUsage records a usage event
func (s *UsageService) TrackUsage(ctx context.Context, projectID primitive.ObjectID, eventType string, value float64, metadata map[string]interface{}) error {
	collection := s.db.Collection(models.UsageMetric{}.TableName())

	usageMetric := models.UsageMetric{
		ProjectID: projectID,
		EventType: eventType,
		Value:     value,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}

	_, err := collection.InsertOne(ctx, usageMetric)
	if err != nil {
		return fmt.Errorf("failed to track usage: %w", err)
	}

	return nil
}

// TrackParticipantMinutes tracks participant minutes from webhook
func (s *UsageService) TrackParticipantMinutes(ctx context.Context, projectID primitive.ObjectID, roomName string, participantID string, durationMinutes float64) error {
	metadata := map[string]interface{}{
		"room_name":      roomName,
		"participant_id": participantID,
	}
	return s.TrackUsage(ctx, projectID, models.EventParticipantLeft, durationMinutes, metadata)
}

// TrackEgressMinutes tracks egress minutes
func (s *UsageService) TrackEgressMinutes(ctx context.Context, projectID primitive.ObjectID, egressID string, durationMinutes float64) error {
	metadata := map[string]interface{}{
		"egress_id": egressID,
	}
	return s.TrackUsage(ctx, projectID, models.EventEgressEnded, durationMinutes, metadata)
}

// TrackStorageUsage tracks storage usage in GB
func (s *UsageService) TrackStorageUsage(ctx context.Context, projectID primitive.ObjectID, sizeGB float64) error {
	metadata := map[string]interface{}{
		"timestamp": time.Now(),
	}
	return s.TrackUsage(ctx, projectID, models.EventStorageUsed, sizeGB, metadata)
}

// TrackBandwidthUsage tracks bandwidth usage in GB
func (s *UsageService) TrackBandwidthUsage(ctx context.Context, projectID primitive.ObjectID, sizeGB float64) error {
	metadata := map[string]interface{}{
		"timestamp": time.Now(),
	}
	return s.TrackUsage(ctx, projectID, models.EventBandwidthUsed, sizeGB, metadata)
}

// TrackAPIRequest tracks an API request
func (s *UsageService) TrackAPIRequest(ctx context.Context, projectID primitive.ObjectID, endpoint string) error {
	metadata := map[string]interface{}{
		"endpoint": endpoint,
	}
	return s.TrackUsage(ctx, projectID, models.EventAPIRequest, 1, metadata)
}

// GetUsageMetrics retrieves usage metrics for a project within a time range
func (s *UsageService) GetUsageMetrics(ctx context.Context, projectID primitive.ObjectID, startDate, endDate time.Time, page, limit int) ([]models.UsageMetric, int64, error) {
	collection := s.db.Collection(models.UsageMetric{}.TableName())

	filter := bson.M{
		"project_id": projectID,
		"timestamp": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	// Count total
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count usage metrics: %w", err)
	}

	// Find with pagination
	opts := options.Find()
	opts.SetSkip(int64((page - 1) * limit))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find usage metrics: %w", err)
	}
	defer cursor.Close(ctx)

	var metrics []models.UsageMetric
	if err := cursor.All(ctx, &metrics); err != nil {
		return nil, 0, fmt.Errorf("failed to decode usage metrics: %w", err)
	}

	return metrics, total, nil
}

// GetUsageSummary calculates aggregated usage for a project
func (s *UsageService) GetUsageSummary(ctx context.Context, projectID primitive.ObjectID, startDate, endDate time.Time) (*models.UsageSummary, error) {
	collection := s.db.Collection(models.UsageMetric{}.TableName())

	summary := &models.UsageSummary{
		ProjectID: projectID.Hex(),
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Aggregate participant minutes
	participantFilter := bson.M{
		"project_id": projectID,
		"event_type": models.EventParticipantLeft,
		"timestamp": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	participantPipeline := mongo.Pipeline{
		{{Key: "$match", Value: participantFilter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: "$value"}}},
		}}},
	}
	participantCursor, err := collection.Aggregate(ctx, participantPipeline)
	if err == nil {
		defer participantCursor.Close(ctx)
		var results []bson.M
		if err := participantCursor.All(ctx, &results); err == nil && len(results) > 0 {
			if total, ok := results[0]["total"].(float64); ok {
				summary.ParticipantMinutes = total
			}
		}
	}

	// Aggregate egress minutes
	egressFilter := bson.M{
		"project_id": projectID,
		"event_type": models.EventEgressEnded,
		"timestamp": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	egressPipeline := mongo.Pipeline{
		{{Key: "$match", Value: egressFilter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: "$value"}}},
		}}},
	}
	egressCursor, err := collection.Aggregate(ctx, egressPipeline)
	if err == nil {
		defer egressCursor.Close(ctx)
		var results []bson.M
		if err := egressCursor.All(ctx, &results); err == nil && len(results) > 0 {
			if total, ok := results[0]["total"].(float64); ok {
				summary.EgressMinutes = total
			}
		}
	}

	// Aggregate storage usage (get latest value)
	storageFilter := bson.M{
		"project_id": projectID,
		"event_type": models.EventStorageUsed,
	}
	storagePipeline := mongo.Pipeline{
		{{Key: "$match", Value: storageFilter}},
		{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
		{{Key: "$limit", Value: 1}},
	}
	storageCursor, err := collection.Aggregate(ctx, storagePipeline)
	if err == nil {
		defer storageCursor.Close(ctx)
		var results []models.UsageMetric
		if err := storageCursor.All(ctx, &results); err == nil && len(results) > 0 {
			summary.StorageGB = results[0].Value
		}
	}

	// Aggregate bandwidth usage
	bandwidthFilter := bson.M{
		"project_id": projectID,
		"event_type": models.EventBandwidthUsed,
		"timestamp": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	bandwidthPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bandwidthFilter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: "$value"}}},
		}}},
	}
	bandwidthCursor, err := collection.Aggregate(ctx, bandwidthPipeline)
	if err == nil {
		defer bandwidthCursor.Close(ctx)
		var results []bson.M
		if err := bandwidthCursor.All(ctx, &results); err == nil && len(results) > 0 {
			if total, ok := results[0]["total"].(float64); ok {
				summary.BandwidthGB = total
			}
		}
	}

	// Count API requests
	apiFilter := bson.M{
		"project_id": projectID,
		"event_type": models.EventAPIRequest,
		"timestamp": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	apiCount, err := collection.CountDocuments(ctx, apiFilter)
	if err == nil {
		summary.APIRequests = apiCount
	}

	// Calculate total cost (will be done by billing service)
	summary.TotalCost = 0

	return summary, nil
}

// CheckLimits checks if usage is approaching or exceeding plan limits
func (s *UsageService) CheckLimits(ctx context.Context, projectID primitive.ObjectID, plan string, startDate, endDate time.Time) ([]models.UsageAlert, error) {
	// Get usage summary
	summary, err := s.GetUsageSummary(ctx, projectID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage summary: %w", err)
	}

	// Get plan limits
	var limits models.PlanLimits
	switch plan {
	case "Free":
		limits = models.FreePlanLimits
	case "Pro":
		limits = models.ProPlanLimits
	case "Enterprise":
		limits = models.EnterprisePlanLimits
	default:
		limits = models.FreePlanLimits
	}

	var alerts []models.UsageAlert

	// Check participant minutes
	if limits.MaxParticipantMinutes > 0 {
		percentage := (summary.ParticipantMinutes / limits.MaxParticipantMinutes) * 100
		if percentage >= float64(limits.AlertThresholdPercentage) {
			alert := models.UsageAlert{
				ProjectID:    projectID,
				MetricType:   "participant_minutes",
				CurrentUsage: summary.ParticipantMinutes,
				Limit:        limits.MaxParticipantMinutes,
				Percentage:   percentage,
				Severity:     s.getSeverity(percentage),
				Message:      fmt.Sprintf("Participant minutes at %.1f%% of limit", percentage),
				Notified:     false,
				CreatedAt:    time.Now(),
			}
			alerts = append(alerts, alert)
		}
	}

	// Check egress minutes
	if limits.MaxEgressMinutes > 0 {
		percentage := (summary.EgressMinutes / limits.MaxEgressMinutes) * 100
		if percentage >= float64(limits.AlertThresholdPercentage) {
			alert := models.UsageAlert{
				ProjectID:    projectID,
				MetricType:   "egress_minutes",
				CurrentUsage: summary.EgressMinutes,
				Limit:        limits.MaxEgressMinutes,
				Percentage:   percentage,
				Severity:     s.getSeverity(percentage),
				Message:      fmt.Sprintf("Egress minutes at %.1f%% of limit", percentage),
				Notified:     false,
				CreatedAt:    time.Now(),
			}
			alerts = append(alerts, alert)
		}
	}

	// Check storage
	if limits.MaxStorageGB > 0 {
		percentage := (summary.StorageGB / limits.MaxStorageGB) * 100
		if percentage >= float64(limits.AlertThresholdPercentage) {
			alert := models.UsageAlert{
				ProjectID:    projectID,
				MetricType:   "storage_gb",
				CurrentUsage: summary.StorageGB,
				Limit:        limits.MaxStorageGB,
				Percentage:   percentage,
				Severity:     s.getSeverity(percentage),
				Message:      fmt.Sprintf("Storage at %.1f%% of limit", percentage),
				Notified:     false,
				CreatedAt:    time.Now(),
			}
			alerts = append(alerts, alert)
		}
	}

	// Check bandwidth
	if limits.MaxBandwidthGB > 0 {
		percentage := (summary.BandwidthGB / limits.MaxBandwidthGB) * 100
		if percentage >= float64(limits.AlertThresholdPercentage) {
			alert := models.UsageAlert{
				ProjectID:    projectID,
				MetricType:   "bandwidth_gb",
				CurrentUsage: summary.BandwidthGB,
				Limit:        limits.MaxBandwidthGB,
				Percentage:   percentage,
				Severity:     s.getSeverity(percentage),
				Message:      fmt.Sprintf("Bandwidth at %.1f%% of limit", percentage),
				Notified:     false,
				CreatedAt:    time.Now(),
			}
			alerts = append(alerts, alert)
		}
	}

	// Check API requests
	if limits.MaxAPIRequests > 0 {
		percentage := (float64(summary.APIRequests) / float64(limits.MaxAPIRequests)) * 100
		if percentage >= float64(limits.AlertThresholdPercentage) {
			alert := models.UsageAlert{
				ProjectID:    projectID,
				MetricType:   "api_requests",
				CurrentUsage: float64(summary.APIRequests),
				Limit:        float64(limits.MaxAPIRequests),
				Percentage:   percentage,
				Severity:     s.getSeverity(percentage),
				Message:      fmt.Sprintf("API requests at %.1f%% of limit", percentage),
				Notified:     false,
				CreatedAt:    time.Now(),
			}
			alerts = append(alerts, alert)
		}
	}

	// Store alerts in database
	if len(alerts) > 0 {
		alertCollection := s.db.Collection(models.UsageAlert{}.TableName())
		for _, alert := range alerts {
			_, _ = alertCollection.InsertOne(ctx, alert)
		}
	}

	return alerts, nil
}

func (s *UsageService) getSeverity(percentage float64) string {
	if percentage >= 95 {
		return models.SeverityCritical
	}
	return models.SeverityWarning
}

// GetAlerts retrieves active alerts for a project
func (s *UsageService) GetAlerts(ctx context.Context, projectID primitive.ObjectID) ([]models.UsageAlert, error) {
	collection := s.db.Collection(models.UsageAlert{}.TableName())

	filter := bson.M{
		"project_id": projectID,
		"created_at": bson.M{
			"$gte": time.Now().Add(-24 * time.Hour), // Last 24 hours
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find alerts: %w", err)
	}
	defer cursor.Close(ctx)

	var alerts []models.UsageAlert
	if err := cursor.All(ctx, &alerts); err != nil {
		return nil, fmt.Errorf("failed to decode alerts: %w", err)
	}

	return alerts, nil
}
