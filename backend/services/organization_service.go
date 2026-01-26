package services

import (
	"context"
	"errors"
	"time"

	"pulse-control-plane/database"
	"pulse-control-plane/models"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrganizationService struct {
	collection *mongo.Collection
}

func NewOrganizationService() *OrganizationService {
	return &OrganizationService{
		collection: database.GetCollection("organizations"),
	}
}

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(ctx context.Context, input *models.OrganizationCreate) (*models.Organization, error) {
	// Check if organization with same admin email already exists
	var existing models.Organization
	err := s.collection.FindOne(ctx, bson.M{
		"admin_email": input.AdminEmail,
		"is_deleted":  false,
	}).Decode(&existing)

	if err == nil {
		return nil, errors.New("organization with this admin email already exists")
	}

	// Create new organization
	org := &models.Organization{
		ID:         primitive.NewObjectID(),
		Name:       input.Name,
		AdminEmail: input.AdminEmail,
		Plan:       input.Plan,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		IsDeleted:  false,
	}

	// Set default plan if not provided
	if org.Plan == "" {
		org.Plan = "Free"
	}

	_, err = s.collection.InsertOne(ctx, org)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create organization")
		return nil, err
	}

	log.Info().Str("org_id", org.ID.Hex()).Str("name", org.Name).Msg("Organization created")
	return org, nil
}

// GetOrganization retrieves an organization by ID
func (s *OrganizationService) GetOrganization(ctx context.Context, id primitive.ObjectID) (*models.Organization, error) {
	var org models.Organization
	err := s.collection.FindOne(ctx, bson.M{
		"_id":        id,
		"is_deleted": false,
	}).Decode(&org)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("organization not found")
		}
		return nil, err
	}

	return &org, nil
}

// ListOrganizations retrieves all organizations with pagination
func (s *OrganizationService) ListOrganizations(ctx context.Context, page, limit int64, searchQuery string) ([]*models.Organization, int64, error) {
	filter := bson.M{"is_deleted": false}

	// Add search filter if provided
	if searchQuery != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": searchQuery, "$options": "i"}},
			{"admin_email": bson.M{"$regex": searchQuery, "$options": "i"}},
		}
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

	var organizations []*models.Organization
	if err = cursor.All(ctx, &organizations); err != nil {
		return nil, 0, err
	}

	return organizations, totalCount, nil
}

// UpdateOrganization updates an organization
func (s *OrganizationService) UpdateOrganization(ctx context.Context, id primitive.ObjectID, input *models.OrganizationUpdate) (*models.Organization, error) {
	// Build update document
	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if input.Name != "" {
		update["$set"].(bson.M)["name"] = input.Name
	}
	if input.Plan != "" {
		update["$set"].(bson.M)["plan"] = input.Plan
	}

	// Update the organization
	result := s.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id, "is_deleted": false},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var org models.Organization
	if err := result.Decode(&org); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("organization not found")
		}
		return nil, err
	}

	log.Info().Str("org_id", org.ID.Hex()).Msg("Organization updated")
	return &org, nil
}

// DeleteOrganization soft deletes an organization
func (s *OrganizationService) DeleteOrganization(ctx context.Context, id primitive.ObjectID) error {
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
		return errors.New("organization not found")
	}

	log.Info().Str("org_id", id.Hex()).Msg("Organization deleted (soft)")
	return nil
}
