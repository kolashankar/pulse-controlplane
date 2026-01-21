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

// SupportService handles support ticket operations
type SupportService struct {
	db *mongo.Database
}

// NewSupportService creates a new support service
func NewSupportService(db *mongo.Database) *SupportService {
	return &SupportService{db: db}
}

// CreateTicket creates a new support ticket
func (s *SupportService) CreateTicket(ctx context.Context, ticket *models.SupportTicket) error {
	// Generate ticket number
	ticketNum, err := s.generateTicketNumber(ctx)
	if err != nil {
		return err
	}
	ticket.TicketNumber = ticketNum
	
	ticket.Status = models.TicketStatusOpen
	ticket.CreatedAt = time.Now()
	ticket.UpdatedAt = time.Now()
	
	coll := s.db.Collection("support_tickets")
	result, err := coll.InsertOne(ctx, ticket)
	if err != nil {
		return err
	}
	
	ticket.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// generateTicketNumber generates a unique ticket number
func (s *SupportService) generateTicketNumber(ctx context.Context) (string, error) {
	coll := s.db.Collection("support_tickets")
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("PULSE-%05d", count+1), nil
}

// GetTicket retrieves a ticket by ID
func (s *SupportService) GetTicket(ctx context.Context, ticketID primitive.ObjectID) (*models.SupportTicket, error) {
	coll := s.db.Collection("support_tickets")
	
	var ticket models.SupportTicket
	err := coll.FindOne(ctx, bson.M{"_id": ticketID}).Decode(&ticket)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("ticket not found")
		}
		return nil, err
	}
	
	return &ticket, nil
}

// ListTickets lists tickets with filters
func (s *SupportService) ListTickets(ctx context.Context, orgID *primitive.ObjectID, status models.TicketStatus, priority models.TicketPriority, page, limit int) ([]models.SupportTicket, int64, error) {
	coll := s.db.Collection("support_tickets")
	
	filter := bson.M{}
	if orgID != nil {
		filter["org_id"] = *orgID
	}
	if status != "" {
		filter["status"] = status
	}
	if priority != "" {
		filter["priority"] = priority
	}
	
	// Count total
	totalCount, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	
	// Find with pagination
	opts := options.Find().
		SetSort(bson.D{{"created_at", -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))
	
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	
	var tickets []models.SupportTicket
	if err := cursor.All(ctx, &tickets); err != nil {
		return nil, 0, err
	}
	
	return tickets, totalCount, nil
}

// UpdateTicket updates ticket fields
func (s *SupportService) UpdateTicket(ctx context.Context, ticketID primitive.ObjectID, updates bson.M) error {
	updates["updated_at"] = time.Now()
	
	// Track status changes
	if status, ok := updates["status"].(models.TicketStatus); ok {
		if status == models.TicketStatusResolved {
			updates["resolved_at"] = time.Now()
			// Calculate resolution time
			ticket, err := s.GetTicket(ctx, ticketID)
			if err == nil {
				resolutionTime := int(time.Since(ticket.CreatedAt).Minutes())
				updates["resolution_time"] = resolutionTime
			}
		} else if status == models.TicketStatusClosed {
			updates["closed_at"] = time.Now()
		}
	}
	
	coll := s.db.Collection("support_tickets")
	result, err := coll.UpdateOne(ctx, bson.M{"_id": ticketID}, bson.M{"$set": updates})
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return errors.New("ticket not found")
	}
	
	return nil
}

// AssignTicket assigns a ticket to an agent
func (s *SupportService) AssignTicket(ctx context.Context, ticketID, agentID primitive.ObjectID) error {
	updates := bson.M{
		"assigned_to": agentID,
		"status":      models.TicketStatusInProgress,
		"updated_at":  time.Now(),
	}
	
	// Set first response time if not set
	ticket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return err
	}
	
	if ticket.FirstResponse == nil {
		now := time.Now()
		updates["first_response"] = now
		responseTime := int(now.Sub(ticket.CreatedAt).Minutes())
		updates["response_time"] = responseTime
	}
	
	return s.UpdateTicket(ctx, ticketID, updates)
}

// AddComment adds a comment to a ticket
func (s *SupportService) AddComment(ctx context.Context, comment *models.TicketComment) error {
	comment.CreatedAt = time.Now()
	
	coll := s.db.Collection("ticket_comments")
	result, err := coll.InsertOne(ctx, comment)
	if err != nil {
		return err
	}
	
	comment.ID = result.InsertedID.(primitive.ObjectID)
	
	// Update ticket updated_at
	ticketColl := s.db.Collection("support_tickets")
	_, err = ticketColl.UpdateOne(ctx, 
		bson.M{"_id": comment.TicketID},
		bson.M{"$set": bson.M{"updated_at": time.Now()}},
	)
	
	return err
}

// GetTicketComments retrieves all comments for a ticket
func (s *SupportService) GetTicketComments(ctx context.Context, ticketID primitive.ObjectID) ([]models.TicketComment, error) {
	coll := s.db.Collection("ticket_comments")
	
	opts := options.Find().SetSort(bson.D{{"created_at", 1}})
	cursor, err := coll.Find(ctx, bson.M{"ticket_id": ticketID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var comments []models.TicketComment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}
	
	return comments, nil
}

// GetTicketStats returns support ticket statistics
func (s *SupportService) GetTicketStats(ctx context.Context, orgID *primitive.ObjectID) (*models.TicketStats, error) {
	coll := s.db.Collection("support_tickets")
	
	filter := bson.M{}
	if orgID != nil {
		filter["org_id"] = *orgID
	}
	
	stats := &models.TicketStats{}
	
	// Total tickets
	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}
	stats.TotalTickets = int(total)
	
	// Count by status
	statuses := []models.TicketStatus{
		models.TicketStatusOpen,
		models.TicketStatusInProgress,
		models.TicketStatusResolved,
		models.TicketStatusClosed,
	}
	
	for _, status := range statuses {
		statusFilter := bson.M{}
		for k, v := range filter {
			statusFilter[k] = v
		}
		statusFilter["status"] = status
		
		count, err := coll.CountDocuments(ctx, statusFilter)
		if err != nil {
			continue
		}
		
		switch status {
		case models.TicketStatusOpen:
			stats.OpenTickets = int(count)
		case models.TicketStatusInProgress:
			stats.InProgressTickets = int(count)
		case models.TicketStatusResolved:
			stats.ResolvedTickets = int(count)
		case models.TicketStatusClosed:
			stats.ClosedTickets = int(count)
		}
	}
	
	// Critical tickets (P0)
	criticalFilter := bson.M{}
	for k, v := range filter {
		criticalFilter[k] = v
	}
	criticalFilter["priority"] = models.PriorityP0
	critical, _ := coll.CountDocuments(ctx, criticalFilter)
	stats.CriticalTickets = int(critical)
	
	// Calculate average response and resolution times
	pipeline := mongo.Pipeline{
		{{"$match", filter}},
		{{"$group", bson.D{
			{"_id", nil},
			{"avg_response", bson.D{{"$avg", "$response_time"}}},
			{"avg_resolution", bson.D{{"$avg", "$resolution_time"}}},
		}}},
	}
	
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err == nil {
		defer cursor.Close(ctx)
		if cursor.Next(ctx) {
			var result struct {
				AvgResponse    float64 `bson:"avg_response"`
				AvgResolution  float64 `bson:"avg_resolution"`
			}
			if err := cursor.Decode(&result); err == nil {
				stats.AvgResponseTime = result.AvgResponse
				stats.AvgResolutionTime = result.AvgResolution
			}
		}
	}
	
	// SLA compliance rate (placeholder)
	if stats.TotalTickets > 0 {
		nonBreachedFilter := bson.M{}
		for k, v := range filter {
			nonBreachedFilter[k] = v
		}
		nonBreachedFilter["sla_breached"] = false
		nonBreached, _ := coll.CountDocuments(ctx, nonBreachedFilter)
		stats.SLAComplianceRate = (float64(nonBreached) / float64(stats.TotalTickets)) * 100
	}
	
	return stats, nil
}
