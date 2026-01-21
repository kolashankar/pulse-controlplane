package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	// Server
	Port        string
	GinMode     string
	Environment string

	// Database
	MongoURI    string
	MongoDBName string

	// LiveKit
	LiveKitHost      string
	LiveKitAPIKey    string
	LiveKitAPISecret string

	// CDN & Storage
	CDNIngestURL      string
	CDNPlaybackURL    string
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string

	// Redis
	RedisURL string

	// CORS
	CORSOrigins []string

	// Security
	JWTSecret     string
	APIKeyPepper  string

	// Rate Limiting
	RateLimitRequestsPerMinute int

	// Logging
	LogLevel string

	// AI Moderation
	GeminiAPIKey       string
	ModerationEnabled  bool

	// Razorpay
	RazorpayKeyID       string
	RazorpayKeySecret   string
	RazorpayWebhookSecret string
}

var AppConfig *Config

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg("No .env file found, using environment variables")
	}

	rateLimitStr := getEnv("RATE_LIMIT_REQUESTS_PER_MINUTE", "100")
	rateLimit, err := strconv.Atoi(rateLimitStr)
	if err != nil {
		rateLimit = 100
	}

	corsOrigins := strings.Split(getEnv("CORS_ORIGINS", "*"), ",")

	config := &Config{
		// Server
		Port:        getEnv("PORT", "8080"),
		GinMode:     getEnv("GIN_MODE", "debug"),
		Environment: getEnv("ENVIRONMENT", "development"),

		// Database
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName: getEnv("MONGO_DB_NAME", "pulse_development"),

		// LiveKit
		LiveKitHost:      getEnv("LIVEKIT_HOST", "wss://livekit.pulse.io"),
		LiveKitAPIKey:    getEnv("LIVEKIT_API_KEY", ""),
		LiveKitAPISecret: getEnv("LIVEKIT_API_SECRET", ""),

		// CDN & Storage
		CDNIngestURL:      getEnv("CDN_INGEST_URL", ""),
		CDNPlaybackURL:    getEnv("CDN_PLAYBACK_URL", ""),
		R2AccountID:       getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:      getEnv("R2_BUCKET_NAME", "pulse-recordings"),

		// Redis
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),

		// CORS
		CORSOrigins: corsOrigins,

		// Security
		JWTSecret:    getEnv("JWT_SECRET", "change-this-secret"),
		APIKeyPepper: getEnv("API_KEY_PEPPER", "change-this-pepper"),

		// Rate Limiting
		RateLimitRequestsPerMinute: rateLimit,

		// Logging
		LogLevel: getEnv("LOG_LEVEL", "info"),

		// AI Moderation
		GeminiAPIKey:      getEnv("GEMINI_API_KEY", "mock-gemini-key-for-testing"),
		ModerationEnabled: getEnv("MODERATION_ENABLED", "true") == "true",

		// Razorpay
		RazorpayKeyID:         getEnv("RAZORPAY_KEY_ID", "rzp_test_mock"),
		RazorpayKeySecret:     getEnv("RAZORPAY_KEY_SECRET", "mock_secret"),
		RazorpayWebhookSecret: getEnv("RAZORPAY_WEBHOOK_SECRET", "mock_webhook_secret"),
	}

	AppConfig = config
	return config, nil
}

// getEnv gets an environment variable with a default fallback
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// IsProduction checks if the environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment checks if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}
