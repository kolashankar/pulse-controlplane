package services

import (
	"context"
	"errors"
	"time"

	"pulse-control-plane/database"
	"pulse-control-plane/models"
	"pulse-control-plane/utils"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProjectService struct {
	collection *mongo.Collection
}

func NewProjectService() *ProjectService {
	return &ProjectService{
		collection: database.GetCollection("projects"),
	}
}

// CreateProject creates a new project with API keys
func (s *ProjectService) CreateProject(ctx context.Context, orgID primitive.ObjectID, input *models.ProjectCreate) (*models.Project, string, error) {
	// Generate API key and secret
	apiKey, err := utils.GenerateAPIKey()
	if err != nil {
		return nil, "", errors.New("failed to generate API key")
	}

	apiSecret, err := utils.GenerateAPISecret()
	if err != nil {
		return nil, "", errors.New("failed to generate API secret")
	}

	// Hash the secret for storage
	hashedSecret, err := utils.HashSecret(apiSecret)
	if err != nil {
		return nil, "", errors.New("failed to hash API secret")
	}

	// Create project
	project := &models.Project{
		ID:                  primitive.NewObjectID(),
		OrgID:               orgID,
		Name:                input.Name,
		PulseAPIKey:         apiKey,
		PulseAPISecret:      hashedSecret,
		WebhookURL:          input.WebhookURL,
		StorageConfig:       input.Storage,
		Region:              input.Region,
		LiveKitURL:          "", // Will be set based on region
		ChatEnabled:         false,
		VideoEnabled:        true, // Default enabled
		ActivityFeedEnabled: false,
		ModerationEnabled:   false,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		IsDeleted:           false,
	}

	// Set LiveKit URL based on region (mock for now)
	project.LiveKitURL = s.getLiveKitURLForRegion(input.Region)

	_, err = s.collection.InsertOne(ctx, project)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create project")
		return nil, "", err
	}

	log.Info().
		Str("project_id", project.ID.Hex()).
		Str("org_id", orgID.Hex()).
		Str("name", project.Name).
		Msg("Project created")

	// Return project and plain API secret (only time it's returned)
	return project, apiSecret, nil
}

// GetProject retrieves a project by ID
func (s *ProjectService) GetProject(ctx context.Context, id primitive.ObjectID) (*models.Project, error) {
	var project models.Project
	err := s.collection.FindOne(ctx, bson.M{
		"_id":        id,
		"is_deleted": false,
	}).Decode(&project)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("project not found")
		}
		return nil, err
	}

	return &project, nil
}

// ListProjects retrieves all projects with pagination
func (s *ProjectService) ListProjects(ctx context.Context, orgID *primitive.ObjectID, page, limit int64, searchQuery string) ([]*models.Project, int64, error) {
	filter := bson.M{"is_deleted": false}

	// Filter by organization if provided
	if orgID != nil {
		filter["org_id"] = *orgID
	}

	// Add search filter if provided
	if searchQuery != "" {
		filter["name"] = bson.M{"$regex": searchQuery, "$options": "i"}
	}

	// Get total count
	totalCount, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Find with pagination
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var projects []*models.Project
	if err = cursor.All(ctx, &projects); err != nil {
		return nil, 0, err
	}

	return projects, totalCount, nil
}

// UpdateProject updates a project
func (s *ProjectService) UpdateProject(ctx context.Context, id primitive.ObjectID, input *models.ProjectUpdate) (*models.Project, error) {
	// Build update document
	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if input.Name != "" {
		update["$set"].(bson.M)["name"] = input.Name
	}
	if input.WebhookURL != "" {
		update["$set"].(bson.M)["webhook_url"] = input.WebhookURL
	}

	// Update storage config if provided
	if input.Storage.Provider != "" {
		update["$set"].(bson.M)["storage_config"] = input.Storage
	}

	// Update the project
	result := s.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id, "is_deleted": false},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var project models.Project
	if err := result.Decode(&project); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("project not found")
		}
		return nil, err
	}

	log.Info().Str("project_id", project.ID.Hex()).Msg("Project updated")
	return &project, nil
}

// DeleteProject soft deletes a project
func (s *ProjectService) DeleteProject(ctx context.Context, id primitive.ObjectID) error {
	// Soft delete by setting is_deleted flag
	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"updated_at": time.Now(),
		},
	}

	result, err := s.collection.UpdateOne(
		ctx,
		bson.M{"_id": id, "is_deleted": false},
		update,
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("project not found")
	}

	log.Info().Str("project_id", id.Hex()).Msg("Project deleted (soft)")
	return nil
}

// RegenerateAPIKeys generates new API keys for a project
func (s *ProjectService) RegenerateAPIKeys(ctx context.Context, id primitive.ObjectID) (string, string, error) {
	// Generate new API key and secret
	apiKey, err := utils.GenerateAPIKey()
	if err != nil {
		return "", "", errors.New("failed to generate API key")
	}

	apiSecret, err := utils.GenerateAPISecret()
	if err != nil {
		return "", "", errors.New("failed to generate API secret")
	}

	// Hash the secret for storage
	hashedSecret, err := utils.HashSecret(apiSecret)
	if err != nil {
		return "", "", errors.New("failed to hash API secret")
	}

	// Update the project
	update := bson.M{
		"$set": bson.M{
			"pulse_api_key":    apiKey,
			"pulse_api_secret": hashedSecret,
			"updated_at":       time.Now(),
		},
	}

	result, err := s.collection.UpdateOne(
		ctx,
		bson.M{"_id": id, "is_deleted": false},
		update,
	)

	if err != nil {
		return "", "", err
	}

	if result.MatchedCount == 0 {
		return "", "", errors.New("project not found")
	}

	log.Info().Str("project_id", id.Hex()).Msg("API keys regenerated")

	// Return new keys (only time secret is returned)
	return apiKey, apiSecret, nil
}

// getLiveKitURLForRegion returns the LiveKit URL for a region
func (s *ProjectService) getLiveKitURLForRegion(region string) string {
	// Mock implementation - in production, this would map to actual LiveKit servers
	regionMap := map[string]string{
		"us-east":     "wss://livekit-us-east.pulse.io",
		"us-west":     "wss://livekit-us-west.pulse.io",
		"eu-west":     "wss://livekit-eu-west.pulse.io",
		"eu-central":  "wss://livekit-eu-central.pulse.io",
		"asia-south":  "wss://livekit-asia-south.pulse.io",
		"asia-east":   "wss://livekit-asia-east.pulse.io",
	}

	if url, ok := regionMap[region]; ok {
		return url
	}
	return "wss://livekit-us-east.pulse.io" // Default
}
