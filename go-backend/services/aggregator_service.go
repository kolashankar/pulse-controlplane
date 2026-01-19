package services

import (
	"context"
	"fmt"
	"time"

	"pulse-control-plane/models"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AggregatorService handles usage metrics aggregation
type AggregatorService struct {
	db *mongo.Database
}

// NewAggregatorService creates a new aggregator service
func NewAggregatorService(db *mongo.Database) *AggregatorService {
	return &AggregatorService{db: db}
}

// AggregateHourlyUsage aggregates usage metrics for the last hour
func (s *AggregatorService) AggregateHourlyUsage(ctx context.Context) error {
	log.Info().Msg("Starting hourly usage aggregation")

	// Calculate time range (last hour)
	now := time.Now()
	periodEnd := now.Truncate(time.Hour)
	periodStart := periodEnd.Add(-time.Hour)

	// Get all projects
	projectCollection := s.db.Collection(models.Project{}.TableName())
	projectCursor, err := projectCollection.Find(ctx, bson.M{"is_deleted": false})
	if err != nil {
		return fmt.Errorf("failed to find projects: %w", err)
	}
	defer projectCursor.Close(ctx)

	var projects []models.Project
	if err := projectCursor.All(ctx, &projects); err != nil {
		return fmt.Errorf("failed to decode projects: %w", err)
	}

	// Aggregate usage for each project
	for _, project := range projects {
		if err := s.aggregateForProject(ctx, project.ID, periodStart, periodEnd, models.PeriodHourly); err != nil {
			log.Error().Err(err).Str("project_id", project.ID.Hex()).Msg("Failed to aggregate usage for project")
			continue
		}
	}

	log.Info().Msgf("Completed hourly aggregation for %d projects", len(projects))
	return nil
}

// AggregateDailyUsage aggregates usage metrics for the last day
func (s *AggregatorService) AggregateDailyUsage(ctx context.Context) error {
	log.Info().Msg("Starting daily usage aggregation")

	// Calculate time range (yesterday)
	now := time.Now()
	periodEnd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	periodStart := periodEnd.Add(-24 * time.Hour)

	// Get all projects
	projectCollection := s.db.Collection(models.Project{}.TableName())
	projectCursor, err := projectCollection.Find(ctx, bson.M{"is_deleted": false})
	if err != nil {
		return fmt.Errorf("failed to find projects: %w", err)
	}
	defer projectCursor.Close(ctx)

	var projects []models.Project
	if err := projectCursor.All(ctx, &projects); err != nil {
		return fmt.Errorf("failed to decode projects: %w", err)
	}

	// Aggregate usage for each project
	for _, project := range projects {
		if err := s.aggregateForProject(ctx, project.ID, periodStart, periodEnd, models.PeriodDaily); err != nil {
			log.Error().Err(err).Str("project_id", project.ID.Hex()).Msg("Failed to aggregate usage for project")
			continue
		}
	}

	log.Info().Msgf("Completed daily aggregation for %d projects", len(projects))
	return nil
}

// AggregateMonthlyUsage aggregates usage metrics for the last month
func (s *AggregatorService) AggregateMonthlyUsage(ctx context.Context) error {
	log.Info().Msg("Starting monthly usage aggregation")

	// Calculate time range (last month)
	now := time.Now()
	periodEnd := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	periodStart := periodEnd.AddDate(0, -1, 0)

	// Get all projects
	projectCollection := s.db.Collection(models.Project{}.TableName())
	projectCursor, err := projectCollection.Find(ctx, bson.M{"is_deleted": false})
	if err != nil {
		return fmt.Errorf("failed to find projects: %w", err)
	}
	defer projectCursor.Close(ctx)

	var projects []models.Project
	if err := projectCursor.All(ctx, &projects); err != nil {
		return fmt.Errorf("failed to decode projects: %w", err)
	}

	// Aggregate usage for each project
	for _, project := range projects {
		if err := s.aggregateForProject(ctx, project.ID, periodStart, periodEnd, models.PeriodMonthly); err != nil {
			log.Error().Err(err).Str("project_id", project.ID.Hex()).Msg("Failed to aggregate usage for project")
			continue
		}
	}

	log.Info().Msgf("Completed monthly aggregation for %d projects", len(projects))
	return nil
}

// aggregateForProject aggregates usage for a single project
func (s *AggregatorService) aggregateForProject(ctx context.Context, projectID primitive.ObjectID, periodStart, periodEnd time.Time, periodType string) error {
	usageCollection := s.db.Collection(models.UsageMetric{}.TableName())
	aggregateCollection := s.db.Collection(models.UsageAggregate{}.TableName())

	// Check if aggregate already exists
	existingFilter := bson.M{
		"project_id":   projectID,
		"period_type":  periodType,
		"period_start": periodStart,
	}

	var existing models.UsageAggregate
	err := aggregateCollection.FindOne(ctx, existingFilter).Decode(&existing)
	if err == nil {
		// Already aggregated
		log.Debug().Str("project_id", projectID.Hex()).Str("period_type", periodType).Msg("Usage already aggregated")
		return nil
	}

	// Aggregate participant minutes
	participantFilter := bson.M{
		"project_id": projectID,
		"event_type": models.EventParticipantLeft,
		"timestamp": bson.M{
			"$gte": periodStart,
			"$lt":  periodEnd,
		},
	}
	participantPipeline := mongo.Pipeline{
		{{Key: "$match", Value: participantFilter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: "$value"}}},
		}}},
	}
	var participantMinutes float64
	participantCursor, err := usageCollection.Aggregate(ctx, participantPipeline)
	if err == nil {
		defer participantCursor.Close(ctx)
		var results []bson.M
		if err := participantCursor.All(ctx, &results); err == nil && len(results) > 0 {
			if total, ok := results[0]["total"].(float64); ok {
				participantMinutes = total
			}
		}
	}

	// Aggregate egress minutes
	egressFilter := bson.M{
		"project_id": projectID,
		"event_type": models.EventEgressEnded,
		"timestamp": bson.M{
			"$gte": periodStart,
			"$lt":  periodEnd,
		},
	}
	egressPipeline := mongo.Pipeline{
		{{Key: "$match", Value: egressFilter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: "$value"}}},
		}}},
	}
	var egressMinutes float64
	egressCursor, err := usageCollection.Aggregate(ctx, egressPipeline)
	if err == nil {
		defer egressCursor.Close(ctx)
		var results []bson.M
		if err := egressCursor.All(ctx, &results); err == nil && len(results) > 0 {
			if total, ok := results[0]["total"].(float64); ok {
				egressMinutes = total
			}
		}
	}

	// Aggregate storage (average for period)
	storageFilter := bson.M{
		"project_id": projectID,
		"event_type": models.EventStorageUsed,
		"timestamp": bson.M{
			"$gte": periodStart,
			"$lt":  periodEnd,
		},
	}
	storagePipeline := mongo.Pipeline{
		{{Key: "$match", Value: storageFilter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "average", Value: bson.D{{Key: "$avg", Value: "$value"}}},
		}}},
	}
	var storageGB float64
	storageCursor, err := usageCollection.Aggregate(ctx, storagePipeline)
	if err == nil {
		defer storageCursor.Close(ctx)
		var results []bson.M
		if err := storageCursor.All(ctx, &results); err == nil && len(results) > 0 {
			if avg, ok := results[0]["average"].(float64); ok {
				storageGB = avg
			}
		}
	}

	// Aggregate bandwidth
	bandwidthFilter := bson.M{
		"project_id": projectID,
		"event_type": models.EventBandwidthUsed,
		"timestamp": bson.M{
			"$gte": periodStart,
			"$lt":  periodEnd,
		},
	}
	bandwidthPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bandwidthFilter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: "$value"}}},
		}}},
	}
	var bandwidthGB float64
	bandwidthCursor, err := usageCollection.Aggregate(ctx, bandwidthPipeline)
	if err == nil {
		defer bandwidthCursor.Close(ctx)
		var results []bson.M
		if err := bandwidthCursor.All(ctx, &results); err == nil && len(results) > 0 {
			if total, ok := results[0]["total"].(float64); ok {
				bandwidthGB = total
			}
		}
	}

	// Count API requests
	apiFilter := bson.M{
		"project_id": projectID,
		"event_type": models.EventAPIRequest,
		"timestamp": bson.M{
			"$gte": periodStart,
			"$lt":  periodEnd,
		},
	}
	apiRequests, _ := usageCollection.CountDocuments(ctx, apiFilter)

	// Calculate cost (will be done by billing service)
	totalCost := 0.0

	// Create aggregate record
	aggregate := models.UsageAggregate{
		ProjectID:          projectID,
		PeriodType:         periodType,
		PeriodStart:        periodStart,
		PeriodEnd:          periodEnd,
		ParticipantMinutes: participantMinutes,
		EgressMinutes:      egressMinutes,
		StorageGB:          storageGB,
		BandwidthGB:        bandwidthGB,
		APIRequests:        apiRequests,
		TotalCost:          totalCost,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Insert aggregate
	_, err = aggregateCollection.InsertOne(ctx, aggregate)
	if err != nil {
		return fmt.Errorf("failed to insert aggregate: %w", err)
	}

	log.Info().Str("project_id", projectID.Hex()).Str("period_type", periodType).Msg("Usage aggregated successfully")
	return nil
}

// GetAggregatedUsage retrieves aggregated usage for a project
func (s *AggregatorService) GetAggregatedUsage(ctx context.Context, projectID primitive.ObjectID, periodType string, startDate, endDate time.Time) ([]models.UsageAggregate, error) {
	collection := s.db.Collection(models.UsageAggregate{}.TableName())

	filter := bson.M{
		"project_id":  projectID,
		"period_type": periodType,
		"period_start": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "period_start", Value: 1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find aggregates: %w", err)
	}
	defer cursor.Close(ctx)

	var aggregates []models.UsageAggregate
	if err := cursor.All(ctx, &aggregates); err != nil {
		return nil, fmt.Errorf("failed to decode aggregates: %w", err)
	}

	return aggregates, nil
}
