package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TeamMember represents a member of an organization with specific roles and permissions
type TeamMember struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID        primitive.ObjectID `bson:"org_id" json:"org_id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	Email        string             `bson:"email" json:"email"`
	Name         string             `bson:"name" json:"name"`
	Role         string             `bson:"role" json:"role"` // Owner, Admin, Developer, Viewer
	Status       string             `bson:"status" json:"status"` // Active, Inactive, Pending
	InvitedBy    primitive.ObjectID `bson:"invited_by,omitempty" json:"invited_by,omitempty"`
	InvitedAt    time.Time          `bson:"invited_at,omitempty" json:"invited_at,omitempty"`
	JoinedAt     time.Time          `bson:"joined_at,omitempty" json:"joined_at,omitempty"`
	LastActiveAt time.Time          `bson:"last_active_at" json:"last_active_at"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// TeamMemberInvite represents the input for inviting a team member
type TeamMemberInvite struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required,min=2,max=100"`
	Role  string `json:"role" binding:"required,oneof=Admin Developer Viewer"`
}

// TeamMemberUpdate represents the input for updating a team member's role
type TeamMemberUpdate struct {
	Role string `json:"role" binding:"required,oneof=Owner Admin Developer Viewer"`
}

// TeamMemberResponse represents a team member in API responses
type TeamMemberResponse struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	InvitedAt    time.Time `json:"invited_at,omitempty"`
	JoinedAt     time.Time `json:"joined_at,omitempty"`
	LastActiveAt time.Time `json:"last_active_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// RolePermissions defines what each role can do
var RolePermissions = map[string][]string{
	"Owner": {
		"manage_billing",
		"manage_team",
		"manage_projects",
		"manage_api_keys",
		"view_audit_logs",
		"manage_webhooks",
		"view_usage",
		"manage_organization",
		"delete_organization",
	},
	"Admin": {
		"manage_team",
		"manage_projects",
		"manage_api_keys",
		"view_audit_logs",
		"manage_webhooks",
		"view_usage",
	},
	"Developer": {
		"manage_projects",
		"manage_api_keys",
		"view_audit_logs",
		"view_usage",
	},
	"Viewer": {
		"view_audit_logs",
		"view_usage",
	},
}

// HasPermission checks if a role has a specific permission
func (tm *TeamMember) HasPermission(permission string) bool {
	permissions, exists := RolePermissions[tm.Role]
	if !exists {
		return false
	}

	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// ToResponse converts TeamMember to TeamMemberResponse
func (tm *TeamMember) ToResponse() TeamMemberResponse {
	return TeamMemberResponse{
		ID:           tm.ID.Hex(),
		Email:        tm.Email,
		Name:         tm.Name,
		Role:         tm.Role,
		Status:       tm.Status,
		InvitedAt:    tm.InvitedAt,
		JoinedAt:     tm.JoinedAt,
		LastActiveAt: tm.LastActiveAt,
		CreatedAt:    tm.CreatedAt,
	}
}

// TableName returns the collection name
func (TeamMember) TableName() string {
	return "team_members"
}
