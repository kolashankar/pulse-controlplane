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

// SLAService handles SLA operations
type SLAService struct {
	db *mongo.Database
}

// NewSLAService creates a new SLA service
func NewSLAService(db *mongo.Database) *SLAService {
	return &SLAService{db: db}
}

// CreateSLATemplate creates a new SLA template
func (s *SLAService) CreateSLATemplate(ctx context.Context, template *models.SLATemplate) error {
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	template.IsActive = true
	
	coll := s.db.Collection("sla_templates")
	result, err := coll.InsertOne(ctx, template)
	if err != nil {
		return err
	}
	
	template.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetSLATemplates lists all SLA templates
func (s *SLAService) GetSLATemplates(ctx context.Context, activeOnly bool) ([]models.SLATemplate, error) {
	coll := s.db.Collection("sla_templates")
	
	filter := bson.M{}
	if activeOnly {
		filter["is_active"] = true
	}
	
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var templates []models.SLATemplate
	if err := cursor.All(ctx, &templates); err != nil {
		return nil, err
	}
	
	return templates, nil
}

// AssignSLAToOrg assigns an SLA to an organization
func (s *SLAService) AssignSLAToOrg(ctx context.Context, orgSLA *models.OrganizationSLA) error {
	orgSLA.CreatedAt = time.Now()
	orgSLA.IsActive = true
	orgSLA.StartDate = time.Now()
	
	coll := s.db.Collection("organization_slas")
	
	// Deactivate any existing active SLA for this org
	_, err := coll.UpdateMany(ctx, 
		bson.M{"org_id": orgSLA.OrgID, "is_active": true},
		bson.M{"$set": bson.M{"is_active": false, "end_date": time.Now()}},
	)
	if err != nil {
		return err
	}
	
	result, err := coll.InsertOne(ctx, orgSLA)
	if err != nil {
		return err
	}
	
	orgSLA.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetOrganizationSLA gets active SLA for an organization
func (s *SLAService) GetOrganizationSLA(ctx context.Context, orgID primitive.ObjectID) (*models.OrganizationSLA, error) {
	coll := s.db.Collection("organization_slas")
	
	var orgSLA models.OrganizationSLA
	err := coll.FindOne(ctx, bson.M{"org_id": orgID, "is_active": true}).Decode(&orgSLA)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("no active SLA found for organization")
		}
		return nil, err
	}
	
	return &orgSLA, nil
}

// TrackSLAMetrics records SLA performance metrics
func (s *SLAService) TrackSLAMetrics(ctx context.Context, metric *models.SLAMetric) error {
	metric.CreatedAt = time.Now()
	
	// Get the SLA template to check compliance
	orgSLA, err := s.GetOrganizationSLA(ctx, metric.OrgID)
	if err != nil {
		return err
	}
	
	var template models.SLATemplate
	templColl := s.db.Collection("sla_templates")
	err = templColl.FindOne(ctx, bson.M{"_id": orgSLA.SLATemplateID}).Decode(&template)
	if err != nil {
		return err
	}
	
	// Check SLA compliance
	metric.UptimeMet = metric.ActualUptime >= template.UptimePercent
	metric.ResponseTimeMet = metric.ActualResponseP95 <= template.APIResponseTime
	metric.OverallCompliance = metric.UptimeMet && metric.ResponseTimeMet
	
	// Calculate credit if SLA breached
	if !metric.OverallCompliance {
		metric.CreditEarned = template.CreditPercent
	}
	
	coll := s.db.Collection("sla_metrics")
	result, err := coll.InsertOne(ctx, metric)
	if err != nil {
		return err
	}
	
	metric.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// RecordSLABreach records an SLA breach event
func (s *SLAService) RecordSLABreach(ctx context.Context, breach *models.SLABreach) error {
	breach.CreatedAt = time.Now()
	breach.Resolved = false
	
	coll := s.db.Collection("sla_breaches")
	result, err := coll.InsertOne(ctx, breach)
	if err != nil {
		return err
	}
	
	breach.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetSLAReport generates SLA report for a period
func (s *SLAService) GetSLAReport(ctx context.Context, orgID primitive.ObjectID, periodStart, periodEnd time.Time) (*models.SLAMetric, error) {
	coll := s.db.Collection("sla_metrics")
	
	var metrics []models.SLAMetric
	cursor, err := coll.Find(ctx, bson.M{
		"org_id": orgID,
		"period_start": bson.M{"$gte": periodStart},
		"period_end": bson.M{"$lte": periodEnd},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	if err := cursor.All(ctx, &metrics); err != nil {
		return nil, err
	}
	
	if len(metrics) == 0 {
		return nil, errors.New("no metrics found for the period")
	}
	
	// Aggregate metrics
	report := &models.SLAMetric{
		OrgID:       orgID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		CreatedAt:   time.Now(),
	}
	
	var totalUptime, totalResponseTime float64
	var breaches, downtime int
	var allCompliant = true
	
	for _, m := range metrics {
		totalUptime += m.ActualUptime
		totalResponseTime += float64(m.ActualResponseP95)
		breaches += m.BreachCount
		downtime += m.DowntimeMinutes
		if !m.OverallCompliance {
			allCompliant = false
		}
	}
	
	count := float64(len(metrics))
	report.ActualUptime = totalUptime / count
	report.ActualResponseP95 = int(totalResponseTime / count)
	report.BreachCount = breaches
	report.DowntimeMinutes = downtime
	report.OverallCompliance = allCompliant
	
	return report, nil
}

// GetSLABreaches gets SLA breaches for an organization
func (s *SLAService) GetSLABreaches(ctx context.Context, orgID primitive.ObjectID, limit int) ([]models.SLABreach, error) {
	coll := s.db.Collection("sla_breaches")
	
	opts := options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(int64(limit))
	cursor, err := coll.Find(ctx, bson.M{"org_id": orgID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var breaches []models.SLABreach
	if err := cursor.All(ctx, &breaches); err != nil {
		return nil, err
	}
	
	return breaches, nil
}
