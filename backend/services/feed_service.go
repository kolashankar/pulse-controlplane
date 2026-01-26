package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pulse-control-plane/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// Fan-out threshold: users with more followers use fan-out on read
	FanOutThreshold = 10000
)

// FeedService handles activity feed operations
type FeedService struct {
	db *mongo.Database
}

// NewFeedService creates a new feed service
func NewFeedService(db *mongo.Database) *FeedService {
	return &FeedService{db: db}
}

// CreateFeed creates a new feed
func (s *FeedService) CreateFeed(ctx context.Context, feed *models.Feed) error {
	feed.CreatedAt = time.Now()
	feed.UpdatedAt = time.Now()
	
	// Default settings
	if feed.Settings.MaxItems == 0 {
		feed.Settings.MaxItems = 1000
	}
	if feed.Settings.Ranking == "" {
		feed.Settings.Ranking = "chronological"
	}
	
	coll := s.db.Collection("feeds")
	result, err := coll.InsertOne(ctx, feed)
	if err != nil {
		return err
	}
	
	feed.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetFeed retrieves a feed by user ID
func (s *FeedService) GetFeed(ctx context.Context, projectID primitive.ObjectID, userID string, feedType models.FeedType) (*models.Feed, error) {
	coll := s.db.Collection("feeds")
	
	var feed models.Feed
	err := coll.FindOne(ctx, bson.M{
		"project_id": projectID,
		"user_id":    userID,
		"feed_type":  feedType,
	}).Decode(&feed)
	
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("feed not found")
		}
		return nil, err
	}
	
	return &feed, nil
}

// CreateActivity creates a new activity and fans it out to followers
func (s *FeedService) CreateActivity(ctx context.Context, activity *models.Activity) error {
	activity.CreatedAt = time.Now()
	if activity.Time.IsZero() {
		activity.Time = time.Now()
	}
	
	// Insert activity
	coll := s.db.Collection("activities")
	result, err := coll.InsertOne(ctx, activity)
	if err != nil {
		return err
	}
	activity.ID = result.InsertedID.(primitive.ObjectID)
	
	// Get follower count to decide fan-out strategy
	followerCount, err := s.GetFollowerCount(ctx, activity.ProjectID, activity.Actor)
	if err != nil {
		followerCount = 0
	}
	
	// Fan-out strategy
	if followerCount < FanOutThreshold {
		// Fan-out on write: add to all followers' feeds
		return s.fanOutOnWrite(ctx, activity)
	} else {
		// Fan-out on read: activity will be fetched dynamically
		// Just store the activity, no need to fan out
		return nil
	}
}

// fanOutOnWrite adds the activity to all followers' feeds
func (s *FeedService) fanOutOnWrite(ctx context.Context, activity *models.Activity) error {
	// Get all followers
	followers, err := s.GetFollowers(ctx, activity.ProjectID, activity.Actor, 1, 10000)
	if err != nil {
		return err
	}
	
	// Create feed items for each follower
	if len(followers) == 0 {
		return nil
	}
	
	feedItems := make([]interface{}, 0, len(followers))
	for _, follower := range followers {
		feedItem := models.FeedItem{
			ActivityID: activity.ID,
			ProjectID:  activity.ProjectID,
			UserID:     follower.Follower,
			Activity:   *activity,
			Score:      float64(activity.Time.Unix()),
			Seen:       false,
			Read:       false,
			CreatedAt:  time.Now(),
		}
		feedItems = append(feedItems, feedItem)
	}
	
	coll := s.db.Collection("feed_items")
	_, err = coll.InsertMany(ctx, feedItems)
	return err
}

// GetFeedItems retrieves feed items for a user
func (s *FeedService) GetFeedItems(ctx context.Context, projectID primitive.ObjectID, userID string, page, limit int) (*models.FeedResponse, error) {
	coll := s.db.Collection("feed_items")
	
	filter := bson.M{
		"project_id": projectID,
		"user_id":    userID,
	}
	
	// Count total
	totalCount, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	// Find with pagination
	opts := options.Find().
		SetSort(bson.D{{"score", -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))
	
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var items []models.FeedItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	
	hasMore := totalCount > int64(page*limit)
	
	return &models.FeedResponse{
		Items:   items,
		Total:   totalCount,
		Page:    page,
		Limit:   limit,
		HasMore: hasMore,
	}, nil
}

// GetAggregatedFeed returns aggregated activities (grouped by verb)
func (s *FeedService) GetAggregatedFeed(ctx context.Context, projectID primitive.ObjectID, userID string, page, limit int) ([]models.AggregatedActivity, error) {
	coll := s.db.Collection("feed_items")
	
	// Aggregation pipeline
	pipeline := mongo.Pipeline{
		{{
			"$match", bson.D{
				{"project_id", projectID},
				{"user_id", userID},
			},
		}},
		{{
			"$sort", bson.D{{"created_at", -1}},
		}},
		{{
			"$skip", int64((page - 1) * limit),
		}},
		{{
			"$limit", int64(limit),
		}},
		{{
			"$group", bson.D{
				{"_id", "$activity.verb"},
				{"actors", bson.D{{"$addToSet", "$activity.actor"}}},
				{"count", bson.D{{"$sum", 1}}},
				{"first_activity", bson.D{{"$first", "$activity"}}},
				{"latest_activity", bson.D{{"$last", "$activity"}}},
				{"created_at", bson.D{{"$first", "$created_at"}}},
				{"updated_at", bson.D{{"$last", "$created_at"}}},
			},
		}},
	}
	
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var results []struct {
		ID             models.ActivityType `bson:"_id"`
		Actors         []string            `bson:"actors"`
		Count          int                 `bson:"count"`
		FirstActivity  models.Activity     `bson:"first_activity"`
		LatestActivity models.Activity     `bson:"latest_activity"`
		CreatedAt      time.Time           `bson:"created_at"`
		UpdatedAt      time.Time           `bson:"updated_at"`
	}
	
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	
	aggregated := make([]models.AggregatedActivity, len(results))
	for i, r := range results {
		aggregated[i] = models.AggregatedActivity{
			ID:             fmt.Sprintf("%s-%d", r.ID, r.CreatedAt.Unix()),
			Verb:           r.ID,
			Actors:         r.Actors,
			ActivityCount:  r.Count,
			FirstActivity:  r.FirstActivity,
			LatestActivity: r.LatestActivity,
			CreatedAt:      r.CreatedAt,
			UpdatedAt:      r.UpdatedAt,
		}
	}
	
	return aggregated, nil
}

// Follow creates a follow relationship
func (s *FeedService) Follow(ctx context.Context, follow *models.Follow) error {
	// Check if already following
	coll := s.db.Collection("follows")
	count, err := coll.CountDocuments(ctx, bson.M{
		"project_id": follow.ProjectID,
		"follower":   follow.Follower,
		"following":  follow.Following,
	})
	if err != nil {
		return err
	}
	
	if count > 0 {
		return errors.New("already following")
	}
	
	follow.CreatedAt = time.Now()
	result, err := coll.InsertOne(ctx, follow)
	if err != nil {
		return err
	}
	
	follow.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// Unfollow removes a follow relationship
func (s *FeedService) Unfollow(ctx context.Context, projectID primitive.ObjectID, follower, following string) error {
	coll := s.db.Collection("follows")
	result, err := coll.DeleteOne(ctx, bson.M{
		"project_id": projectID,
		"follower":   follower,
		"following":  following,
	})
	if err != nil {
		return err
	}
	
	if result.DeletedCount == 0 {
		return errors.New("not following")
	}
	
	return nil
}

// GetFollowers retrieves followers of a user
func (s *FeedService) GetFollowers(ctx context.Context, projectID primitive.ObjectID, userID string, page, limit int) ([]models.Follow, error) {
	coll := s.db.Collection("follows")
	
	opts := options.Find().
		SetSort(bson.D{{"created_at", -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))
	
	cursor, err := coll.Find(ctx, bson.M{
		"project_id": projectID,
		"following":  userID,
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var followers []models.Follow
	if err := cursor.All(ctx, &followers); err != nil {
		return nil, err
	}
	
	return followers, nil
}

// GetFollowing retrieves users that a user is following
func (s *FeedService) GetFollowing(ctx context.Context, projectID primitive.ObjectID, userID string, page, limit int) ([]models.Follow, error) {
	coll := s.db.Collection("follows")
	
	opts := options.Find().
		SetSort(bson.D{{"created_at", -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))
	
	cursor, err := coll.Find(ctx, bson.M{
		"project_id": projectID,
		"follower":   userID,
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var following []models.Follow
	if err := cursor.All(ctx, &following); err != nil {
		return nil, err
	}
	
	return following, nil
}

// GetFollowerCount returns the count of followers
func (s *FeedService) GetFollowerCount(ctx context.Context, projectID primitive.ObjectID, userID string) (int64, error) {
	coll := s.db.Collection("follows")
	return coll.CountDocuments(ctx, bson.M{
		"project_id": projectID,
		"following":  userID,
	})
}

// GetFollowingCount returns the count of users being followed
func (s *FeedService) GetFollowingCount(ctx context.Context, projectID primitive.ObjectID, userID string) (int64, error) {
	coll := s.db.Collection("follows")
	return coll.CountDocuments(ctx, bson.M{
		"project_id": projectID,
		"follower":   userID,
	})
}

// GetFollowStats returns follower/following statistics
func (s *FeedService) GetFollowStats(ctx context.Context, projectID primitive.ObjectID, userID string) (*models.FollowStats, error) {
	followersCount, err := s.GetFollowerCount(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	
	followingCount, err := s.GetFollowingCount(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	
	return &models.FollowStats{
		UserID:         userID,
		FollowersCount: followersCount,
		FollowingCount: followingCount,
	}, nil
}

// MarkAsSeen marks feed items as seen
func (s *FeedService) MarkAsSeen(ctx context.Context, projectID primitive.ObjectID, userID string, itemIDs []primitive.ObjectID) error {
	coll := s.db.Collection("feed_items")
	
	_, err := coll.UpdateMany(ctx,
		bson.M{
			"project_id": projectID,
			"user_id":    userID,
			"_id":        bson.M{"$in": itemIDs},
		},
		bson.M{"$set": bson.M{"seen": true}},
	)
	
	return err
}

// MarkAsRead marks feed items as read
func (s *FeedService) MarkAsRead(ctx context.Context, projectID primitive.ObjectID, userID string, itemIDs []primitive.ObjectID) error {
	coll := s.db.Collection("feed_items")
	
	_, err := coll.UpdateMany(ctx,
		bson.M{
			"project_id": projectID,
			"user_id":    userID,
			"_id":        bson.M{"$in": itemIDs},
		},
		bson.M{"$set": bson.M{"read": true, "seen": true}},
	)
	
	return err
}

// DeleteActivity removes an activity and its feed items
func (s *FeedService) DeleteActivity(ctx context.Context, activityID primitive.ObjectID) error {
	// Delete activity
	activityColl := s.db.Collection("activities")
	_, err := activityColl.DeleteOne(ctx, bson.M{"_id": activityID})
	if err != nil {
		return err
	}
	
	// Delete feed items
	feedItemColl := s.db.Collection("feed_items")
	_, err = feedItemColl.DeleteMany(ctx, bson.M{"activity_id": activityID})
	return err
}
