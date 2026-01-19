package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"pulse-control-plane/database"
	"pulse-control-plane/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// StatusService handles system status and monitoring
type StatusService struct {
	db          *mongo.Database
	projectsColl *mongo.Collection
}

// NewStatusService creates a new status service
func NewStatusService() *StatusService {
	db := database.GetDB()
	return &StatusService{
		db:          db,
		projectsColl: db.Collection(models.Project{}.TableName()),
	}
}

// SystemStatus represents the overall system status
type SystemStatus struct {
	Status          string                 `json:"status"` // Operational, Degraded, Down
	Version         string                 `json:"version"`
	Uptime          string                 `json:"uptime"`
	Database        ServiceStatus          `json:"database"`
	API             ServiceStatus          `json:"api"`
	LiveKit         ServiceStatus          `json:"livekit"`
	Regions         []RegionStatus         `json:"regions"`
	LastChecked     time.Time              `json:"last_checked"`
	ActiveProjects  int64                  `json:"active_projects"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ServiceStatus represents the status of a single service
type ServiceStatus struct {
	Status       string    `json:"status"` // Up, Down, Degraded
	ResponseTime int64     `json:"response_time_ms"`
	LastChecked  time.Time `json:"last_checked"`
	Message      string    `json:"message,omitempty"`
}

// RegionStatus represents the status of a region
type RegionStatus struct {
	Region       string    `json:"region"`
	Status       string    `json:"status"`
	Latency      int64     `json:"latency_ms"`
	LastChecked  time.Time `json:"last_checked"`
	ActiveRooms  int       `json:"active_rooms"`
	Message      string    `json:"message,omitempty"`
}

// ProjectHealth represents the health status of a project
type ProjectHealth struct {
	ProjectID        string                 `json:"project_id"`
	ProjectName      string                 `json:"project_name"`
	Status           string                 `json:"status"`
	Region           string                 `json:"region"`
	ActiveRooms      int                    `json:"active_rooms"`
	ActiveParticipants int                  `json:"active_participants"`
	APIKeyValid      bool                   `json:"api_key_valid"`
	WebhookConfigured bool                  `json:"webhook_configured"`
	LastActivity     time.Time              `json:"last_activity,omitempty"`
	Issues           []string               `json:"issues,omitempty"`
	Metrics          map[string]interface{} `json:"metrics,omitempty"`
}

var startTime = time.Now()

// GetSystemStatus returns the overall system status
func (s *StatusService) GetSystemStatus(ctx context.Context) (*SystemStatus, error) {
	now := time.Now()

	// Check database status
	dbStatus := s.checkDatabaseStatus(ctx)

	// Check API status (always up if we can execute this code)
	apiStatus := ServiceStatus{
		Status:       "Up",
		ResponseTime: 0,
		LastChecked:  now,
		Message:      "API is operational",
	}

	// Check LiveKit status (placeholder - would need actual LiveKit connection)
	liveKitStatus := s.checkLiveKitStatus(ctx)

	// Check region status
	regions := s.checkRegionStatus(ctx)

	// Count active projects
	activeProjects, err := s.projectsColl.CountDocuments(ctx, bson.M{"is_deleted": false})
	if err != nil {
		activeProjects = 0
	}

	// Determine overall status
	overallStatus := "Operational"
	if dbStatus.Status == "Down" || liveKitStatus.Status == "Down" {
		overallStatus = "Down"
	} else if dbStatus.Status == "Degraded" || liveKitStatus.Status == "Degraded" {
		overallStatus = "Degraded"
	}

	uptime := time.Since(startTime).Round(time.Second).String()

	return &SystemStatus{
		Status:         overallStatus,
		Version:        "1.0.0",
		Uptime:         uptime,
		Database:       dbStatus,
		API:            apiStatus,
		LiveKit:        liveKitStatus,
		Regions:        regions,
		LastChecked:    now,
		ActiveProjects: activeProjects,
		Metadata: map[string]interface{}{
			"environment": "production",
			"go_version":  "1.21+",
		},
	}, nil
}

// GetProjectHealth returns the health status of a specific project
func (s *StatusService) GetProjectHealth(ctx context.Context, projectID primitive.ObjectID) (*ProjectHealth, error) {
	// Get project details
	var project models.Project
	err := s.projectsColl.FindOne(ctx, bson.M{"_id": projectID, "is_deleted": false}).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	issues := []string{}

	// Check if API key is valid
	apiKeyValid := project.PulseAPIKey != "" && project.PulseAPISecret != ""
	if !apiKeyValid {
		issues = append(issues, "API key is missing or invalid")
	}

	// Check webhook configuration
	webhookConfigured := project.WebhookURL != ""
	if !webhookConfigured {
		issues = append(issues, "Webhook URL not configured")
	}

	// Check region configuration
	if project.Region == "" {
		issues = append(issues, "Region not configured")
	}

	// Determine status
	status := "Healthy"
	if len(issues) > 0 {
		status = "Warning"
	}
	if !apiKeyValid {
		status = "Critical"
	}

	return &ProjectHealth{
		ProjectID:         project.ID.Hex(),
		ProjectName:       project.Name,
		Status:            status,
		Region:            project.Region,
		ActiveRooms:       0, // Would need LiveKit integration
		ActiveParticipants: 0, // Would need LiveKit integration
		APIKeyValid:       apiKeyValid,
		WebhookConfigured: webhookConfigured,
		LastActivity:      project.UpdatedAt,
		Issues:            issues,
		Metrics: map[string]interface{}{
			"created_at": project.CreatedAt,
			"updated_at": project.UpdatedAt,
		},
	}, nil
}

// checkDatabaseStatus checks MongoDB connectivity
func (s *StatusService) checkDatabaseStatus(ctx context.Context) ServiceStatus {
	start := time.Now()
	
	// Simple ping to check database
	err := s.db.Client().Ping(ctx, nil)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		return ServiceStatus{
			Status:       "Down",
			ResponseTime: elapsed,
			LastChecked:  time.Now(),
			Message:      fmt.Sprintf("Database connection failed: %v", err),
		}
	}

	status := "Up"
	message := "Database is operational"
	
	// If response time is high, mark as degraded
	if elapsed > 100 {
		status = "Degraded"
		message = "Database response time is high"
	}

	return ServiceStatus{
		Status:       status,
		ResponseTime: elapsed,
		LastChecked:  time.Now(),
		Message:      message,
	}
}

// checkLiveKitStatus checks LiveKit server status
func (s *StatusService) checkLiveKitStatus(ctx context.Context) ServiceStatus {
	// Placeholder implementation
	// In production, this would make actual HTTP/gRPC calls to LiveKit servers
	
	return ServiceStatus{
		Status:       "Up",
		ResponseTime: 25,
		LastChecked:  time.Now(),
		Message:      "LiveKit servers operational",
	}
}

// checkRegionStatus checks the status of all regions
func (s *StatusService) checkRegionStatus(ctx context.Context) []RegionStatus {
	// Placeholder implementation
	// In production, this would check actual LiveKit servers in each region
	
	regions := []string{"us-east", "us-west", "eu-west", "asia-south"}
	statuses := make([]RegionStatus, 0, len(regions))

	for _, region := range regions {
		statuses = append(statuses, RegionStatus{
			Region:      region,
			Status:      "Up",
			Latency:     30 + int64(len(region)*5), // Mock latency
			LastChecked: time.Now(),
			ActiveRooms: 0,
			Message:     fmt.Sprintf("Region %s is operational", region),
		})
	}

	return statuses
}

// GetRegionAvailability returns the availability of all regions
func (s *StatusService) GetRegionAvailability(ctx context.Context) ([]RegionStatus, error) {
	return s.checkRegionStatus(ctx), nil
}

// PingService performs a health check on an external service
func (s *StatusService) PingService(ctx context.Context, url string, timeout time.Duration) ServiceStatus {
	start := time.Now()
	
	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return ServiceStatus{
			Status:      "Down",
			LastChecked: time.Now(),
			Message:     fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	resp, err := client.Do(req)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		return ServiceStatus{
			Status:       "Down",
			ResponseTime: elapsed,
			LastChecked:  time.Now(),
			Message:      fmt.Sprintf("Request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	status := "Up"
	message := "Service is operational"

	if resp.StatusCode >= 500 {
		status = "Down"
		message = fmt.Sprintf("Service returned error: %d", resp.StatusCode)
	} else if resp.StatusCode >= 400 {
		status = "Degraded"
		message = fmt.Sprintf("Service returned client error: %d", resp.StatusCode)
	} else if elapsed > 1000 {
		status = "Degraded"
		message = "Service response time is high"
	}

	return ServiceStatus{
		Status:       status,
		ResponseTime: elapsed,
		LastChecked:  time.Now(),
		Message:      message,
	}
}
