package queue

import (
	"context"
	"sync"
	"time"

	"pulse-control-plane/models"
	"pulse-control-plane/services"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RetryQueue handles webhook retry logic
type RetryQueue struct {
	webhookService *services.WebhookService
	retryMap       map[primitive.ObjectID]*time.Timer
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewRetryQueue creates a new retry queue
func NewRetryQueue() *RetryQueue {
	ctx, cancel := context.WithCancel(context.Background())
	return &RetryQueue{
		webhookService: services.NewWebhookService(),
		retryMap:       make(map[primitive.ObjectID]*time.Timer),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// ScheduleRetry schedules a webhook for retry
func (q *RetryQueue) ScheduleRetry(webhookLogID primitive.ObjectID, retryAt time.Time) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Cancel existing timer if any
	if timer, exists := q.retryMap[webhookLogID]; exists {
		timer.Stop()
	}

	// Calculate delay
	delay := time.Until(retryAt)
	if delay < 0 {
		delay = 0
	}

	// Create new timer
	timer := time.AfterFunc(delay, func() {
		q.executeRetry(webhookLogID)
	})

	q.retryMap[webhookLogID] = timer

	log.Info().
		Str("webhook_log_id", webhookLogID.Hex()).
		Dur("delay", delay).
		Msg("Webhook retry scheduled")
}

// executeRetry executes a webhook retry
func (q *RetryQueue) executeRetry(webhookLogID primitive.ObjectID) {
	log.Info().Str("webhook_log_id", webhookLogID.Hex()).Msg("Executing webhook retry")

	// Remove from retry map
	q.mu.Lock()
	delete(q.retryMap, webhookLogID)
	q.mu.Unlock()

	// Note: In a production implementation, you would:
	// 1. Fetch the webhook log from database
	// 2. Re-attempt delivery using webhookService
	// 3. Update the webhook log with the result
	// 4. Schedule another retry if needed

	// This is a placeholder for the actual retry logic
	log.Info().Str("webhook_log_id", webhookLogID.Hex()).Msg("Webhook retry completed")
}

// CancelRetry cancels a scheduled retry
func (q *RetryQueue) CancelRetry(webhookLogID primitive.ObjectID) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if timer, exists := q.retryMap[webhookLogID]; exists {
		timer.Stop()
		delete(q.retryMap, webhookLogID)
		log.Info().Str("webhook_log_id", webhookLogID.Hex()).Msg("Webhook retry cancelled")
	}
}

// Start starts the retry queue background worker
func (q *RetryQueue) Start() {
	log.Info().Msg("Starting retry queue worker")

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-q.ctx.Done():
				log.Info().Msg("Retry queue worker stopped")
				return
			case <-ticker.C:
				q.processRetries()
			}
		}
	}()
}

// Stop stops the retry queue
func (q *RetryQueue) Stop() {
	log.Info().Msg("Stopping retry queue")
	q.cancel()

	// Cancel all pending retries
	q.mu.Lock()
	for id, timer := range q.retryMap {
		timer.Stop()
		delete(q.retryMap, id)
	}
	q.mu.Unlock()
}

// processRetries processes pending retries
func (q *RetryQueue) processRetries() {
	log.Debug().Msg("Processing pending webhook retries")

	// Note: In a production implementation, you would:
	// 1. Query database for webhook logs with status=retrying and next_retry_at <= now
	// 2. For each log, schedule a retry
	// 3. Update the retry map

	// This is a placeholder for the actual retry processing logic
}

// GetStats returns queue statistics
func (q *RetryQueue) GetStats() map[string]interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return map[string]interface{}{
		"pending_retries": len(q.retryMap),
		"queue_running":   q.ctx.Err() == nil,
	}
}
