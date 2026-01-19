package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pulse-control-plane/config"
	"pulse-control-plane/database"
	"pulse-control-plane/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TokenClaims represents custom JWT claims for LiveKit tokens
type TokenClaims struct {
	jwt.RegisteredClaims
	Video *VideoGrant `json:"video,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// VideoGrant represents permissions for video/room access
type VideoGrant struct {
	RoomJoin     bool   `json:"roomJoin,omitempty"`
	RoomCreate   bool   `json:"roomCreate,omitempty"`
	RoomList     bool   `json:"roomList,omitempty"`
	RoomAdmin    bool   `json:"roomAdmin,omitempty"`
	RoomName     string `json:"roomName,omitempty"`
	CanPublish   bool   `json:"canPublish,omitempty"`
	CanSubscribe bool   `json:"canSubscribe,omitempty"`
}

// TokenRequest represents the input for token creation
type TokenRequest struct {
	RoomName     string            `json:"room_name" binding:"required"`
	Participant  string            `json:"participant_name" binding:"required"`
	CanPublish   bool              `json:"can_publish"`
	CanSubscribe bool              `json:"can_subscribe"`
	Metadata     map[string]string `json:"metadata"`
	ClientIP     string            `json:"client_ip"`     // Optional: for region selection
	PreferredRegion string         `json:"preferred_region"` // Optional: user preference
}

// TokenResponse represents the token creation response
type TokenResponse struct {
	Token      string    `json:"token"`
	ServerURL  string    `json:"server_url"`
	ExpiresAt  time.Time `json:"expires_at"`
	ProjectID  string    `json:"project_id"`
	RoomName   string    `json:"room_name"`
	Participant string   `json:"participant_name"`
	Region     string    `json:"region"`      // Selected region
	FallbackURLs []string `json:"fallback_urls"` // Fallback server URLs
}

type TokenService struct {
	config          *config.Config
	projectService  *ProjectService
	regionService   *RegionService
}

func NewTokenService(cfg *config.Config) *TokenService {
	return &TokenService{
		config:         cfg,
		projectService: NewProjectService(),
		regionService:  NewRegionService(),
	}
}

// CreateToken generates a LiveKit JWT token for a project
func (s *TokenService) CreateToken(ctx context.Context, projectID primitive.ObjectID, req *TokenRequest) (*TokenResponse, error) {
	// Get project details
	project, err := s.projectService.GetProject(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	// Check if video is enabled for this project
	if !project.VideoEnabled {
		return nil, errors.New("video/audio features are not enabled for this project")
	}

	// Create token with permissions
	token, expiresAt, err := s.generateLiveKitToken(project, req)
	if err != nil {
		log.Error().Err(err).Str("project_id", projectID.Hex()).Msg("Failed to generate token")
		return nil, errors.New("failed to generate token")
	}

	log.Info().
		Str("project_id", projectID.Hex()).
		Str("room", req.RoomName).
		Str("participant", req.Participant).
		Msg("Token created")

	return &TokenResponse{
		Token:       token,
		ServerURL:   project.LiveKitURL,
		ExpiresAt:   expiresAt,
		ProjectID:   projectID.Hex(),
		RoomName:    req.RoomName,
		Participant: req.Participant,
	}, nil
}

// ValidateToken validates an existing token
func (s *TokenService) ValidateToken(ctx context.Context, tokenString string) (bool, map[string]interface{}, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return false, nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		// Extract relevant information
		info := map[string]interface{}{
			"subject":    claims.Subject,
			"issuer":     claims.Issuer,
			"expires_at": claims.ExpiresAt.Time,
			"metadata":   claims.Metadata,
		}

		if claims.Video != nil {
			info["video_grant"] = claims.Video
		}

		return true, info, nil
	}

	return false, nil, errors.New("invalid token")
}

// generateLiveKitToken generates a JWT token for LiveKit
func (s *TokenService) generateLiveKitToken(project *models.Project, req *TokenRequest) (string, time.Time, error) {
	// Token expiry (4 hours by default)
	expiresAt := time.Now().Add(4 * time.Hour)

	// Create video grant
	videoGrant := &VideoGrant{
		RoomJoin:     true,
		RoomName:     req.RoomName,
		CanPublish:   req.CanPublish,
		CanSubscribe: req.CanSubscribe,
	}

	// Add project metadata
	metadata := req.Metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["project_id"] = project.ID.Hex()
	metadata["org_id"] = project.OrgID.Hex()

	// Create claims
	claims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   req.Participant,
			Issuer:    "pulse-control-plane",
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
		Video:    videoGrant,
		Metadata: metadata,
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with JWT secret (in production, use LiveKit API secret)
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// GetProjectByAPIKey retrieves a project by API key (used by middleware)
func (s *TokenService) GetProjectByAPIKey(ctx context.Context, apiKey string) (*models.Project, error) {
	projectCollection := database.GetCollection("projects")
	var project models.Project

	err := projectCollection.FindOne(ctx, bson.M{
		"pulse_api_key": apiKey,
		"is_deleted":    false,
	}).Decode(&project)

	if err != nil {
		return nil, errors.New("invalid API key")
	}

	return &project, nil
}
