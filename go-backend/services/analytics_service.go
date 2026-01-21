package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"pulse-control-plane/database"
	"pulse-control-plane/models"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AnalyticsService struct {
	db           *mongo.Database
	usageService *UsageService
}

func NewAnalyticsService(db *mongo.Database, usageService *UsageService) *AnalyticsService {
	return &AnalyticsService{
		db:           db,
		usageService: usageService,
	}
}

// CreateCustomMetric creates a new custom metric definition
func (s *AnalyticsService) CreateCustomMetric(ctx context.Context, metric *models.CustomMetric) error {
	collection := database.DB.Collection(models.CustomMetric{}.TableName())

	metric.ID = primitive.NewObjectID()
	metric.IsActive = true
	metric.CreatedAt = time.Now()
	metric.UpdatedAt = time.Now()

	_, err := collection.InsertOne(ctx, metric)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create custom metric")
		return err
	}

	log.Info().
		Str("metric_id", metric.ID.Hex()).
		Str("name", metric.Name).
		Msg("Custom metric created")

	return nil
}

// GetCustomMetrics returns all custom metrics for a project
func (s *AnalyticsService) GetCustomMetrics(ctx context.Context, projectID primitive.ObjectID) ([]models.CustomMetric, error) {
	collection := database.DB.Collection(models.CustomMetric{}.TableName())

	cursor, err := collection.Find(ctx, bson.M{"project_id": projectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var metrics []models.CustomMetric
	if err := cursor.All(ctx, &metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}

// CreateAlert creates a new metric alert
func (s *AnalyticsService) CreateAlert(ctx context.Context, alert *models.MetricAlert) error {
	collection := database.DB.Collection(models.MetricAlert{}.TableName())

	alert.ID = primitive.NewObjectID()
	alert.IsActive = true
	alert.TriggerCount = 0
	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()

	_, err := collection.InsertOne(ctx, alert)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create alert")
		return err
	}

	log.Info().
		Str("alert_id", alert.ID.Hex()).
		Str("metric", alert.MetricName).
		Msg("Alert created")

	return nil
}

// GetAlerts returns all alerts for a project
func (s *AnalyticsService) GetAlerts(ctx context.Context, projectID primitive.ObjectID) ([]models.MetricAlert, error) {
	collection := database.DB.Collection(models.MetricAlert{}.TableName())

	cursor, err := collection.Find(ctx, bson.M{"project_id": projectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var alerts []models.MetricAlert
	if err := cursor.All(ctx, &alerts); err != nil {
		return nil, err
	}

	return alerts, nil
}

// CheckAlerts checks all active alerts and triggers if thresholds are met
func (s *AnalyticsService) CheckAlerts(ctx context.Context, projectID primitive.ObjectID) ([]models.AlertTrigger, error) {
	alerts, err := s.GetAlerts(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var triggers []models.AlertTrigger

	for _, alert := range alerts {
		if !alert.IsActive {
			continue
		}

		// Get metric value
		value, err := s.GetMetricValue(ctx, projectID, alert.MetricName, alert.Duration)
		if err != nil {
			log.Error().Err(err).Str("metric", alert.MetricName).Msg("Failed to get metric value")
			continue
		}

		// Check condition
		triggered := false
		switch alert.Condition {
		case "gt":
			triggered = value > alert.Threshold
		case "gte":
			triggered = value >= alert.Threshold
		case "lt":
			triggered = value < alert.Threshold
		case "lte":
			triggered = value <= alert.Threshold
		case "eq":
			triggered = value == alert.Threshold
		}

		if triggered {
			trigger := models.AlertTrigger{
				ID:          primitive.NewObjectID(),
				AlertID:     alert.ID,
				ProjectID:   projectID,
				MetricValue: value,
				Threshold:   alert.Threshold,
				Message:     fmt.Sprintf("Alert '%s': %s is %.2f (threshold: %.2f)", alert.AlertName, alert.MetricName, value, alert.Threshold),
				Severity:    alert.Severity,
				Status:      "triggered",
				TriggeredAt: time.Now(),
			}

			// Save trigger
			collection := database.DB.Collection(models.AlertTrigger{}.TableName())
			_, err := collection.InsertOne(ctx, trigger)
			if err != nil {
				log.Error().Err(err).Msg("Failed to save alert trigger")
				continue
			}

			// Update alert last triggered time
			alertCollection := database.DB.Collection(models.MetricAlert{}.TableName())
			now := time.Now()
			_, err = alertCollection.UpdateOne(
				ctx,
				bson.M{"_id": alert.ID},
				bson.M{
					"$set": bson.M{
						"last_triggered": now,
						"updated_at":     now,
					},
					"$inc": bson.M{
						"trigger_count": 1,
					},
				},
			)
			if err != nil {
				log.Error().Err(err).Msg("Failed to update alert")
			}

			triggers = append(triggers, trigger)

			log.Warn().
				Str("alert", alert.AlertName).
				Str("metric", alert.MetricName).
				Float64("value", value).
				Float64("threshold", alert.Threshold).
				Msg("Alert triggered")
		}
	}

	return triggers, nil
}

// GetMetricValue calculates the metric value for alert checking
func (s *AnalyticsService) GetMetricValue(ctx context.Context, projectID primitive.ObjectID, metricName string, durationMinutes int) (float64, error) {
	// Get usage data for the duration
	startTime := time.Now().Add(-time.Duration(durationMinutes) * time.Minute)

	collection := database.DB.Collection(models.UsageMetric{}.TableName())

	// Map metric names to usage event types
	eventType := metricName
	switch metricName {
	case "participant_minutes":
		eventType = "participant_minutes"
	case "egress_minutes":
		eventType = "egress_minutes"
	case "storage_gb":
		eventType = "storage_usage"
	case "bandwidth_gb":
		eventType = "bandwidth_usage"
	case "api_requests":
		eventType = "api_request"
	}

	filter := bson.M{
		"project_id": projectID,
		"event_type": eventType,
		"timestamp":  bson.M{"$gte": startTime},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var metrics []models.UsageMetric
	if err := cursor.All(ctx, &metrics); err != nil {
		return 0, err
	}

	// Calculate sum
	total := 0.0
	for _, m := range metrics {
		total += m.Value
	}

	return total, nil
}

// GetRealTimeDashboard returns real-time analytics dashboard data
func (s *AnalyticsService) GetRealTimeDashboard(ctx context.Context, projectID primitive.ObjectID) (*models.AnalyticsDashboard, error) {
	// Get current metrics
	metrics, err := s.GetRealTimeMetrics(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Get active alerts
	activeAlerts, err := s.GetActiveAlertsCount(ctx, projectID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get active alerts count")
		activeAlerts = 0
	}

	// Get recent triggers
	recentTriggers, err := s.GetRecentTriggers(ctx, projectID, 10)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get recent triggers")
		recentTriggers = []models.AlertTrigger{}
	}

	// Get top events
	topEvents, err := s.GetTopEvents(ctx, projectID, 5)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get top events")
		topEvents = []models.EventSummary{}
	}

	dashboard := &models.AnalyticsDashboard{
		ProjectID:      projectID.Hex(),
		Timestamp:      time.Now(),
		Metrics:        metrics,
		ActiveAlerts:   activeAlerts,
		RecentTriggers: recentTriggers,
		TopEvents:      topEvents,
	}

	return dashboard, nil
}

// GetRealTimeMetrics returns current metric values with trends
func (s *AnalyticsService) GetRealTimeMetrics(ctx context.Context, projectID primitive.ObjectID) ([]models.RealTimeMetric, error) {
	now := time.Now()
	last15Min := now.Add(-15 * time.Minute)
	last30Min := now.Add(-30 * time.Minute)

	// Define metrics to track
	metricNames := []string{
		"participant_minutes",
		"egress_minutes",
		"storage_usage",
		"bandwidth_usage",
		"api_request",
	}

	var realTimeMetrics []models.RealTimeMetric

	for _, metricName := range metricNames {
		// Get current value (last 15 minutes)
		currentValue, _ := s.GetMetricValue(ctx, projectID, metricName, 15)

		// Get previous value (15-30 minutes ago)
		collection := database.DB.Collection(models.UsageMetric{}.TableName())
		filter := bson.M{
			"project_id": projectID,
			"event_type": metricName,
			"timestamp":  bson.M{"$gte": last30Min, "$lt": last15Min},
		}

		cursor, err := collection.Find(ctx, filter)
		if err != nil {
			continue
		}

		var prevMetrics []models.UsageMetric
		cursor.All(ctx, &prevMetrics)
		cursor.Close(ctx)

		prevValue := 0.0
		for _, m := range prevMetrics {
			prevValue += m.Value
		}

		// Calculate trend
		trend := "stable"
		changeRate := 0.0
		if prevValue > 0 {
			changeRate = ((currentValue - prevValue) / prevValue) * 100
			if changeRate > 5 {
				trend = "up"
			} else if changeRate < -5 {
				trend = "down"
			}
		} else if currentValue > 0 {
			trend = "up"
			changeRate = 100.0
		}

		// Get unit
		unit := "count"
		displayName := metricName
		switch metricName {
		case "participant_minutes":
			unit = "minutes"
			displayName = "Participant Minutes"
		case "egress_minutes":
			unit = "minutes"
			displayName = "Egress Minutes"
		case "storage_usage":
			unit = "GB"
			displayName = "Storage Usage"
		case "bandwidth_usage":
			unit = "GB"
			displayName = "Bandwidth Usage"
		case "api_request":
			unit = "requests"
			displayName = "API Requests"
		}

		realTimeMetrics = append(realTimeMetrics, models.RealTimeMetric{
			MetricName: displayName,
			Value:      math.Round(currentValue*100) / 100,
			Unit:       unit,
			Timestamp:  now,
			Trend:      trend,
			ChangeRate: math.Round(changeRate*100) / 100,
		})
	}

	return realTimeMetrics, nil
}

// GetActiveAlertsCount returns count of active alerts
func (s *AnalyticsService) GetActiveAlertsCount(ctx context.Context, projectID primitive.ObjectID) (int, error) {
	collection := database.DB.Collection(models.MetricAlert{}.TableName())

	count, err := collection.CountDocuments(ctx, bson.M{
		"project_id": projectID,
		"is_active":  true,
	})

	return int(count), err
}

// GetRecentTriggers returns recent alert triggers
func (s *AnalyticsService) GetRecentTriggers(ctx context.Context, projectID primitive.ObjectID, limit int) ([]models.AlertTrigger, error) {
	collection := database.DB.Collection(models.AlertTrigger{}.TableName())

	opts := options.Find().
		SetSort(bson.D{{Key: "triggered_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, bson.M{"project_id": projectID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var triggers []models.AlertTrigger
	if err := cursor.All(ctx, &triggers); err != nil {
		return nil, err
	}

	return triggers, nil
}

// GetTopEvents returns top event types by count
func (s *AnalyticsService) GetTopEvents(ctx context.Context, projectID primitive.ObjectID, limit int) ([]models.EventSummary, error) {
	collection := database.DB.Collection(models.UsageMetric{}.TableName())

	// Aggregate top events from last 24 hours
	last24Hours := time.Now().Add(-24 * time.Hour)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"project_id": projectID,
				"timestamp":  bson.M{"$gte": last24Hours},
			},
		},
		{
			"$group": bson.M{
				"_id":           "$event_type",
				"count":         bson.M{"$sum": 1},
				"last_occurred": bson.M{"$max": "$timestamp"},
			},
		},
		{
			"$sort": bson.M{"count": -1},
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.EventSummary
	for cursor.Next(ctx) {
		var result struct {
			EventType    string    `bson:"_id"`
			Count        int       `bson:"count"`
			LastOccurred time.Time `bson:"last_occurred"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		results = append(results, models.EventSummary{
			EventType:    result.EventType,
			Count:        result.Count,
			LastOccurred: result.LastOccurred,
		})
	}

	return results, nil
}

// ExportAnalytics exports analytics data in specified format
func (s *AnalyticsService) ExportAnalytics(ctx context.Context, projectID primitive.ObjectID, exportType string, dateFrom, dateTo time.Time, metricTypes []string) (*models.AnalyticsExport, error) {
	// Create export record
	export := &models.AnalyticsExport{
		ID:         primitive.NewObjectID(),
		ProjectID:  projectID,
		ExportType: exportType,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		Metrics:    metricTypes,
		Status:     "processing",
		CreatedAt:  time.Now(),
	}

	collection := database.DB.Collection(models.AnalyticsExport{}.TableName())
	_, err := collection.InsertOne(ctx, export)
	if err != nil {
		return nil, err
	}

	// Process export asynchronously
	go s.processExport(context.Background(), export)

	return export, nil
}

// processExport processes the export in background
func (s *AnalyticsService) processExport(ctx context.Context, export *models.AnalyticsExport) {
	// Get usage metrics
	collection := database.DB.Collection(models.UsageMetric{}.TableName())

	filter := bson.M{
		"project_id": export.ProjectID,
		"timestamp":  bson.M{"$gte": export.DateFrom, "$lte": export.DateTo},
	}

	if len(export.Metrics) > 0 {
		filter["event_type"] = bson.M{"$in": export.Metrics}
	}

	cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}))
	if err != nil {
		s.updateExportStatus(ctx, export.ID, "failed", "", 0)
		return
	}
	defer cursor.Close(ctx)

	var metrics []models.UsageMetric
	if err := cursor.All(ctx, &metrics); err != nil {
		s.updateExportStatus(ctx, export.ID, "failed", "", 0)
		return
	}

	// Generate file
	fileName := fmt.Sprintf("export_%s_%d.%s", export.ProjectID.Hex(), time.Now().Unix(), export.ExportType)
	filePath := filepath.Join("/tmp", fileName)

	var fileSize int64
	switch export.ExportType {
	case "csv":
		fileSize, err = s.generateCSV(filePath, metrics)
	case "json":
		fileSize, err = s.generateJSON(filePath, metrics)
	default:
		err = errors.New("unsupported export type")
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to generate export file")
		s.updateExportStatus(ctx, export.ID, "failed", "", 0)
		return
	}

	// In production, upload to S3/R2 and get URL
	fileURL := fmt.Sprintf("/downloads/%s", fileName)

	completedAt := time.Now()
	s.updateExportStatus(ctx, export.ID, "completed", fileURL, fileSize)

	log.Info().
		Str("export_id", export.ID.Hex()).
		Str("file", fileName).
		Int64("size", fileSize).
		Msg("Export completed")

	// Update export record
	exportCollection := database.DB.Collection(models.AnalyticsExport{}.TableName())
	exportCollection.UpdateOne(ctx, bson.M{"_id": export.ID}, bson.M{
		"$set": bson.M{
			"completed_at": completedAt,
		},
	})
}

// updateExportStatus updates export status
func (s *AnalyticsService) updateExportStatus(ctx context.Context, exportID primitive.ObjectID, status, fileURL string, fileSize int64) {
	collection := database.DB.Collection(models.AnalyticsExport{}.TableName())

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"file_url":  fileURL,
			"file_size": fileSize,
		},
	}

	collection.UpdateOne(ctx, bson.M{"_id": exportID}, update)
}

// generateCSV generates CSV export file
func (s *AnalyticsService) generateCSV(filePath string, metrics []models.UsageMetric) (int64, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Timestamp", "Event Type", "Value", "Metadata"}
	if err := writer.Write(header); err != nil {
		return 0, err
	}

	// Write data
	for _, metric := range metrics {
		metadataJSON, _ := json.Marshal(metric.Metadata)
		row := []string{
			metric.Timestamp.Format(time.RFC3339),
			metric.EventType,
			fmt.Sprintf("%.2f", metric.Value),
			string(metadataJSON),
		}
		if err := writer.Write(row); err != nil {
			return 0, err
		}
	}

	// Get file size
	info, err := file.Stat()
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

// generateJSON generates JSON export file
func (s *AnalyticsService) generateJSON(filePath string, metrics []models.UsageMetric) (int64, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(metrics); err != nil {
		return 0, err
	}

	// Get file size
	info, err := file.Stat()
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

// ForecastUsage generates usage forecast using linear regression
func (s *AnalyticsService) ForecastUsage(ctx context.Context, projectID primitive.ObjectID, metricType string, days int) ([]models.UsageForecast, error) {
	// Get historical data (last 30 days)
	startDate := time.Now().Add(-30 * 24 * time.Hour)

	collection := database.DB.Collection(models.UsageMetric{}.TableName())

	// Aggregate daily usage
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"project_id": projectID,
				"event_type": metricType,
				"timestamp":  bson.M{"$gte": startDate},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$dateToString": bson.M{
						"format": "%Y-%m-%d",
						"date":   "$timestamp",
					},
				},
				"total": bson.M{"$sum": "$value"},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var dailyData []struct {
		Date  string  `bson:"_id"`
		Total float64 `bson:"total"`
	}

	if err := cursor.All(ctx, &dailyData); err != nil {
		return nil, err
	}

	if len(dailyData) < 7 {
		return nil, errors.New("insufficient data for forecasting (need at least 7 days)")
	}

	// Simple linear regression
	n := float64(len(dailyData))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0

	for i, data := range dailyData {
		x := float64(i)
		y := data.Total

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope and intercept
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// Calculate standard error for confidence intervals
	variance := 0.0
	for i, data := range dailyData {
		predicted := slope*float64(i) + intercept
		variance += math.Pow(data.Total-predicted, 2)
	}
	stdError := math.Sqrt(variance / n)

	// Generate forecasts
	var forecasts []models.UsageForecast
	lastIndex := float64(len(dailyData))

	for i := 1; i <= days; i++ {
		x := lastIndex + float64(i)
		predicted := slope*x + intercept

		// Ensure non-negative predictions
		if predicted < 0 {
			predicted = 0
		}

		// Calculate confidence interval (95% confidence)
		margin := 1.96 * stdError
		confidenceLow := predicted - margin
		confidenceHigh := predicted + margin

		if confidenceLow < 0 {
			confidenceLow = 0
		}

		forecast := models.UsageForecast{
			ID:             primitive.NewObjectID(),
			ProjectID:      projectID,
			MetricType:     metricType,
			ForecastDate:   time.Now().Add(time.Duration(i) * 24 * time.Hour),
			PredictedValue: math.Round(predicted*100) / 100,
			ConfidenceLow:  math.Round(confidenceLow*100) / 100,
			ConfidenceHigh: math.Round(confidenceHigh*100) / 100,
			Model:          "linear_regression",
			Accuracy:       math.Round((1-(stdError/sumY))*10000) / 100, // Percentage
			CreatedAt:      time.Now(),
		}

		forecasts = append(forecasts, forecast)
	}

	// Save forecasts to database
	forecastCollection := database.DB.Collection(models.UsageForecast{}.TableName())
	var documents []interface{}
	for _, f := range forecasts {
		documents = append(documents, f)
	}

	_, err = forecastCollection.InsertMany(ctx, documents)
	if err != nil {
		log.Error().Err(err).Msg("Failed to save forecasts")
		// Continue anyway, return the forecasts
	}

	return forecasts, nil
}
