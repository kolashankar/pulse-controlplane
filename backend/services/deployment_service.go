package services

import (
	"context"
	"errors"
	"time"

	"pulse-control-plane/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DeploymentService handles deployment configuration operations
type DeploymentService struct {
	db *mongo.Database
}

// NewDeploymentService creates a new deployment service
func NewDeploymentService(db *mongo.Database) *DeploymentService {
	return &DeploymentService{db: db}
}

// CreateDeploymentConfig creates a new deployment configuration
func (s *DeploymentService) CreateDeploymentConfig(ctx context.Context, config *models.DeploymentConfig) error {
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	config.IsActive = true
	
	coll := s.db.Collection("deployment_configs")
	
	// Deactivate any existing active deployment for this org
	_, err := coll.UpdateMany(ctx,
		bson.M{"org_id": config.OrgID, "is_active": true},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		return err
	}
	
	result, err := coll.InsertOne(ctx, config)
	if err != nil {
		return err
	}
	
	config.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetDeploymentConfig retrieves deployment config for an organization
func (s *DeploymentService) GetDeploymentConfig(ctx context.Context, orgID primitive.ObjectID) (*models.DeploymentConfig, error) {
	coll := s.db.Collection("deployment_configs")
	
	var config models.DeploymentConfig
	err := coll.FindOne(ctx, bson.M{"org_id": orgID, "is_active": true}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("deployment configuration not found")
		}
		return nil, err
	}
	
	return &config, nil
}

// UpdateDeploymentConfig updates deployment configuration
func (s *DeploymentService) UpdateDeploymentConfig(ctx context.Context, id primitive.ObjectID, updates bson.M) error {
	updates["updated_at"] = time.Now()
	
	coll := s.db.Collection("deployment_configs")
	result, err := coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": updates})
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return errors.New("deployment configuration not found")
	}
	
	return nil
}

// DeleteDeploymentConfig deletes deployment configuration
func (s *DeploymentService) DeleteDeploymentConfig(ctx context.Context, id primitive.ObjectID) error {
	coll := s.db.Collection("deployment_configs")
	result, err := coll.UpdateOne(ctx, 
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"is_active": false, "updated_at": time.Now()}},
	)
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return errors.New("deployment configuration not found")
	}
	
	return nil
}

// ValidateLicense validates a license key
func (s *DeploymentService) ValidateLicense(ctx context.Context, licenseKey string) (bool, error) {
	coll := s.db.Collection("deployment_configs")
	
	var config models.DeploymentConfig
	err := coll.FindOne(ctx, bson.M{"license_key": licenseKey, "is_active": true}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, errors.New("invalid license key")
		}
		return false, err
	}
	
	// Check license expiry
	if config.LicenseExpiry != nil && config.LicenseExpiry.Before(time.Now()) {
		return false, errors.New("license has expired")
	}
	
	return true, nil
}

// TrackDeploymentMetrics records deployment health metrics
func (s *DeploymentService) TrackDeploymentMetrics(ctx context.Context, metrics *models.DeploymentMetrics) error {
	metrics.Timestamp = time.Now()
	
	coll := s.db.Collection("deployment_metrics")
	result, err := coll.InsertOne(ctx, metrics)
	if err != nil {
		return err
	}
	
	metrics.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetDeploymentMetrics retrieves recent deployment metrics
func (s *DeploymentService) GetDeploymentMetrics(ctx context.Context, deploymentID primitive.ObjectID, limit int) ([]models.DeploymentMetrics, error) {
	coll := s.db.Collection("deployment_metrics")
	
	opts := options.Find().SetSort(bson.D{{"timestamp", -1}}).SetLimit(int64(limit))
	cursor, err := coll.Find(ctx, bson.M{"deployment_id": deploymentID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var metrics []models.DeploymentMetrics
	if err := cursor.All(ctx, &metrics); err != nil {
		return nil, err
	}
	
	return metrics, nil
}

// GetLatestDeploymentMetrics retrieves the latest metrics
func (s *DeploymentService) GetLatestDeploymentMetrics(ctx context.Context, deploymentID primitive.ObjectID) (*models.DeploymentMetrics, error) {
	coll := s.db.Collection("deployment_metrics")
	
	opts := options.FindOne().SetSort(bson.D{{"timestamp", -1}})
	var metrics models.DeploymentMetrics
	err := coll.FindOne(ctx, bson.M{"deployment_id": deploymentID}, opts).Decode(&metrics)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("no metrics found")
		}
		return nil, err
	}
	
	return &metrics, nil
}

// ListDeploymentConfigs lists all deployment configurations
func (s *DeploymentService) ListDeploymentConfigs(ctx context.Context, deploymentType models.DeploymentType, page, limit int) ([]models.DeploymentConfig, int64, error) {
	coll := s.db.Collection("deployment_configs")
	
	filter := bson.M{"is_active": true}
	if deploymentType != "" {
		filter["deployment_type"] = deploymentType
	}
	
	// Count total
	totalCount, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	
	// Find with pagination
	opts := options.Find().
		SetSort(bson.D{{"created_at", -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))
	
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	
	var configs []models.DeploymentConfig
	if err := cursor.All(ctx, &configs); err != nil {
		return nil, 0, err
	}
	
	return configs, totalCount, nil
}
