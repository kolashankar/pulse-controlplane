package services

import (
	"context"
	"fmt"
	"time"

	"pulse-control-plane/database"
	"pulse-control-plane/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EgressService handles egress operations
type EgressService struct {
	collection *mongo.Collection
	cdnService *CDNService
}

// NewEgressService creates a new egress service
func NewEgressService() *EgressService {
	return &EgressService{
		collection: database.GetCollection("egresses"),
		cdnService: NewCDNService(),
	}
}

// StartEgress starts a new egress session
func (s *EgressService) StartEgress(ctx context.Context, projectID primitive.ObjectID, project *models.Project, req *models.EgressRequest) (*models.Egress, error) {
	now := time.Now()
	
	// Create egress record
	egress := &models.Egress{
		ID: primitive.NewObjectID(),
		ProjectID: projectID,
		RoomName: req.RoomName,
		EgressType: req.EgressType,
		OutputType: req.OutputType,
		LayoutType: req.LayoutType,
		Status: models.EgressStatusPending,
		StorageBucket: project.StorageConfig.Bucket,
		StorageRegion: project.StorageConfig.Region,
		StorageAccessKey: project.StorageConfig.AccessKeyID,
		StorageSecretKey: project.StorageConfig.SecretAccessKey,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	// Generate output URL based on output type
	var outputURL string
	var err error
	
	switch req.OutputType {
	case models.OutputTypeHLS:
		// Generate HLS output path
		filename := req.Filename
		if filename == "" {
			filename = fmt.Sprintf("%s_%d", req.RoomName, time.Now().Unix())
		}
		outputURL = fmt.Sprintf("s3://%s/hls/%s/%s.m3u8", 
			egress.StorageBucket, 
			egress.ProjectID.Hex(), 
			filename)
		
		// Generate CDN playback URL
		egress.CDNPlaybackURL = s.cdnService.GenerateHLSPlaybackURL(egress.ProjectID.Hex(), filename)
		
	case models.OutputTypeRTMP:
		if req.RTMPURL == "" {
			return nil, fmt.Errorf("rtmp_url is required for RTMP output")
		}
		outputURL = req.RTMPURL
		egress.RTMPURL = req.RTMPURL
		
	case models.OutputTypeFile:
		// Generate file output path
		filename := req.Filename
		if filename == "" {
			filename = fmt.Sprintf("%s_%d.mp4", req.RoomName, time.Now().Unix())
		}
		outputURL = fmt.Sprintf("s3://%s/recordings/%s/%s", 
			egress.StorageBucket, 
			egress.ProjectID.Hex(), 
			filename)
		
		// Generate CDN URL for file
		egress.CDNPlaybackURL = s.cdnService.GenerateFileURL(egress.ProjectID.Hex(), filename)
	}
	
	egress.OutputURL = outputURL
	
	// In a real implementation, this would call LiveKit Egress API
	// For now, we'll simulate the egress creation
	egress.LiveKitEgressID = fmt.Sprintf("EG_%s", primitive.NewObjectID().Hex())
	egress.Status = models.EgressStatusActive
	startTime := time.Now()
	egress.StartedAt = &startTime
	
	// Save to database
	_, err = s.collection.InsertOne(ctx, egress)
	if err != nil {
		return nil, fmt.Errorf("failed to create egress: %w", err)
	}
	
	return egress, nil
}

// StopEgress stops an active egress session
func (s *EgressService) StopEgress(ctx context.Context, egressID primitive.ObjectID) (*models.Egress, error) {
	// Get egress
	egress, err := s.GetEgress(ctx, egressID)
	if err != nil {
		return nil, err
	}
	
	if egress.Status != models.EgressStatusActive {
		return nil, fmt.Errorf("egress is not active")
	}
	
	// In a real implementation, this would call LiveKit Egress Stop API
	// For now, we'll simulate stopping
	
	endTime := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status": models.EgressStatusEnded,
			"ended_at": endTime,
			"updated_at": time.Now(),
		},
	}
	
	if egress.StartedAt != nil {
		duration := endTime.Sub(*egress.StartedAt).Seconds()
		update["$set"].(bson.M)["duration_seconds"] = int64(duration)
	}
	
	_, err = s.collection.UpdateOne(ctx, bson.M{"_id": egressID}, update)
	if err != nil {
		return nil, fmt.Errorf("failed to stop egress: %w", err)
	}
	
	// Get updated egress
	return s.GetEgress(ctx, egressID)
}

// GetEgress retrieves an egress by ID
func (s *EgressService) GetEgress(ctx context.Context, egressID primitive.ObjectID) (*models.Egress, error) {
	var egress models.Egress
	err := s.collection.FindOne(ctx, bson.M{"_id": egressID}).Decode(&egress)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("egress not found")
		}
		return nil, fmt.Errorf("failed to get egress: %w", err)
	}
	return &egress, nil
}

// ListEgresses retrieves all egresses for a project
func (s *EgressService) ListEgresses(ctx context.Context, projectID primitive.ObjectID, page, limit int) ([]models.Egress, int64, error) {
	filter := bson.M{"project_id": projectID}
	
	// Count total
	total, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count egresses: %w", err)
	}
	
	// Calculate skip
	skip := (page - 1) * limit
	
	// Find with pagination
	cursor, err := s.collection.Find(ctx, filter, 
		&mongo.Options{
			Skip: &skip,
			Limit: &limit,
			Sort: bson.M{"created_at": -1},
		},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list egresses: %w", err)
	}
	defer cursor.Close(ctx)
	
	var egresses []models.Egress
	if err = cursor.All(ctx, &egresses); err != nil {
		return nil, 0, fmt.Errorf("failed to decode egresses: %w", err)
	}
	
	return egresses, total, nil
}

// UpdateEgressStatus updates the status of an egress (called by webhooks)
func (s *EgressService) UpdateEgressStatus(ctx context.Context, liveKitEgressID string, status models.EgressStatus, errorMsg string) error {
	update := bson.M{
		"$set": bson.M{
			"status": status,
			"updated_at": time.Now(),
		},
	}
	
	if errorMsg != "" {
		update["$set"].(bson.M)["error"] = errorMsg
	}
	
	if status == models.EgressStatusEnded {
		endTime := time.Now()
		update["$set"].(bson.M)["ended_at"] = endTime
	}
	
	_, err := s.collection.UpdateOne(ctx, bson.M{"livekit_egress_id": liveKitEgressID}, update)
	if err != nil {
		return fmt.Errorf("failed to update egress status: %w", err)
	}
	
	return nil
}

// ToResponse converts Egress to EgressResponse (safe for API)
func (s *EgressService) ToResponse(egress *models.Egress) *models.EgressResponse {
	return &models.EgressResponse{
		ID: egress.ID.Hex(),
		ProjectID: egress.ProjectID.Hex(),
		RoomName: egress.RoomName,
		EgressType: egress.EgressType,
		OutputType: egress.OutputType,
		LayoutType: egress.LayoutType,
		Status: egress.Status,
		Error: egress.Error,
		CDNPlaybackURL: egress.CDNPlaybackURL,
		DurationSeconds: egress.DurationSeconds,
		FileSizeBytes: egress.FileSizeBytes,
		StartedAt: egress.StartedAt,
		EndedAt: egress.EndedAt,
		CreatedAt: egress.CreatedAt,
	}
}
