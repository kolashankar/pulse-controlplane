package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Organization represents a customer organization/account
type Organization struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string             `bson:"name" json:"name" binding:"required"`
	AdminEmail string             `bson:"admin_email" json:"admin_email" binding:"required,email"`
	Plan       string             `bson:"plan" json:"plan"` // Free, Pro, Enterprise
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
	IsDeleted  bool               `bson:"is_deleted" json:"-"` // Soft delete flag
}

// OrganizationCreate represents the input for creating an organization
type OrganizationCreate struct {
	Name       string `json:"name" binding:"required,min=3,max=100"`
	AdminEmail string `json:"admin_email" binding:"required,email"`
	Plan       string `json:"plan" binding:"omitempty,oneof=Free Pro Enterprise"`
}

// OrganizationUpdate represents the input for updating an organization
type OrganizationUpdate struct {
	Name string `json:"name" binding:"omitempty,min=3,max=100"`
	Plan string `json:"plan" binding:"omitempty,oneof=Free Pro Enterprise"`
}

// TableName returns the collection name
func (Organization) TableName() string {
	return "organizations"
}

// NewOrganization creates a new organization with default values
func NewOrganization(name, adminEmail string) *Organization {
	return &Organization{
		ID:         primitive.NewObjectID(),
		Name:       name,
		AdminEmail: adminEmail,
		Plan:       "Free", // Default plan
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		IsDeleted:  false,
	}
}
