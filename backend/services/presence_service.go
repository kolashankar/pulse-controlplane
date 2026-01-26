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

const (
	// TypingIndicatorTTL is the duration after which typing indicators expire
	TypingIndicatorTTL = 10 * time.Second
	// PresenceTTL is the duration after which a user is considered offline
	PresenceTTL = 5 * time.Minute
	// ActivityTTL is the duration for activity tracking
	ActivityTTL = 30 * time.Minute
)

// PresenceService handles real-time presence operations
type PresenceService struct {
	db *mongo.Database
}

// NewPresenceService creates a new presence service
func NewPresenceService(db *mongo.Database) *PresenceService {
	return &PresenceService{db: db}
}

// SetOnline marks a user as online
func (s *PresenceService) SetOnline(ctx context.Context, presence *models.UserPresence) error {
	presence.Status = models.PresenceOnline
	presence.LastSeen = time.Now()
	presence.UpdatedAt = time.Now()
	
	coll := s.db.Collection("user_presence")
	
	// Upsert: update if exists, insert if not
	filter := bson.M{
		"project_id": presence.ProjectID,
		"user_id":    presence.UserID,
	}
	
	update := bson.M{
		"$set": bson.M{
			"status":         presence.Status,
			"status_message": presence.StatusMessage,
			"last_seen":      presence.LastSeen,
			"current_room":   presence.CurrentRoom,
			"device":         presence.Device,
			"ip_address":     presence.IPAddress,
			"user_agent":     presence.UserAgent,
			"updated_at":     presence.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
			"project_id": presence.ProjectID,
			"user_id":    presence.UserID,
		},
	}
	
	opts := options.Update().SetUpsert(true)
	result, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	
	if result.UpsertedID != nil {
		presence.ID = result.UpsertedID.(primitive.ObjectID)
	}
	
	return nil
}

// SetOffline marks a user as offline
func (s *PresenceService) SetOffline(ctx context.Context, projectID primitive.ObjectID, userID string) error {
	coll := s.db.Collection("user_presence")
	
	update := bson.M{
		"$set": bson.M{
			"status":       models.PresenceOffline,
			"last_seen":    time.Now(),
			"current_room": "",
			"updated_at":   time.Now(),
		},
	}
	
	_, err := coll.UpdateOne(ctx, bson.M{
		"project_id": projectID,
		"user_id":    userID,
	}, update)
	
	return err
}

// SetStatus updates user status (away, busy, etc.)
func (s *PresenceService) SetStatus(ctx context.Context, projectID primitive.ObjectID, userID string, status models.PresenceStatus, message string) error {
	coll := s.db.Collection("user_presence")
	
	update := bson.M{
		"$set": bson.M{
			"status":         status,
			"status_message": message,
			"last_seen":      time.Now(),
			"updated_at":     time.Now(),
		},
	}
	
	result, err := coll.UpdateOne(ctx, bson.M{
		"project_id": projectID,
		"user_id":    userID,
	}, update)
	
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return errors.New("user presence not found")
	}
	
	return nil
}

// GetUserStatus retrieves a user's presence status
func (s *PresenceService) GetUserStatus(ctx context.Context, projectID primitive.ObjectID, userID string) (*models.UserPresence, error) {
	coll := s.db.Collection("user_presence")
	
	var presence models.UserPresence
	err := coll.FindOne(ctx, bson.M{
		"project_id": projectID,
		"user_id":    userID,
	}).Decode(&presence)
	
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// User never online, return offline status
			return &models.UserPresence{
				ProjectID: projectID,
				UserID:    userID,
				Status:    models.PresenceOffline,
				LastSeen:  time.Time{},
			}, nil
		}
		return nil, err
	}
	
	// Check if user is stale (no activity in PresenceTTL)
	if time.Since(presence.LastSeen) > PresenceTTL {
		presence.Status = models.PresenceOffline
	}
	
	return &presence, nil
}

// GetBulkStatus retrieves presence status for multiple users
func (s *PresenceService) GetBulkStatus(ctx context.Context, projectID primitive.ObjectID, userIDs []string) (map[string]models.PresenceInfo, error) {
	coll := s.db.Collection("user_presence")
	
	cursor, err := coll.Find(ctx, bson.M{
		"project_id": projectID,
		"user_id":    bson.M{"$in": userIDs},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	result := make(map[string]models.PresenceInfo)
	
	var presences []models.UserPresence
	if err := cursor.All(ctx, &presences); err != nil {
		return nil, err
	}
	
	for _, p := range presences {
		// Check if stale
		status := p.Status
		if time.Since(p.LastSeen) > PresenceTTL {
			status = models.PresenceOffline
		}
		
		result[p.UserID] = models.PresenceInfo{
			UserID:        p.UserID,
			Status:        status,
			StatusMessage: p.StatusMessage,
			LastSeen:      p.LastSeen,
			CurrentRoom:   p.CurrentRoom,
		}
	}
	
	// Add offline status for users not found
	for _, userID := range userIDs {
		if _, exists := result[userID]; !exists {
			result[userID] = models.PresenceInfo{
				UserID:   userID,
				Status:   models.PresenceOffline,
				LastSeen: time.Time{},
			}
		}
	}
	
	return result, nil
}

// SetTyping sets a typing indicator
func (s *PresenceService) SetTyping(ctx context.Context, typing *models.TypingIndicator) error {
	typing.ExpiresAt = time.Now().Add(TypingIndicatorTTL)
	typing.CreatedAt = time.Now()
	
	coll := s.db.Collection("typing_indicators")
	
	// Upsert typing indicator
	filter := bson.M{
		"project_id": typing.ProjectID,
		"room_id":    typing.RoomID,
		"user_id":    typing.UserID,
	}
	
	update := bson.M{
		"$set": bson.M{
			"is_typing":  typing.IsTyping,
			"expires_at": typing.ExpiresAt,
			"created_at": typing.CreatedAt,
		},
	}
	
	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetTypingUsers retrieves users currently typing in a room
func (s *PresenceService) GetTypingUsers(ctx context.Context, projectID primitive.ObjectID, roomID string) ([]string, error) {
	coll := s.db.Collection("typing_indicators")
	
	// Find non-expired typing indicators
	cursor, err := coll.Find(ctx, bson.M{
		"project_id": projectID,
		"room_id":    roomID,
		"is_typing":  true,
		"expires_at": bson.M{"$gt": time.Now()},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var indicators []models.TypingIndicator
	if err := cursor.All(ctx, &indicators); err != nil {
		return nil, err
	}
	
	userIDs := make([]string, 0, len(indicators))
	for _, indicator := range indicators {
		userIDs = append(userIDs, indicator.UserID)
	}
	
	return userIDs, nil
}

// GetRoomPresence retrieves presence information for a room
func (s *PresenceService) GetRoomPresence(ctx context.Context, projectID primitive.ObjectID, roomID string) (*models.RoomPresence, error) {
	coll := s.db.Collection("user_presence")
	
	// Find users in the room
	cursor, err := coll.Find(ctx, bson.M{
		"project_id":   projectID,
		"current_room": roomID,
		"status":       bson.M{"$ne": models.PresenceOffline},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var participants []models.UserPresence
	if err := cursor.All(ctx, &participants); err != nil {
		return nil, err
	}
	
	// Filter out stale presence
	activeParticipants := make([]models.UserPresence, 0)
	for _, p := range participants {
		if time.Since(p.LastSeen) <= PresenceTTL {
			activeParticipants = append(activeParticipants, p)
		}
	}
	
	// Get typing users
	typingUsers, err := s.GetTypingUsers(ctx, projectID, roomID)
	if err != nil {
		typingUsers = []string{}
	}
	
	return &models.RoomPresence{
		RoomID:       roomID,
		Participants: activeParticipants,
		Count:        len(activeParticipants),
		TypingUsers:  typingUsers,
	}, nil
}

// UpdateActivity updates user activity
func (s *PresenceService) UpdateActivity(ctx context.Context, activity *models.UserActivity) error {
	activity.LastActivity = time.Now()
	
	coll := s.db.Collection("user_activities")
	
	// Upsert activity
	filter := bson.M{
		"project_id":    activity.ProjectID,
		"user_id":       activity.UserID,
		"resource_id":   activity.ResourceID,
		"resource_type": activity.ResourceType,
	}
	
	update := bson.M{
		"$set": bson.M{
			"activity_type": activity.ActivityType,
			"metadata":      activity.Metadata,
			"last_activity": activity.LastActivity,
		},
		"$setOnInsert": bson.M{
			"started_at": time.Now(),
		},
	}
	
	opts := options.Update().SetUpsert(true)
	result, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	
	if result.UpsertedID != nil {
		activity.ID = result.UpsertedID.(primitive.ObjectID)
	}
	
	return nil
}

// GetUserActivities retrieves recent activities for a user
func (s *PresenceService) GetUserActivities(ctx context.Context, projectID primitive.ObjectID, userID string) ([]models.UserActivity, error) {
	coll := s.db.Collection("user_activities")
	
	// Find recent activities (within ActivityTTL)
	opts := options.Find().SetSort(bson.D{{"last_activity", -1}})
	cursor, err := coll.Find(ctx, bson.M{
		"project_id":    projectID,
		"user_id":       userID,
		"last_activity": bson.M{"$gt": time.Now().Add(-ActivityTTL)},
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var activities []models.UserActivity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, err
	}
	
	return activities, nil
}

// CleanupStalePresence removes stale presence data
func (s *PresenceService) CleanupStalePresence(ctx context.Context) error {
	// Clean up old typing indicators
	typingColl := s.db.Collection("typing_indicators")
	_, err := typingColl.DeleteMany(ctx, bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	})
	if err != nil {
		return err
	}
	
	// Clean up old activities
	activityColl := s.db.Collection("user_activities")
	_, err = activityColl.DeleteMany(ctx, bson.M{
		"last_activity": bson.M{"$lt": time.Now().Add(-ActivityTTL)},
	})
	
	return err
}

// RunCleanupLoop runs a background cleanup loop
func (s *PresenceService) RunCleanupLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.CleanupStalePresence(ctx); err != nil {
				// Log error but continue
				continue
			}
		}
	}
}
