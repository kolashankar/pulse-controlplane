package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Invitation represents an invitation to join an organization
type Invitation struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID      primitive.ObjectID `bson:"org_id" json:"org_id"`
	Email      string             `bson:"email" json:"email"`
	Name       string             `bson:"name" json:"name"`
	Role       string             `bson:"role" json:"role"`
	Token      string             `bson:"token" json:"token"` // Secure random token
	InvitedBy  primitive.ObjectID `bson:"invited_by" json:"invited_by"`
	Status     string             `bson:"status" json:"status"` // Pending, Accepted, Expired, Revoked
	ExpiresAt  time.Time          `bson:"expires_at" json:"expires_at"`
	AcceptedAt time.Time          `bson:"accepted_at,omitempty" json:"accepted_at,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

// InvitationResponse represents an invitation in API responses
type InvitationResponse struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	Role       string    `json:"role"`
	Status     string    `json:"status"`
	ExpiresAt  time.Time `json:"expires_at"`
	AcceptedAt time.Time `json:"accepted_at,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// IsExpired checks if the invitation has expired
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsValid checks if the invitation is valid (pending and not expired)
func (i *Invitation) IsValid() bool {
	return i.Status == "Pending" && !i.IsExpired()
}

// ToResponse converts Invitation to InvitationResponse
func (i *Invitation) ToResponse() InvitationResponse {
	return InvitationResponse{
		ID:         i.ID.Hex(),
		Email:      i.Email,
		Name:       i.Name,
		Role:       i.Role,
		Status:     i.Status,
		ExpiresAt:  i.ExpiresAt,
		AcceptedAt: i.AcceptedAt,
		CreatedAt:  i.CreatedAt,
	}
}

// TableName returns the collection name
func (Invitation) TableName() string {
	return "invitations"
}
