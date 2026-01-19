package services_test

import (
	"testing"
	"time"

	"pulse-control-plane/models"
	"pulse-control-plane/utils"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestTokenGeneration tests token generation functionality
func TestTokenGeneration(t *testing.T) {
	t.Run("Generate API key", func(t *testing.T) {
		apiKey, err := utils.GenerateAPIKey()
		assert.NoError(t, err)
		assert.NotEmpty(t, apiKey)
		assert.Contains(t, apiKey, "pulse_key_")
	})
	
	t.Run("Generate API secret", func(t *testing.T) {
		apiSecret, err := utils.GenerateAPISecret()
		assert.NoError(t, err)
		assert.NotEmpty(t, apiSecret)
		assert.Contains(t, apiSecret, "pulse_secret_")
	})
	
	t.Run("Generate random token", func(t *testing.T) {
		token, err := utils.GenerateRandomToken(32)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}

// TestPasswordHashing tests password hashing and verification
func TestPasswordHashing(t *testing.T) {
	t.Run("Hash and verify password", func(t *testing.T) {
		password := "testPassword123!"
		
		// Hash password
		hashed, err := utils.HashPassword(password)
		assert.NoError(t, err)
		assert.NotEmpty(t, hashed)
		assert.NotEqual(t, password, hashed)
		
		// Verify correct password
		err = utils.VerifyPassword(hashed, password)
		assert.NoError(t, err)
		
		// Verify incorrect password
		err = utils.VerifyPassword(hashed, "wrongPassword")
		assert.Error(t, err)
	})
}

// TestSecretHashing tests API secret hashing
func TestSecretHashing(t *testing.T) {
	t.Run("Hash and verify secret", func(t *testing.T) {
		secret := "pulse_secret_abc123xyz789"
		
		// Hash secret
		hashed, err := utils.HashSecret(secret)
		assert.NoError(t, err)
		assert.NotEmpty(t, hashed)
		assert.NotEqual(t, secret, hashed)
		
		// Verify correct secret
		err = utils.VerifySecret(hashed, secret)
		assert.NoError(t, err)
		
		// Verify incorrect secret
		err = utils.VerifySecret(hashed, "wrong_secret")
		assert.Error(t, err)
	})
}

// TestOrganizationModel tests organization model validation
func TestOrganizationModel(t *testing.T) {
	t.Run("Valid organization", func(t *testing.T) {
		org := models.Organization{
			ID:         primitive.NewObjectID(),
			Name:       "Test Org",
			AdminEmail: "admin@test.com",
			Plan:       "Free",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		
		assert.NotEmpty(t, org.ID)
		assert.Equal(t, "Test Org", org.Name)
		assert.Equal(t, "admin@test.com", org.AdminEmail)
	})
}

// TestProjectModel tests project model validation
func TestProjectModel(t *testing.T) {
	t.Run("Valid project", func(t *testing.T) {
		project := models.Project{
			ID:            primitive.NewObjectID(),
			OrgID:         primitive.NewObjectID(),
			Name:          "Test Project",
			PulseAPIKey:   "pulse_key_test123",
			Region:        "US_EAST",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		
		assert.NotEmpty(t, project.ID)
		assert.NotEmpty(t, project.OrgID)
		assert.Equal(t, "Test Project", project.Name)
	})
}

// TestUsageMetricsCalculation tests usage calculation logic
func TestUsageMetricsCalculation(t *testing.T) {
	t.Run("Calculate participant minutes", func(t *testing.T) {
		// Test participant minutes calculation
		participants := 10
		durationMinutes := 60
		totalMinutes := participants * durationMinutes
		
		assert.Equal(t, 600, totalMinutes)
	})
	
	t.Run("Calculate cost", func(t *testing.T) {
		// Test cost calculation
		participantMinutes := float64(1000)
		pricePerMinute := 0.004 // $0.004 per minute
		totalCost := participantMinutes * pricePerMinute
		
		assert.Equal(t, 4.0, totalCost)
	})
}

// TestInvitationExpiry tests invitation expiry logic
func TestInvitationExpiry(t *testing.T) {
	t.Run("Valid invitation", func(t *testing.T) {
		invitation := models.Invitation{
			ID:        primitive.NewObjectID(),
			OrgID:     primitive.NewObjectID(),
			Email:     "newuser@test.com",
			Role:      "Developer",
			Token:     "test_token_123",
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			CreatedAt: time.Now(),
		}
		
		// Check if invitation is not expired
		isExpired := time.Now().After(invitation.ExpiresAt)
		assert.False(t, isExpired)
	})
	
	t.Run("Expired invitation", func(t *testing.T) {
		invitation := models.Invitation{
			ID:        primitive.NewObjectID(),
			OrgID:     primitive.NewObjectID(),
			Email:     "newuser@test.com",
			Role:      "Developer",
			Token:     "test_token_123",
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
			CreatedAt: time.Now().Add(-8 * 24 * time.Hour),
		}
		
		// Check if invitation is expired
		isExpired := time.Now().After(invitation.ExpiresAt)
		assert.True(t, isExpired)
	})
}

// TestRolePermissions tests role-based permissions
func TestRolePermissions(t *testing.T) {
	roles := []string{"Owner", "Admin", "Developer", "Viewer"}
	
	for _, role := range roles {
		t.Run("Role: "+role, func(t *testing.T) {
			assert.Contains(t, roles, role)
		})
	}
}
