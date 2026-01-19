package database

import (
	"context"
	"fmt"
	"time"

	"pulse-control-plane/config"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client   *mongo.Client
	Database *mongo.Database
)

// ConnectMongoDB establishes connection to MongoDB
func ConnectMongoDB(cfg *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create client options
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	Client = client
	Database = client.Database(cfg.MongoDBName)

	log.Info().Str("database", cfg.MongoDBName).Msg("Connected to MongoDB successfully")

	// Create indexes
	if err := createIndexes(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to create indexes")
		return err
	}

	return nil
}

// DisconnectMongoDB closes the MongoDB connection
func DisconnectMongoDB() error {
	if Client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	log.Info().Msg("Disconnected from MongoDB")
	return nil
}

// createIndexes creates all necessary MongoDB indexes
func createIndexes(ctx context.Context) error {
	log.Info().Msg("Creating MongoDB indexes...")

	// Organizations indexes
	orgCollection := Database.Collection("organizations")
	orgIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "admin_email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "is_deleted", Value: 1}},
		},
	}
	if _, err := orgCollection.Indexes().CreateMany(ctx, orgIndexes); err != nil {
		return fmt.Errorf("failed to create organization indexes: %w", err)
	}

	// Projects indexes
	projectCollection := Database.Collection("projects")
	projectIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "pulse_api_key", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "org_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_deleted", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "org_id", Value: 1}, {Key: "name", Value: 1}},
		},
	}
	if _, err := projectCollection.Indexes().CreateMany(ctx, projectIndexes); err != nil {
		return fmt.Errorf("failed to create project indexes: %w", err)
	}

	// Users indexes
	userCollection := Database.Collection("users")
	userIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "org_id", Value: 1}},
		},
	}
	if _, err := userCollection.Indexes().CreateMany(ctx, userIndexes); err != nil {
		return fmt.Errorf("failed to create user indexes: %w", err)
	}

	// Usage metrics indexes with TTL (90 days)
	usageCollection := Database.Collection("usage_metrics")
	usageIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "project_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "project_id", Value: 1}, {Key: "timestamp", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "event_type", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "timestamp", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(7776000), // 90 days TTL
		},
	}
	if _, err := usageCollection.Indexes().CreateMany(ctx, usageIndexes); err != nil {
		return fmt.Errorf("failed to create usage metrics indexes: %w", err)
	}

	log.Info().Msg("MongoDB indexes created successfully")
	return nil
}

// GetCollection returns a MongoDB collection
func GetCollection(name string) *mongo.Collection {
	return Database.Collection(name)
}
