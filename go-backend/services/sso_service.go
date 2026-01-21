package services

import (
	"context"
	"errors"
	"time"

	"pulse-control-plane/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// SSOService handles SSO operations
type SSOService struct {
	db *mongo.Database
}

// NewSSOService creates a new SSO service
func NewSSOService(db *mongo.Database) *SSOService {
	return &SSOService{db: db}
}

// CreateSSOConfig creates a new SSO configuration
func (s *SSOService) CreateSSOConfig(ctx context.Context, config *models.SSOConfig) error {
	// Hash client secret if provided
	if config.ClientSecret != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(config.ClientSecret), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		config.ClientSecret = string(hashed)
	}
	
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	
	coll := s.db.Collection("sso_configs")
	result, err := coll.InsertOne(ctx, config)
	if err != nil {
		return err
	}
	
	config.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetSSOConfig retrieves SSO config for an organization
func (s *SSOService) GetSSOConfig(ctx context.Context, orgID primitive.ObjectID) (*models.SSOConfig, error) {
	coll := s.db.Collection("sso_configs")
	
	var config models.SSOConfig
	err := coll.FindOne(ctx, bson.M{"org_id": orgID}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("SSO configuration not found")
		}
		return nil, err
	}
	
	return &config, nil
}

// UpdateSSOConfig updates SSO configuration
func (s *SSOService) UpdateSSOConfig(ctx context.Context, id primitive.ObjectID, updates bson.M) error {
	// Hash client secret if being updated
	if secret, ok := updates["client_secret"].(string); ok && secret != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		updates["client_secret"] = string(hashed)
	}
	
	updates["updated_at"] = time.Now()
	
	coll := s.db.Collection("sso_configs")
	result, err := coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": updates})
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return errors.New("SSO configuration not found")
	}
	
	return nil
}

// DeleteSSOConfig deletes SSO configuration
func (s *SSOService) DeleteSSOConfig(ctx context.Context, id primitive.ObjectID) error {
	coll := s.db.Collection("sso_configs")
	result, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	
	if result.DeletedCount == 0 {
		return errors.New("SSO configuration not found")
	}
	
	return nil
}

// ValidateOAuthCallback validates OAuth callback and creates session
func (s *SSOService) ValidateOAuthCallback(ctx context.Context, orgID primitive.ObjectID, provider models.SSOProvider, code string) (*models.SSOSession, error) {
	// Get SSO config
	config, err := s.GetSSOConfig(ctx, orgID)
	if err != nil {
		return nil, err
	}
	
	if !config.Enabled {
		return nil, errors.New("SSO is not enabled for this organization")
	}
	
	if config.Provider != provider {
		return nil, errors.New("provider mismatch")
	}
	
	// In production, exchange code for access token with OAuth provider
	// This is a placeholder implementation
	session := &models.SSOSession{
		OrgID:      orgID,
		Provider:   provider,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		CreatedAt:  time.Now(),
	}
	
	coll := s.db.Collection("sso_sessions")
	result, err := coll.InsertOne(ctx, session)
	if err != nil {
		return nil, err
	}
	
	session.ID = result.InsertedID.(primitive.ObjectID)
	return session, nil
}

// ValidateSAMLAssertion validates SAML assertion
func (s *SSOService) ValidateSAMLAssertion(ctx context.Context, orgID primitive.ObjectID, assertion string) (*models.SSOSession, error) {
	// Get SSO config
	config, err := s.GetSSOConfig(ctx, orgID)
	if err != nil {
		return nil, err
	}
	
	if !config.Enabled {
		return nil, errors.New("SSO is not enabled for this organization")
	}
	
	if config.Provider != models.SSOProviderSAML {
		return nil, errors.New("SAML is not configured")
	}
	
	// In production, validate SAML assertion with certificate
	// This is a placeholder implementation
	session := &models.SSOSession{
		OrgID:     orgID,
		Provider:  models.SSOProviderSAML,
		ExpiresAt: time.Now().Add(8 * time.Hour),
		CreatedAt: time.Now(),
	}
	
	coll := s.db.Collection("sso_sessions")
	result, err := coll.InsertOne(ctx, session)
	if err != nil {
		return nil, err
	}
	
	session.ID = result.InsertedID.(primitive.ObjectID)
	return session, nil
}
