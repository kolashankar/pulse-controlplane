package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"pulse-control-plane/database"
	"pulse-control-plane/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AuditService handles audit logging operations
type AuditService struct {
	db   *mongo.Database
	coll *mongo.Collection
}

// NewAuditService creates a new audit service
func NewAuditService() *AuditService {
	db := database.GetDB()
	return &AuditService{
		db:   db,
		coll: db.Collection(models.AuditLog{}.TableName()),
	}
}

// LogAction logs an audit action
func (s *AuditService) LogAction(ctx context.Context, log *models.AuditLog) error {
	log.CreatedAt = time.Now()
	log.Timestamp = time.Now()

	if log.Status == "" {
		log.Status = "Success"
	}

	_, err := s.coll.InsertOne(ctx, log)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetAuditLogs retrieves audit logs with filtering
func (s *AuditService) GetAuditLogs(ctx context.Context, filter models.AuditLogFilter) ([]models.AuditLog, int64, error) {
	// Build MongoDB filter
	mongoFilter := bson.M{}

	if filter.OrgID != "" {
		orgID, err := primitive.ObjectIDFromHex(filter.OrgID)
		if err == nil {
			mongoFilter["org_id"] = orgID
		}
	}

	if filter.UserEmail != "" {
		mongoFilter["user_email"] = bson.M{"$regex": filter.UserEmail, "$options": "i"}
	}

	if filter.Action != "" {
		mongoFilter["action"] = filter.Action
	}

	if filter.Resource != "" {
		mongoFilter["resource"] = filter.Resource
	}

	if filter.ResourceID != "" {
		mongoFilter["resource_id"] = filter.ResourceID
	}

	if filter.Status != "" {
		mongoFilter["status"] = filter.Status
	}

	// Date range filter
	if !filter.StartDate.IsZero() || !filter.EndDate.IsZero() {
		dateFilter := bson.M{}
		if !filter.StartDate.IsZero() {
			dateFilter["$gte"] = filter.StartDate
		}
		if !filter.EndDate.IsZero() {
			dateFilter["$lte"] = filter.EndDate
		}
		mongoFilter["timestamp"] = dateFilter
	}

	// Pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 50
	}

	skip := (page - 1) * limit

	// Query options
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := s.coll.Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []models.AuditLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, fmt.Errorf("failed to decode audit logs: %w", err)
	}

	total, err := s.coll.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return logs, total, nil
}

// ExportAuditLogs exports audit logs to CSV format
func (s *AuditService) ExportAuditLogs(ctx context.Context, filter models.AuditLogFilter) (string, error) {
	// Get all logs matching the filter (no pagination for export)
	filter.Page = 1
	filter.Limit = 10000 // Max export limit

	logs, _, err := s.GetAuditLogs(ctx, filter)
	if err != nil {
		return "", err
	}

	// Create CSV
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	// Write header
	header := []string{"ID", "Timestamp", "User Email", "Action", "Resource", "Resource ID", "Resource Name", "IP Address", "Status"}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for _, log := range logs {
		row := []string{
			log.ID.Hex(),
			log.Timestamp.Format(time.RFC3339),
			log.UserEmail,
			log.Action,
			log.Resource,
			log.ResourceID,
			log.ResourceName,
			log.IPAddress,
			log.Status,
		}
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("CSV writer error: %w", err)
	}

	return builder.String(), nil
}

// GetAuditStats gets statistics about audit logs
func (s *AuditService) GetAuditStats(ctx context.Context, orgID primitive.ObjectID, days int) (map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	filter := bson.M{
		"org_id":    orgID,
		"timestamp": bson.M{"$gte": startDate},
	}

	// Count total actions
	total, err := s.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count total logs: %w", err)
	}

	// Count by action type
	actionPipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$action",
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"count": -1}}},
		{{Key: "$limit", Value: 10}},
	}

	actionCursor, err := s.coll.Aggregate(ctx, actionPipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate by action: %w", err)
	}
	defer actionCursor.Close(ctx)

	var actionStats []bson.M
	if err := actionCursor.All(ctx, &actionStats); err != nil {
		return nil, fmt.Errorf("failed to decode action stats: %w", err)
	}

	// Count by user
	userPipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$user_email",
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"count": -1}}},
		{{Key: "$limit", Value: 10}},
	}

	userCursor, err := s.coll.Aggregate(ctx, userPipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate by user: %w", err)
	}
	defer userCursor.Close(ctx)

	var userStats []bson.M
	if err := userCursor.All(ctx, &userStats); err != nil {
		return nil, fmt.Errorf("failed to decode user stats: %w", err)
	}

	// Count failures
	failureFilter := bson.M{
		"org_id":    orgID,
		"timestamp": bson.M{"$gte": startDate},
		"status":    "Failed",
	}
	failures, err := s.coll.CountDocuments(ctx, failureFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to count failures: %w", err)
	}

	return map[string]interface{}{
		"total_actions":     total,
		"failed_actions":    failures,
		"success_rate":      float64(total-failures) / float64(total) * 100,
		"top_actions":       actionStats,
		"top_users":         userStats,
		"period_days":       days,
		"start_date":        startDate,
		"end_date":          time.Now(),
	}, nil
}

// CleanupOldLogs removes audit logs older than the retention period (1 year)
func (s *AuditService) CleanupOldLogs(ctx context.Context, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		retentionDays = 365 // Default to 1 year
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	filter := bson.M{
		"created_at": bson.M{"$lt": cutoffDate},
	}

	result, err := s.coll.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old logs: %w", err)
	}

	return result.DeletedCount, nil
}

// GetRecentLogs gets the most recent audit logs for an organization
func (s *AuditService) GetRecentLogs(ctx context.Context, orgID primitive.ObjectID, limit int) ([]models.AuditLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	filter := bson.M{"org_id": orgID}
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := s.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent logs: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []models.AuditLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode logs: %w", err)
	}

	return logs, nil
}
