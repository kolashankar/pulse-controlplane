package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SSOProvider represents SSO provider types
type SSOProvider string

const (
	SSOProviderGoogle    SSOProvider = "google"
	SSOProviderMicrosoft SSOProvider = "microsoft"
	SSOProviderGitHub    SSOProvider = "github"
	SSOProviderSAML      SSOProvider = "saml"
)

// SSOConfig stores SSO configuration for an organization
type SSOConfig struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID          primitive.ObjectID `bson:"org_id" json:"org_id" binding:"required"`
	Provider       SSOProvider        `bson:"provider" json:"provider" binding:"required"`
	Enabled        bool               `bson:"enabled" json:"enabled"`
	
	// OAuth 2.0 Configuration
	ClientID       string `bson:"client_id,omitempty" json:"client_id,omitempty"`
	ClientSecret   string `bson:"client_secret,omitempty" json:"-"` // Never expose in JSON
	RedirectURL    string `bson:"redirect_url,omitempty" json:"redirect_url,omitempty"`
	Scopes         []string `bson:"scopes,omitempty" json:"scopes,omitempty"`
	
	// SAML 2.0 Configuration
	EntityID       string `bson:"entity_id,omitempty" json:"entity_id,omitempty"`
	SSOURL         string `bson:"sso_url,omitempty" json:"sso_url,omitempty"`
	Certificate    string `bson:"certificate,omitempty" json:"certificate,omitempty"`
	MetadataURL    string `bson:"metadata_url,omitempty" json:"metadata_url,omitempty"`
	
	// Domain restrictions
	AllowedDomains []string `bson:"allowed_domains,omitempty" json:"allowed_domains,omitempty"`
	
	// Auto-provisioning
	AutoProvision  bool   `bson:"auto_provision" json:"auto_provision"`
	DefaultRole    string `bson:"default_role" json:"default_role"` // viewer, developer, admin
	
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
}

// SSOSession represents an active SSO session
type SSOSession struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID          primitive.ObjectID `bson:"org_id" json:"org_id"`
	UserID         primitive.ObjectID `bson:"user_id" json:"user_id"`
	Provider       SSOProvider        `bson:"provider" json:"provider"`
	ExternalID     string             `bson:"external_id" json:"external_id"` // ID from SSO provider
	Email          string             `bson:"email" json:"email"`
	AccessToken    string             `bson:"access_token" json:"-"`
	RefreshToken   string             `bson:"refresh_token,omitempty" json:"-"`
	ExpiresAt      time.Time          `bson:"expires_at" json:"expires_at"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}
