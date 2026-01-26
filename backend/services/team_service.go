package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"pulse-control-plane/database"
	"pulse-control-plane/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TeamService handles team member operations
type TeamService struct {
	db                *mongo.Database
	teamMembersColl   *mongo.Collection
	invitationsColl   *mongo.Collection
	organizationsColl *mongo.Collection
}

// NewTeamService creates a new team service
func NewTeamService() *TeamService {
	db := database.GetDB()
	return &TeamService{
		db:                db,
		teamMembersColl:   db.Collection(models.TeamMember{}.TableName()),
		invitationsColl:   db.Collection(models.Invitation{}.TableName()),
		organizationsColl: db.Collection(models.Organization{}.TableName()),
	}
}

// ListTeamMembers lists all team members for an organization
func (s *TeamService) ListTeamMembers(ctx context.Context, orgID primitive.ObjectID, page, limit int) ([]models.TeamMember, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	skip := (page - 1) * limit

	filter := bson.M{"org_id": orgID}
	
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := s.teamMembersColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list team members: %w", err)
	}
	defer cursor.Close(ctx)

	var members []models.TeamMember
	if err := cursor.All(ctx, &members); err != nil {
		return nil, 0, fmt.Errorf("failed to decode team members: %w", err)
	}

	total, err := s.teamMembersColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count team members: %w", err)
	}

	return members, total, nil
}

// InviteTeamMember invites a new team member
func (s *TeamService) InviteTeamMember(ctx context.Context, orgID primitive.ObjectID, invitedBy primitive.ObjectID, invite models.TeamMemberInvite) (*models.Invitation, error) {
	// Check if organization exists
	var org models.Organization
	err := s.organizationsColl.FindOne(ctx, bson.M{"_id": orgID, "is_deleted": false}).Decode(&org)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("organization not found")
		}
		return nil, fmt.Errorf("failed to find organization: %w", err)
	}

	// Check if user is already a team member
	existingMember := &models.TeamMember{}
	err = s.teamMembersColl.FindOne(ctx, bson.M{"org_id": orgID, "email": invite.Email}).Decode(existingMember)
	if err == nil {
		return nil, errors.New("user is already a team member")
	}

	// Check if there's a pending invitation
	existingInvitation := &models.Invitation{}
	err = s.invitationsColl.FindOne(ctx, bson.M{
		"org_id": orgID,
		"email":  invite.Email,
		"status": "Pending",
	}).Decode(existingInvitation)
	if err == nil {
		return nil, errors.New("invitation already sent to this email")
	}

	// Generate secure token
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create invitation
	invitation := &models.Invitation{
		OrgID:     orgID,
		Email:     invite.Email,
		Name:      invite.Name,
		Role:      invite.Role,
		Token:     token,
		InvitedBy: invitedBy,
		Status:    "Pending",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := s.invitationsColl.InsertOne(ctx, invitation)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	invitation.ID = result.InsertedID.(primitive.ObjectID)

	return invitation, nil
}

// AcceptInvitation accepts an invitation and creates a team member
func (s *TeamService) AcceptInvitation(ctx context.Context, token string) (*models.TeamMember, error) {
	// Find invitation by token
	var invitation models.Invitation
	err := s.invitationsColl.FindOne(ctx, bson.M{"token": token}).Decode(&invitation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("invitation not found")
		}
		return nil, fmt.Errorf("failed to find invitation: %w", err)
	}

	// Validate invitation
	if !invitation.IsValid() {
		return nil, errors.New("invitation is no longer valid")
	}

	// Create team member
	now := time.Now()
	member := &models.TeamMember{
		OrgID:        invitation.OrgID,
		Email:        invitation.Email,
		Name:         invitation.Name,
		Role:         invitation.Role,
		Status:       "Active",
		InvitedBy:    invitation.InvitedBy,
		InvitedAt:    invitation.CreatedAt,
		JoinedAt:     now,
		LastActiveAt: now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	result, err := s.teamMembersColl.InsertOne(ctx, member)
	if err != nil {
		return nil, fmt.Errorf("failed to create team member: %w", err)
	}

	member.ID = result.InsertedID.(primitive.ObjectID)

	// Update invitation status
	_, err = s.invitationsColl.UpdateOne(
		ctx,
		bson.M{"_id": invitation.ID},
		bson.M{
			"$set": bson.M{
				"status":      "Accepted",
				"accepted_at": now,
				"updated_at":  now,
			},
		},
	)
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to update invitation status: %v\n", err)
	}

	return member, nil
}

// RemoveTeamMember removes a team member from the organization
func (s *TeamService) RemoveTeamMember(ctx context.Context, orgID primitive.ObjectID, userID primitive.ObjectID) error {
	// Check if member is the owner
	var member models.TeamMember
	err := s.teamMembersColl.FindOne(ctx, bson.M{"_id": userID, "org_id": orgID}).Decode(&member)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("team member not found")
		}
		return fmt.Errorf("failed to find team member: %w", err)
	}

	if member.Role == "Owner" {
		return errors.New("cannot remove the organization owner")
	}

	// Delete team member
	result, err := s.teamMembersColl.DeleteOne(ctx, bson.M{"_id": userID, "org_id": orgID})
	if err != nil {
		return fmt.Errorf("failed to delete team member: %w", err)
	}

	if result.DeletedCount == 0 {
		return errors.New("team member not found")
	}

	return nil
}

// UpdateTeamMemberRole updates a team member's role
func (s *TeamService) UpdateTeamMemberRole(ctx context.Context, orgID primitive.ObjectID, userID primitive.ObjectID, newRole string) (*models.TeamMember, error) {
	// Check if member exists
	var member models.TeamMember
	err := s.teamMembersColl.FindOne(ctx, bson.M{"_id": userID, "org_id": orgID}).Decode(&member)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("team member not found")
		}
		return nil, fmt.Errorf("failed to find team member: %w", err)
	}

	// Cannot change owner's role
	if member.Role == "Owner" {
		return nil, errors.New("cannot change the role of the organization owner")
	}

	// Update role
	update := bson.M{
		"$set": bson.M{
			"role":       newRole,
			"updated_at": time.Now(),
		},
	}

	err = s.teamMembersColl.FindOneAndUpdate(
		ctx,
		bson.M{"_id": userID, "org_id": orgID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&member)

	if err != nil {
		return nil, fmt.Errorf("failed to update team member role: %w", err)
	}

	return &member, nil
}

// GetTeamMember gets a team member by ID
func (s *TeamService) GetTeamMember(ctx context.Context, orgID primitive.ObjectID, userID primitive.ObjectID) (*models.TeamMember, error) {
	var member models.TeamMember
	err := s.teamMembersColl.FindOne(ctx, bson.M{"_id": userID, "org_id": orgID}).Decode(&member)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("team member not found")
		}
		return nil, fmt.Errorf("failed to find team member: %w", err)
	}

	return &member, nil
}

// ListPendingInvitations lists all pending invitations for an organization
func (s *TeamService) ListPendingInvitations(ctx context.Context, orgID primitive.ObjectID) ([]models.Invitation, error) {
	filter := bson.M{
		"org_id": orgID,
		"status": "Pending",
	}

	cursor, err := s.invitationsColl.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}
	defer cursor.Close(ctx)

	var invitations []models.Invitation
	if err := cursor.All(ctx, &invitations); err != nil {
		return nil, fmt.Errorf("failed to decode invitations: %w", err)
	}

	return invitations, nil
}

// RevokeInvitation revokes a pending invitation
func (s *TeamService) RevokeInvitation(ctx context.Context, orgID primitive.ObjectID, invitationID primitive.ObjectID) error {
	update := bson.M{
		"$set": bson.M{
			"status":     "Revoked",
			"updated_at": time.Now(),
		},
	}

	result, err := s.invitationsColl.UpdateOne(
		ctx,
		bson.M{"_id": invitationID, "org_id": orgID, "status": "Pending"},
		update,
	)

	if err != nil {
		return fmt.Errorf("failed to revoke invitation: %w", err)
	}

	if result.ModifiedCount == 0 {
		return errors.New("invitation not found or already processed")
	}

	return nil
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
