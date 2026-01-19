package services

import (
	"context"
	"fmt"
	"time"

	"pulse-control-plane/database"
	"pulse-control-plane/models"
	"pulse-control-plane/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// IngressService handles ingress operations
type IngressService struct {
	collection *mongo.Collection
}

// NewIngressService creates a new ingress service
func NewIngressService() *IngressService {
	return &IngressService{
		collection: database.GetCollection("ingresses"),
	}
}

// CreateIngress creates a new ingress endpoint
func (s *IngressService) CreateIngress(ctx context.Context, projectID primitive.ObjectID, project *models.Project, req *models.IngressRequest) (*models.Ingress, error) {
	now := time.Now()
	
	// Create ingress record
	ingress := &models.Ingress{
		ID: primitive.NewObjectID(),
		ProjectID: projectID,
		RoomName: req.RoomName,
		ParticipantName: req.ParticipantName,
		IngressType: req.IngressType,
		Status: models.IngressStatusActive,
		AudioEnabled: req.AudioEnabled,
		VideoEnabled: req.VideoEnabled,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	// In a real implementation, this would call LiveKit Ingress API
	// For now, we'll simulate ingress creation
	ingress.LiveKitIngressID = fmt.Sprintf("IN_%s", primitive.NewObjectID().Hex())
	
	// Generate ingress URLs based on type
	switch req.IngressType {
	case models.IngressTypeRTMP:
		// Generate RTMP URL and stream key
		streamKey := utils.GenerateRandomString(32)
		ingress.RTMPURL = fmt.Sprintf("rtmp://%s/live", project.LiveKitURL)
		ingress.RTMPStreamKey = streamKey
		
	case models.IngressTypeWHIP:
		// Generate WHIP endpoint URL
		ingress.WHIPURL = fmt.Sprintf("https://%s/whip/%s", project.LiveKitURL, ingress.LiveKitIngressID)
		
	case models.IngressTypeURL:
		if req.SourceURL == "" {
			return nil, fmt.Errorf("source_url is required for URL ingress type")
		}
		ingress.SourceURL = req.SourceURL
	}
	
	// Save to database
	_, err := s.collection.InsertOne(ctx, ingress)
	if err != nil {
		return nil, fmt.Errorf("failed to create ingress: %w", err)
	}
	
	return ingress, nil
}

// GetIngress retrieves an ingress by ID
func (s *IngressService) GetIngress(ctx context.Context, ingressID primitive.ObjectID) (*models.Ingress, error) {
	var ingress models.Ingress
	err := s.collection.FindOne(ctx, bson.M{"_id": ingressID, "deleted_at": nil}).Decode(&ingress)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("ingress not found")
		}
		return nil, fmt.Errorf("failed to get ingress: %w", err)
	}
	return &ingress, nil
}

// ListIngresses retrieves all ingresses for a project
func (s *IngressService) ListIngresses(ctx context.Context, projectID primitive.ObjectID, page, limit int) ([]models.Ingress, int64, error) {
	filter := bson.M{
		"project_id": projectID,
		"deleted_at": nil,
	}
	
	// Count total
	total, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count ingresses: %w", err)
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
		return nil, 0, fmt.Errorf("failed to list ingresses: %w", err)
	}
	defer cursor.Close(ctx)
	
	var ingresses []models.Ingress
	if err = cursor.All(ctx, &ingresses); err != nil {
		return nil, 0, fmt.Errorf("failed to decode ingresses: %w", err)
	}
	
	return ingresses, total, nil
}

// DeleteIngress soft deletes an ingress
func (s *IngressService) DeleteIngress(ctx context.Context, ingressID primitive.ObjectID) error {
	// Check if ingress exists
	_, err := s.GetIngress(ctx, ingressID)
	if err != nil {
		return err
	}
	
	// In a real implementation, this would call LiveKit Ingress Delete API
	// For now, we'll just mark as deleted
	
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"status": models.IngressStatusInactive,
			"updated_at": now,
		},
	}
	
	_, err = s.collection.UpdateOne(ctx, bson.M{"_id": ingressID}, update)
	if err != nil {
		return fmt.Errorf("failed to delete ingress: %w", err)
	}
	
	return nil
}

// UpdateIngressStatus updates the status of an ingress (called by webhooks)
func (s *IngressService) UpdateIngressStatus(ctx context.Context, liveKitIngressID string, status models.IngressStatus, errorMsg string) error {
	update := bson.M{
		"$set": bson.M{
			"status": status,
			"updated_at": time.Now(),
		},
	}
	
	if errorMsg != "" {
		update["$set"].(bson.M)["error"] = errorMsg
	}
	
	_, err := s.collection.UpdateOne(ctx, bson.M{"livekit_ingress_id": liveKitIngressID}, update)
	if err != nil {
		return fmt.Errorf("failed to update ingress status: %w", err)
	}
	
	return nil
}

// ToResponse converts Ingress to IngressResponse (safe for API)
func (s *IngressService) ToResponse(ingress *models.Ingress) *models.IngressResponse {
	return &models.IngressResponse{
		ID: ingress.ID.Hex(),
		ProjectID: ingress.ProjectID.Hex(),
		RoomName: ingress.RoomName,
		ParticipantName: ingress.ParticipantName,
		IngressType: ingress.IngressType,
		Status: ingress.Status,
		Error: ingress.Error,
		RTMPURL: ingress.RTMPURL,
		RTMPStreamKey: ingress.RTMPStreamKey,
		WHIPURL: ingress.WHIPURL,
		AudioEnabled: ingress.AudioEnabled,
		VideoEnabled: ingress.VideoEnabled,
		CreatedAt: ingress.CreatedAt,
	}
}
