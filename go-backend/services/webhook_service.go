package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"pulse-control-plane/database"
	"pulse-control-plane/models"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// WebhookService handles webhook delivery
type WebhookService struct {
	collection *mongo.Collection
	httpClient *http.Client
	webhookSecret string
}

// NewWebhookService creates a new webhook service
func NewWebhookService() *WebhookService {
	return &WebhookService{
		collection: database.GetCollection("webhook_logs"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		webhookSecret: os.Getenv("WEBHOOK_SECRET"),
	}
}

// SendWebhook sends a webhook to the customer's endpoint
func (s *WebhookService) SendWebhook(ctx context.Context, project *models.Project, payload *models.WebhookPayload) error {
	if project.WebhookURL == "" {
		log.Debug().Str("project_id", project.ID.Hex()).Msg("No webhook URL configured, skipping")
		return nil
	}
	
	// Create webhook log
	webhookLog := &models.WebhookLog{
		ID: primitive.NewObjectID(),
		ProjectID: project.ID,
		EventType: payload.Event,
		Payload: convertToMap(payload),
		WebhookURL: project.WebhookURL,
		Status: models.WebhookStatusPending,
		Attempts: 0,
		MaxAttempts: 5, // Max 5 attempts (initial + 4 retries)
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Save webhook log
	_, err := s.collection.InsertOne(ctx, webhookLog)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create webhook log")
		return err
	}
	
	// Send webhook (async in background)
	go s.deliverWebhook(webhookLog)
	
	return nil
}

// deliverWebhook attempts to deliver a webhook
func (s *WebhookService) deliverWebhook(webhookLog *models.WebhookLog) {
	ctx := context.Background()
	
	// Prepare payload
	payloadBytes, err := json.Marshal(webhookLog.Payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal webhook payload")
		s.updateWebhookLog(ctx, webhookLog.ID, models.WebhookStatusFailed, 0, "", err.Error())
		return
	}
	
	// Generate HMAC signature
	signature := s.generateSignature(payloadBytes)
	webhookLog.Signature = signature
	
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", webhookLog.WebhookURL, bytes.NewReader(payloadBytes))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create webhook request")
		s.updateWebhookLog(ctx, webhookLog.ID, models.WebhookStatusFailed, 0, "", err.Error())
		return
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Pulse-Signature", signature)
	req.Header.Set("X-Pulse-Event", string(webhookLog.EventType))
	req.Header.Set("User-Agent", "Pulse-Webhook/1.0")
	
	// Attempt delivery
	for attempt := 1; attempt <= webhookLog.MaxAttempts; attempt++ {
		log.Info().Int("attempt", attempt).Str("url", webhookLog.WebhookURL).Msg("Attempting webhook delivery")
		
		resp, err := s.httpClient.Do(req)
		now := time.Now()
		
		if err != nil {
			log.Error().Err(err).Int("attempt", attempt).Msg("Webhook delivery failed")
			
			if attempt < webhookLog.MaxAttempts {
				// Calculate retry delay (exponential backoff: 5m, 10m, 30m)
				retryDelay := s.calculateRetryDelay(attempt)
				nextRetry := now.Add(retryDelay)
				
				s.updateWebhookLog(ctx, webhookLog.ID, models.WebhookStatusRetrying, attempt, "", err.Error())
				s.scheduleRetry(ctx, webhookLog.ID, &nextRetry)
				
				log.Info().Dur("retry_in", retryDelay).Msg("Webhook retry scheduled")
				time.Sleep(retryDelay)
				continue
			} else {
				s.updateWebhookLog(ctx, webhookLog.ID, models.WebhookStatusFailed, attempt, "", err.Error())
				return
			}
		}
		
		// Read response body
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		// Check response status
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Info().Int("status", resp.StatusCode).Msg("Webhook delivered successfully")
			s.updateWebhookLog(ctx, webhookLog.ID, models.WebhookStatusDelivered, attempt, string(body), "")
			return
		}
		
		// Non-2xx status code
		log.Warn().Int("status", resp.StatusCode).Str("body", string(body)).Msg("Webhook delivery failed with non-2xx status")
		
		if attempt < webhookLog.MaxAttempts {
			retryDelay := s.calculateRetryDelay(attempt)
			nextRetry := now.Add(retryDelay)
			
			s.updateWebhookLog(ctx, webhookLog.ID, models.WebhookStatusRetrying, attempt, string(body), fmt.Sprintf("HTTP %d", resp.StatusCode))
			s.scheduleRetry(ctx, webhookLog.ID, &nextRetry)
			
			time.Sleep(retryDelay)
			continue
		} else {
			s.updateWebhookLog(ctx, webhookLog.ID, models.WebhookStatusFailed, attempt, string(body), fmt.Sprintf("HTTP %d", resp.StatusCode))
			return
		}
	}
}

// calculateRetryDelay calculates the retry delay based on attempt number
// 5 minutes, 10 minutes, 30 minutes
func (s *WebhookService) calculateRetryDelay(attempt int) time.Duration {
	switch attempt {
	case 1:
		return 5 * time.Minute
	case 2:
		return 10 * time.Minute
	case 3:
		return 30 * time.Minute
	default:
		return 30 * time.Minute
	}
}

// scheduleRetry updates the next retry time
func (s *WebhookService) scheduleRetry(ctx context.Context, webhookLogID primitive.ObjectID, nextRetry *time.Time) {
	update := bson.M{
		"$set": bson.M{
			"next_retry_at": nextRetry,
		},
	}
	s.collection.UpdateOne(ctx, bson.M{"_id": webhookLogID}, update)
}

// updateWebhookLog updates the webhook log status
func (s *WebhookService) updateWebhookLog(ctx context.Context, webhookLogID primitive.ObjectID, status models.WebhookDeliveryStatus, attempts int, responseBody, errorMsg string) {
	update := bson.M{
		"$set": bson.M{
			"status": status,
			"attempts": attempts,
			"last_attempt_at": time.Now(),
			"updated_at": time.Now(),
		},
	}
	
	if responseBody != "" {
		update["$set"].(bson.M)["response_body"] = responseBody
	}
	
	if errorMsg != "" {
		update["$set"].(bson.M)["error"] = errorMsg
	}
	
	s.collection.UpdateOne(ctx, bson.M{"_id": webhookLogID}, update)
}

// generateSignature generates HMAC signature for webhook payload
func (s *WebhookService) generateSignature(payload []byte) string {
	if s.webhookSecret == "" {
		// Use project-specific secret in production
		s.webhookSecret = "default-webhook-secret"
	}
	
	h := hmac.New(sha256.New, []byte(s.webhookSecret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature verifies the HMAC signature of an incoming webhook
func (s *WebhookService) VerifySignature(payload []byte, signature string) bool {
	expectedSignature := s.generateSignature(payload)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// GetWebhookLogs retrieves webhook logs for a project
func (s *WebhookService) GetWebhookLogs(ctx context.Context, projectID primitive.ObjectID, page, limit int) ([]models.WebhookLog, int64, error) {
	filter := bson.M{"project_id": projectID}
	
	// Count total
	total, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count webhook logs: %w", err)
	}
	
	// Calculate skip
	skip := (page - 1) * limit
	
	// Find with pagination
	cursor, err := s.collection.Find(ctx, filter, 
		&mongo.Options{
			Skip: &skip,
			Limit: &limit,
			Sort: bson.M{"created_at": -1},
		},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list webhook logs: %w", err)
	}
	defer cursor.Close(ctx)
	
	var logs []models.WebhookLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, 0, fmt.Errorf("failed to decode webhook logs: %w", err)
	}
	
	return logs, total, nil
}

// convertToMap converts a struct to map[string]interface{}
func convertToMap(v interface{}) map[string]interface{} {
	data, _ := json.Marshal(v)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}
