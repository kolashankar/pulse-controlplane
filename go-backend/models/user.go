package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a team member
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email" binding:"required,email"`
	Name         string             `bson:"name" json:"name"`
	PasswordHash string             `bson:"password_hash" json:"-"` // Never expose
	Role         string             `bson:"role" json:"role"` // Owner, Admin, Developer, Viewer
	OrgID        primitive.ObjectID `bson:"org_id" json:"org_id"`
	IsActive     bool               `bson:"is_active" json:"is_active"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// UserCreate represents the input for creating a user
type UserCreate struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"omitempty,oneof=Admin Developer Viewer"`
}

// UserLogin represents the login input
type UserLogin struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// TableName returns the collection name
func (User) TableName() string {
	return "users"
}
