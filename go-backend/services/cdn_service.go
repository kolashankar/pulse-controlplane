package services

import (
	"fmt"
	"os"
)

// CDNService handles CDN-related operations
type CDNService struct {
	cdnDomain string
}

// NewCDNService creates a new CDN service
func NewCDNService() *CDNService {
	cdnDomain := os.Getenv("CDN_PLAYBACK_URL")
	if cdnDomain == "" {
		cdnDomain = "https://cdn.pulse.io" // Default CDN domain
	}
	return &CDNService{
		cdnDomain: cdnDomain,
	}
}

// GenerateHLSPlaybackURL generates a CDN URL for HLS playback
func (s *CDNService) GenerateHLSPlaybackURL(projectID, filename string) string {
	return fmt.Sprintf("%s/hls/%s/%s.m3u8", s.cdnDomain, projectID, filename)
}

// GenerateFileURL generates a CDN URL for file download
func (s *CDNService) GenerateFileURL(projectID, filename string) string {
	return fmt.Sprintf("%s/recordings/%s/%s", s.cdnDomain, projectID, filename)
}

// GenerateSegmentURL generates a CDN URL for HLS segment
func (s *CDNService) GenerateSegmentURL(projectID, filename, segment string) string {
	return fmt.Sprintf("%s/hls/%s/%s/%s", s.cdnDomain, projectID, filename, segment)
}

// GetS3Config returns S3 configuration for storage
func (s *CDNService) GetS3Config() map[string]string {
	return map[string]string{
		"region": os.Getenv("R2_REGION"),
		"bucket": os.Getenv("R2_BUCKET_NAME"),
		"access_key_id": os.Getenv("R2_ACCESS_KEY_ID"),
		"secret_access_key": os.Getenv("R2_SECRET_ACCESS_KEY"),
		"endpoint": os.Getenv("R2_ENDPOINT"),
	}
}

// ValidateStorageConfig validates storage configuration
func (s *CDNService) ValidateStorageConfig(bucket, region, accessKey, secretKey string) error {
	if bucket == "" {
		return fmt.Errorf("storage bucket is required")
	}
	if region == "" {
		return fmt.Errorf("storage region is required")
	}
	if accessKey == "" {
		return fmt.Errorf("storage access key is required")
	}
	if secretKey == "" {
		return fmt.Errorf("storage secret key is required")
	}
	return nil
}
