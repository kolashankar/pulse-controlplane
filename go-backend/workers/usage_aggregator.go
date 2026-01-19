package workers

import (
	"context"
	"time"

	"pulse-control-plane/services"

	"github.com/rs/zerolog/log"
)

// UsageAggregatorWorker handles periodic usage aggregation
type UsageAggregatorWorker struct {
	aggregatorService *services.AggregatorService
	stopChan          chan struct{}
	doneChan          chan struct{}
}

// NewUsageAggregatorWorker creates a new usage aggregator worker
func NewUsageAggregatorWorker(aggregatorService *services.AggregatorService) *UsageAggregatorWorker {
	return &UsageAggregatorWorker{
		aggregatorService: aggregatorService,
		stopChan:          make(chan struct{}),
		doneChan:          make(chan struct{}),
	}
}

// Start starts the background worker
func (w *UsageAggregatorWorker) Start() {
	log.Info().Msg("Starting usage aggregator worker")

	// Run immediately on startup
	go w.runAggregation()

	// Setup tickers for periodic aggregation
	hourlyTicker := time.NewTicker(1 * time.Hour)
	dailyTicker := time.NewTicker(24 * time.Hour)

	go func() {
		defer close(w.doneChan)
		defer hourlyTicker.Stop()
		defer dailyTicker.Stop()

		for {
			select {
			case <-hourlyTicker.C:
				log.Info().Msg("Hourly aggregation triggered")
				w.aggregateHourly()

			case <-dailyTicker.C:
				log.Info().Msg("Daily aggregation triggered")
				w.aggregateDaily()
				// Also run monthly aggregation on the 1st of each month
				if time.Now().Day() == 1 {
					log.Info().Msg("Monthly aggregation triggered")
					w.aggregateMonthly()
				}

			case <-w.stopChan:
				log.Info().Msg("Stopping usage aggregator worker")
				return
			}
		}
	}()

	log.Info().Msg("Usage aggregator worker started successfully")
}

// Stop stops the background worker
func (w *UsageAggregatorWorker) Stop() {
	log.Info().Msg("Stopping usage aggregator worker...")
	close(w.stopChan)
	<-w.doneChan
	log.Info().Msg("Usage aggregator worker stopped")
}

// runAggregation runs initial aggregation on startup
func (w *UsageAggregatorWorker) runAggregation() {
	log.Info().Msg("Running initial usage aggregation")
	
	// Run hourly aggregation for any missing hours
	w.aggregateHourly()
	
	// Check if we need to run daily aggregation
	now := time.Now()
	if now.Hour() == 0 {
		w.aggregateDaily()
	}
	
	// Check if we need to run monthly aggregation
	if now.Day() == 1 && now.Hour() == 0 {
		w.aggregateMonthly()
	}
}

// aggregateHourly performs hourly aggregation
func (w *UsageAggregatorWorker) aggregateHourly() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	log.Info().Msg("Starting hourly usage aggregation")
	
	if err := w.aggregatorService.AggregateHourlyUsage(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to aggregate hourly usage")
		return
	}
	
	log.Info().Msg("Hourly usage aggregation completed successfully")
}

// aggregateDaily performs daily aggregation
func (w *UsageAggregatorWorker) aggregateDaily() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	log.Info().Msg("Starting daily usage aggregation")
	
	if err := w.aggregatorService.AggregateDailyUsage(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to aggregate daily usage")
		return
	}
	
	log.Info().Msg("Daily usage aggregation completed successfully")
}

// aggregateMonthly performs monthly aggregation
func (w *UsageAggregatorWorker) aggregateMonthly() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	log.Info().Msg("Starting monthly usage aggregation")
	
	if err := w.aggregatorService.AggregateMonthlyUsage(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to aggregate monthly usage")
		return
	}
	
	log.Info().Msg("Monthly usage aggregation completed successfully")
}

// GetStatus returns the worker status
func (w *UsageAggregatorWorker) GetStatus() string {
	select {
	case <-w.stopChan:
		return "stopped"
	default:
		return "running"
	}
}
