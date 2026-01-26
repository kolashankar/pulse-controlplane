package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"time"

	"pulse-control-plane/database"
	"pulse-control-plane/models"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RegionService struct{}

func NewRegionService() *RegionService {
	return &RegionService{}
}

// InitializeDefaultRegions creates default region configurations
func (s *RegionService) InitializeDefaultRegions(ctx context.Context) error {
	collection := database.Database.Collection(models.RegionConfig{}.TableName())

	// Check if regions already exist
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return err
	}

	if count > 0 {
		log.Info().Msg("Regions already initialized")
		return nil
	}

	// Default regions configuration
	defaultRegions := []models.RegionConfig{
		{
			Code:            "us-east",
			Name:            "US East (Virginia)",
			LiveKitURL:      "wss://us-east.livekit.pulse.io",
			LatencyEndpoint: "https://us-east-ping.pulse.io/health",
			IsActive:        true,
			Priority:        1,
			MaxCapacity:     10000,
			CurrentLoad:     0,
			HealthStatus:    "healthy",
			AverageLatency:  50.0,
			FailoverRegions: []string{"us-west", "eu-west"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			Code:            "us-west",
			Name:            "US West (California)",
			LiveKitURL:      "wss://us-west.livekit.pulse.io",
			LatencyEndpoint: "https://us-west-ping.pulse.io/health",
			IsActive:        true,
			Priority:        2,
			MaxCapacity:     10000,
			CurrentLoad:     0,
			HealthStatus:    "healthy",
			AverageLatency:  55.0,
			FailoverRegions: []string{"us-east", "asia-east"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			Code:            "eu-west",
			Name:            "Europe West (Ireland)",
			LiveKitURL:      "wss://eu-west.livekit.pulse.io",
			LatencyEndpoint: "https://eu-west-ping.pulse.io/health",
			IsActive:        true,
			Priority:        1,
			MaxCapacity:     10000,
			CurrentLoad:     0,
			HealthStatus:    "healthy",
			AverageLatency:  60.0,
			FailoverRegions: []string{"eu-central", "us-east"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			Code:            "eu-central",
			Name:            "Europe Central (Frankfurt)",
			LiveKitURL:      "wss://eu-central.livekit.pulse.io",
			LatencyEndpoint: "https://eu-central-ping.pulse.io/health",
			IsActive:        true,
			Priority:        2,
			MaxCapacity:     10000,
			CurrentLoad:     0,
			HealthStatus:    "healthy",
			AverageLatency:  65.0,
			FailoverRegions: []string{"eu-west", "asia-south"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			Code:            "asia-south",
			Name:            "Asia South (Mumbai)",
			LiveKitURL:      "wss://asia-south.livekit.pulse.io",
			LatencyEndpoint: "https://asia-south-ping.pulse.io/health",
			IsActive:        true,
			Priority:        1,
			MaxCapacity:     10000,
			CurrentLoad:     0,
			HealthStatus:    "healthy",
			AverageLatency:  70.0,
			FailoverRegions: []string{"asia-east", "eu-central"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			Code:            "asia-east",
			Name:            "Asia East (Tokyo)",
			LiveKitURL:      "wss://asia-east.livekit.pulse.io",
			LatencyEndpoint: "https://asia-east-ping.pulse.io/health",
			IsActive:        true,
			Priority:        2,
			MaxCapacity:     10000,
			CurrentLoad:     0,
			HealthStatus:    "healthy",
			AverageLatency:  75.0,
			FailoverRegions: []string{"asia-south", "us-west"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	// Insert default regions
	var documents []interface{}
	for _, region := range defaultRegions {
		documents = append(documents, region)
	}

	_, err = collection.InsertMany(ctx, documents)
	if err != nil {
		return err
	}

	log.Info().Int("count", len(defaultRegions)).Msg("Initialized default regions")
	return nil
}

// GetAllRegions returns all region configurations
func (s *RegionService) GetAllRegions(ctx context.Context) ([]models.RegionConfig, error) {
	collection := database.Database.Collection(models.RegionConfig{}.TableName())

	cursor, err := collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "priority", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var regions []models.RegionConfig
	if err := cursor.All(ctx, &regions); err != nil {
		return nil, err
	}

	return regions, nil
}

// GetRegionByCode returns a specific region by code
func (s *RegionService) GetRegionByCode(ctx context.Context, code string) (*models.RegionConfig, error) {
	collection := database.Database.Collection(models.RegionConfig{}.TableName())

	var region models.RegionConfig
	err := collection.FindOne(ctx, bson.M{"code": code}).Decode(&region)
	if err != nil {
		return nil, err
	}

	return &region, nil
}

// GetHealthyRegions returns all healthy and active regions
func (s *RegionService) GetHealthyRegions(ctx context.Context) ([]models.RegionConfig, error) {
	collection := database.Database.Collection(models.RegionConfig{}.TableName())

	filter := bson.M{
		"is_active":     true,
		"health_status": "healthy",
	}

	cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "priority", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var regions []models.RegionConfig
	if err := cursor.All(ctx, &regions); err != nil {
		return nil, err
	}

	return regions, nil
}

// CheckRegionHealth performs health check on a region
func (s *RegionService) CheckRegionHealth(ctx context.Context, regionCode string) (*models.RegionHealth, error) {
	region, err := s.GetRegionByCode(ctx, regionCode)
	if err != nil {
		return nil, err
	}

	// Perform actual health check
	latency, status := s.performHealthCheck(region.LatencyEndpoint)

	// Update region health in database
	collection := database.Database.Collection(models.RegionConfig{}.TableName())
	update := bson.M{
		"$set": bson.M{
			"health_status":     status,
			"average_latency":   latency,
			"last_health_check": time.Now(),
			"updated_at":        time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"code": regionCode}, update)
	if err != nil {
		log.Error().Err(err).Str("region", regionCode).Msg("Failed to update region health")
	}

	loadPercentage := 0.0
	if region.MaxCapacity > 0 {
		loadPercentage = (float64(region.CurrentLoad) / float64(region.MaxCapacity)) * 100
	}

	return &models.RegionHealth{
		Code:           region.Code,
		Name:           region.Name,
		Status:         status,
		Latency:        latency,
		Load:           region.CurrentLoad,
		Capacity:       region.MaxCapacity,
		LoadPercentage: loadPercentage,
		LastChecked:    time.Now(),
	}, nil
}

// performHealthCheck performs actual HTTP health check
func (s *RegionService) performHealthCheck(endpoint string) (float64, string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	start := time.Now()
	resp, err := client.Get(endpoint)
	latency := float64(time.Since(start).Milliseconds())

	if err != nil {
		log.Warn().Err(err).Str("endpoint", endpoint).Msg("Health check failed")
		return 999.0, "down"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		if latency > 500 {
			return latency, "degraded"
		}
		return latency, "healthy"
	}

	return latency, "degraded"
}

// GetAllRegionHealth returns health status for all regions
func (s *RegionService) GetAllRegionHealth(ctx context.Context) ([]models.RegionHealth, error) {
	regions, err := s.GetAllRegions(ctx)
	if err != nil {
		return nil, err
	}

	var healthStatuses []models.RegionHealth
	for _, region := range regions {
		health, err := s.CheckRegionHealth(ctx, region.Code)
		if err != nil {
			log.Error().Err(err).Str("region", region.Code).Msg("Failed to check region health")
			// Add degraded status if health check fails
			loadPercentage := 0.0
			if region.MaxCapacity > 0 {
				loadPercentage = (float64(region.CurrentLoad) / float64(region.MaxCapacity)) * 100
			}
			healthStatuses = append(healthStatuses, models.RegionHealth{
				Code:           region.Code,
				Name:           region.Name,
				Status:         "degraded",
				Latency:        999.0,
				Load:           region.CurrentLoad,
				Capacity:       region.MaxCapacity,
				LoadPercentage: loadPercentage,
				LastChecked:    time.Now(),
			})
			continue
		}
		healthStatuses = append(healthStatuses, *health)
	}

	return healthStatuses, nil
}

// FindNearestRegion finds the best region based on client location and latencies
func (s *RegionService) FindNearestRegion(ctx context.Context, req *models.NearestRegionRequest) (*models.NearestRegionResponse, error) {
	// Get all healthy regions
	healthyRegions, err := s.GetHealthyRegions(ctx)
	if err != nil {
		return nil, err
	}

	if len(healthyRegions) == 0 {
		return nil, errors.New("no healthy regions available")
	}

	// If client provided latency measurements, use them
	if len(req.Latencies) > 0 {
		return s.selectByClientLatency(ctx, healthyRegions, req.Latencies)
	}

	// If user has preference and that region is healthy, use it
	if req.Preference != "" {
		for _, region := range healthyRegions {
			if region.Code == req.Preference {
				return s.buildRegionResponse(ctx, &region)
			}
		}
	}

	// Default: select region with lowest load and priority
	sort.Slice(healthyRegions, func(i, j int) bool {
		if healthyRegions[i].Priority == healthyRegions[j].Priority {
			return healthyRegions[i].CurrentLoad < healthyRegions[j].CurrentLoad
		}
		return healthyRegions[i].Priority < healthyRegions[j].Priority
	})

	return s.buildRegionResponse(ctx, &healthyRegions[0])
}

// selectByClientLatency selects region based on client-measured latencies
func (s *RegionService) selectByClientLatency(ctx context.Context, regions []models.RegionConfig, latencies []models.RegionLatency) (*models.NearestRegionResponse, error) {
	// Create latency map
	latencyMap := make(map[string]float64)
	for _, l := range latencies {
		latencyMap[l.RegionCode] = l.Latency
	}

	// Score each region (lower is better)
	type regionScore struct {
		region *models.RegionConfig
		score  float64
	}

	var scores []regionScore
	for i := range regions {
		region := &regions[i]
		latency, hasLatency := latencyMap[region.Code]
		if !hasLatency {
			latency = region.AverageLatency // Use default if not measured
		}

		// Calculate score: latency + load factor + priority factor
		loadFactor := 0.0
		if region.MaxCapacity > 0 {
			loadFactor = (float64(region.CurrentLoad) / float64(region.MaxCapacity)) * 100
		}

		score := latency + (loadFactor * 2) + (float64(region.Priority) * 10)
		scores = append(scores, regionScore{region: region, score: score})
	}

	// Sort by score
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score < scores[j].score
	})

	if len(scores) == 0 {
		return nil, errors.New("no suitable region found")
	}

	return s.buildRegionResponse(ctx, scores[0].region)
}

// buildRegionResponse builds response with primary and fallback regions
func (s *RegionService) buildRegionResponse(ctx context.Context, primary *models.RegionConfig) (*models.NearestRegionResponse, error) {
	// Get fallback regions
	var fallbacks []*models.RegionConfig
	for _, code := range primary.FailoverRegions {
		region, err := s.GetRegionByCode(ctx, code)
		if err != nil {
			log.Warn().Err(err).Str("region", code).Msg("Failover region not found")
			continue
		}
		if region.IsActive && region.HealthStatus == "healthy" {
			fallbacks = append(fallbacks, region)
		}
	}

	return &models.NearestRegionResponse{
		PrimaryRegion:   primary,
		FallbackRegions: fallbacks,
		RecommendedURL:  primary.LiveKitURL,
	}, nil
}

// UpdateRegionLoad updates the current load for a region
func (s *RegionService) UpdateRegionLoad(ctx context.Context, regionCode string, loadDelta int) error {
	collection := database.Database.Collection(models.RegionConfig{}.TableName())

	update := bson.M{
		"$inc": bson.M{
			"current_load": loadDelta,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"code": regionCode}, update)
	return err
}

// GetRegionStats returns aggregated statistics for regions
func (s *RegionService) GetRegionStats(ctx context.Context) (map[string]interface{}, error) {
	regions, err := s.GetAllRegions(ctx)
	if err != nil {
		return nil, err
	}

	totalCapacity := 0
	totalLoad := 0
	healthyCount := 0
	degradedCount := 0
	downCount := 0

	for _, region := range regions {
		if !region.IsActive {
			continue
		}

		totalCapacity += region.MaxCapacity
		totalLoad += region.CurrentLoad

		switch region.HealthStatus {
		case "healthy":
			healthyCount++
		case "degraded":
			degradedCount++
		case "down":
			downCount++
		}
	}

	overallHealth := "healthy"
	if downCount > 0 {
		overallHealth = "degraded"
	}
	if healthyCount == 0 {
		overallHealth = "critical"
	}

	utilizationRate := 0.0
	if totalCapacity > 0 {
		utilizationRate = (float64(totalLoad) / float64(totalCapacity)) * 100
	}

	stats := map[string]interface{}{
		"total_regions":     len(regions),
		"active_regions":    healthyCount + degradedCount,
		"healthy_regions":   healthyCount,
		"degraded_regions":  degradedCount,
		"down_regions":      downCount,
		"total_capacity":    totalCapacity,
		"total_load":        totalLoad,
		"utilization_rate":  math.Round(utilizationRate*100) / 100,
		"overall_health":    overallHealth,
		"last_updated":      time.Now(),
	}

	return stats, nil
}

// RunHealthCheckLoop runs periodic health checks for all regions
func (s *RegionService) RunHealthCheckLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Info().Dur("interval", interval).Msg("Starting region health check loop")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Stopping region health check loop")
			return
		case <-ticker.C:
			regions, err := s.GetAllRegions(ctx)
			if err != nil {
				log.Error().Err(err).Msg("Failed to get regions for health check")
				continue
			}

			for _, region := range regions {
				if !region.IsActive {
					continue
				}

				_, err := s.CheckRegionHealth(ctx, region.Code)
				if err != nil {
					log.Error().Err(err).Str("region", region.Code).Msg("Health check failed")
				}
			}

			log.Debug().Msg(fmt.Sprintf("Completed health check for %d regions", len(regions)))
		}
	}
}
