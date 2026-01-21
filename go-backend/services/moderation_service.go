package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"pulse-control-plane/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ModerationService handles content moderation operations
type ModerationService struct {
	db             *mongo.Database
	geminiAPIKey   string
	geminiEnabled  bool
	profanityWords []string
}

// NewModerationService creates a new moderation service
func NewModerationService(db *mongo.Database, geminiAPIKey string) *ModerationService {
	// Default profanity list (can be extended)
	profanityWords := []string{
		"badword1", "badword2", "spam", "scam", // Add more as needed
	}

	return &ModerationService{
		db:             db,
		geminiAPIKey:   geminiAPIKey,
		geminiEnabled:  geminiAPIKey != "" && geminiAPIKey != "mock-gemini-key-for-testing",
		profanityWords: profanityWords,
	}
}

// GetOrCreateConfig retrieves or creates moderation config for a project
func (s *ModerationService) GetOrCreateConfig(ctx context.Context, projectID primitive.ObjectID) (*models.ModerationConfig, error) {
	coll := s.db.Collection("moderation_configs")

	var config models.ModerationConfig
	err := coll.FindOne(ctx, bson.M{"project_id": projectID}).Decode(&config)

	if err == mongo.ErrNoDocuments {
		// Create default config
		config = models.ModerationConfig{
			ProjectID:          projectID,
			Enabled:            true,
			AutoModeration:     false,
			AIEnabled:          s.geminiEnabled,
			ToxicityThreshold:  0.7,
			ProfanityThreshold: 0.6,
			SpamThreshold:      0.8,
			AutoAction:         models.ModerationActionFlag,
			NotifyAdmins:       true,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		result, err := coll.InsertOne(ctx, config)
		if err != nil {
			return nil, err
		}
		config.ID = result.InsertedID.(primitive.ObjectID)
		return &config, nil
	}

	if err != nil {
		return nil, err
	}

	return &config, nil
}

// AnalyzeText analyzes text content for moderation
func (s *ModerationService) AnalyzeText(ctx context.Context, projectID primitive.ObjectID, content, contentID, userID string, metadata map[string]interface{}) (*models.ContentAnalysis, error) {
	startTime := time.Now()

	// Get project configuration
	config, err := s.GetOrCreateConfig(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if !config.Enabled {
		return nil, errors.New("moderation is disabled for this project")
	}

	analysis := &models.ContentAnalysis{
		ProjectID:      projectID,
		ContentType:    models.ContentTypeText,
		ContentID:      contentID,
		Content:        content,
		UserID:         userID,
		Metadata:       metadata,
		MatchedRules:   []string{},
		CreatedAt:      time.Now(),
	}

	// Check whitelist
	if whitelisted, _ := s.IsWhitelisted(ctx, projectID, "user", userID); whitelisted {
		analysis.IsFlagged = false
		analysis.RecommendedAction = models.ModerationActionAllow
		analysis.Severity = models.ModerationSeverityLow
		analysis.Reason = "User is whitelisted"
		analysis.ProcessingTime = time.Since(startTime).Milliseconds()
		return s.saveAnalysis(ctx, analysis)
	}

	// Check blacklist
	if blacklisted, reason := s.IsBlacklisted(ctx, projectID, "user", userID); blacklisted {
		analysis.IsFlagged = true
		analysis.RecommendedAction = models.ModerationActionBlock
		analysis.Severity = models.ModerationSeverityCritical
		analysis.Reason = reason
		analysis.Toxicity = 1.0
		analysis.ProcessingTime = time.Since(startTime).Milliseconds()
		return s.saveAnalysis(ctx, analysis)
	}

	// Rule-based analysis
	ruleAnalysis, err := s.applyRules(ctx, projectID, content)
	if err == nil && ruleAnalysis != nil {
		analysis.MatchedRules = ruleAnalysis.MatchedRules
		analysis.Toxicity = ruleAnalysis.Toxicity
		analysis.Profanity = ruleAnalysis.Profanity
		analysis.Spam = ruleAnalysis.Spam
	}

	// AI analysis with Gemini (if enabled and configured)
	if config.AIEnabled && s.geminiEnabled {
		aiAnalysis := s.analyzeWithGemini(ctx, content)
		if aiAnalysis != nil {
			// Merge AI analysis with rule-based
			analysis.Toxicity = max(analysis.Toxicity, aiAnalysis.Toxicity)
			analysis.Profanity = max(analysis.Profanity, aiAnalysis.Profanity)
			analysis.Spam = max(analysis.Spam, aiAnalysis.Spam)
			analysis.Hate = aiAnalysis.Hate
			analysis.Sexual = aiAnalysis.Sexual
			analysis.Violence = aiAnalysis.Violence
			analysis.AnalysisMethod = "hybrid"
		} else {
			analysis.AnalysisMethod = "rule-based"
		}
	} else {
		analysis.AnalysisMethod = "rule-based"
	}

	// Determine severity and action
	analysis.Severity = s.determineSeverity(analysis.Toxicity, analysis.Profanity, analysis.Spam, analysis.Hate, analysis.Sexual, analysis.Violence)
	analysis.IsFlagged = s.shouldFlag(analysis.Toxicity, analysis.Profanity, analysis.Spam, config)
	analysis.RecommendedAction = s.determineAction(analysis.Severity, analysis.IsFlagged, config)
	analysis.Reason = s.generateReason(analysis)

	analysis.ProcessingTime = time.Since(startTime).Milliseconds()

	return s.saveAnalysis(ctx, analysis)
}

// AnalyzeImage analyzes image content (mock implementation for now)
func (s *ModerationService) AnalyzeImage(ctx context.Context, projectID primitive.ObjectID, imageURL, contentID, userID string, metadata map[string]interface{}) (*models.ContentAnalysis, error) {
	startTime := time.Now()

	config, err := s.GetOrCreateConfig(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if !config.Enabled {
		return nil, errors.New("moderation is disabled for this project")
	}

	analysis := &models.ContentAnalysis{
		ProjectID:       projectID,
		ContentType:     models.ContentTypeImage,
		ContentID:       contentID,
		Content:         imageURL,
		UserID:          userID,
		Metadata:        metadata,
		AnalysisMethod:  "mock",
		Toxicity:        0.1,
		Sexual:          0.1,
		Violence:        0.1,
		IsFlagged:       false,
		Severity:        models.ModerationSeverityLow,
		RecommendedAction: models.ModerationActionAllow,
		Reason:          "Image analysis not fully implemented (mock)",
		MatchedRules:    []string{},
		ProcessingTime:  time.Since(startTime).Milliseconds(),
		CreatedAt:       time.Now(),
	}

	return s.saveAnalysis(ctx, analysis)
}

// applyRules applies rule-based moderation
func (s *ModerationService) applyRules(ctx context.Context, projectID primitive.ObjectID, content string) (*models.ContentAnalysis, error) {
	coll := s.db.Collection("moderation_rules")

	cursor, err := coll.Find(ctx, bson.M{
		"project_id": projectID,
		"enabled":    true,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rules []models.ModerationRule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, err
	}

	analysis := &models.ContentAnalysis{
		MatchedRules: []string{},
	}

	contentLower := strings.ToLower(content)

	// Check profanity
	profanityCount := 0
	for _, word := range s.profanityWords {
		if strings.Contains(contentLower, word) {
			profanityCount++
		}
	}
	analysis.Profanity = float64(profanityCount) / 10.0 // Normalize
	if analysis.Profanity > 1.0 {
		analysis.Profanity = 1.0
	}

	// Apply custom rules
	for _, rule := range rules {
		matched := false
		checkContent := content
		if !rule.CaseSensitive {
			checkContent = contentLower
		}

		switch rule.RuleType {
		case "keyword":
			pattern := rule.Pattern
			if !rule.CaseSensitive {
				pattern = strings.ToLower(pattern)
			}
			if rule.WholeWord {
				// Word boundary check
				re := regexp.MustCompile(`\b` + regexp.QuoteMeta(pattern) + `\b`)
				matched = re.MatchString(checkContent)
			} else {
				matched = strings.Contains(checkContent, pattern)
			}

		case "regex":
			re, err := regexp.Compile(rule.Pattern)
			if err == nil {
				matched = re.MatchString(checkContent)
			}
		}

		if matched {
			analysis.MatchedRules = append(analysis.MatchedRules, rule.ID.Hex())
			
			// Increase scores based on severity
			switch rule.Severity {
			case models.ModerationSeverityLow:
				analysis.Toxicity += 0.2
			case models.ModerationSeverityMedium:
				analysis.Toxicity += 0.4
			case models.ModerationSeverityHigh:
				analysis.Toxicity += 0.6
			case models.ModerationSeverityCritical:
				analysis.Toxicity += 0.8
			}
		}
	}

	// Cap toxicity at 1.0
	if analysis.Toxicity > 1.0 {
		analysis.Toxicity = 1.0
	}

	// Simple spam detection (can be improved)
	if strings.Contains(contentLower, "click here") || 
	   strings.Contains(contentLower, "buy now") ||
	   strings.Contains(contentLower, "limited time") {
		analysis.Spam = 0.7
	}

	return analysis, nil
}

// analyzeWithGemini performs AI analysis using Gemini API (mock for now)
func (s *ModerationService) analyzeWithGemini(ctx context.Context, content string) *models.ContentAnalysis {
	if !s.geminiEnabled {
		return nil
	}

	// Mock implementation - In production, integrate with actual Gemini API
	// This would make an HTTP request to Gemini API
	analysis := &models.ContentAnalysis{
		Toxicity:  0.3,
		Profanity: 0.2,
		Spam:      0.1,
		Hate:      0.1,
		Sexual:    0.05,
		Violence:  0.05,
	}

	return analysis
}

// saveAnalysis saves content analysis to database
func (s *ModerationService) saveAnalysis(ctx context.Context, analysis *models.ContentAnalysis) (*models.ContentAnalysis, error) {
	coll := s.db.Collection("content_analysis")

	result, err := coll.InsertOne(ctx, analysis)
	if err != nil {
		return nil, err
	}

	analysis.ID = result.InsertedID.(primitive.ObjectID)
	return analysis, nil
}

// determineSeverity determines the severity level based on scores
func (s *ModerationService) determineSeverity(toxicity, profanity, spam, hate, sexual, violence float64) models.ModerationSeverity {
	maxScore := max(toxicity, profanity, spam, hate, sexual, violence)

	if maxScore >= 0.9 {
		return models.ModerationSeverityCritical
	} else if maxScore >= 0.7 {
		return models.ModerationSeverityHigh
	} else if maxScore >= 0.5 {
		return models.ModerationSeverityMedium
	}
	return models.ModerationSeverityLow
}

// shouldFlag determines if content should be flagged
func (s *ModerationService) shouldFlag(toxicity, profanity, spam float64, config *models.ModerationConfig) bool {
	return toxicity >= config.ToxicityThreshold ||
		profanity >= config.ProfanityThreshold ||
		spam >= config.SpamThreshold
}

// determineAction determines the recommended action
func (s *ModerationService) determineAction(severity models.ModerationSeverity, isFlagged bool, config *models.ModerationConfig) models.ModerationAction {
	if !isFlagged {
		return models.ModerationActionAllow
	}

	if config.AutoModeration {
		return config.AutoAction
	}

	switch severity {
	case models.ModerationSeverityCritical:
		return models.ModerationActionBlock
	case models.ModerationSeverityHigh:
		return models.ModerationActionFlag
	case models.ModerationSeverityMedium:
		return models.ModerationActionWarn
	default:
		return models.ModerationActionAllow
	}
}

// generateReason generates a human-readable reason
func (s *ModerationService) generateReason(analysis *models.ContentAnalysis) string {
	var reasons []string

	if analysis.Toxicity >= 0.7 {
		reasons = append(reasons, "high toxicity")
	}
	if analysis.Profanity >= 0.6 {
		reasons = append(reasons, "profanity detected")
	}
	if analysis.Spam >= 0.8 {
		reasons = append(reasons, "spam detected")
	}
	if analysis.Hate >= 0.7 {
		reasons = append(reasons, "hate speech")
	}
	if analysis.Sexual >= 0.7 {
		reasons = append(reasons, "sexual content")
	}
	if analysis.Violence >= 0.7 {
		reasons = append(reasons, "violent content")
	}

	if len(analysis.MatchedRules) > 0 {
		reasons = append(reasons, fmt.Sprintf("matched %d rules", len(analysis.MatchedRules)))
	}

	if len(reasons) == 0 {
		return "Content appears safe"
	}

	return "Flagged for: " + strings.Join(reasons, ", ")
}

// CreateRule creates a new moderation rule
func (s *ModerationService) CreateRule(ctx context.Context, projectID primitive.ObjectID, req *models.CreateRuleRequest) (*models.ModerationRule, error) {
	rule := &models.ModerationRule{
		ProjectID:     projectID,
		Name:          req.Name,
		Description:   req.Description,
		RuleType:      req.RuleType,
		Pattern:       req.Pattern,
		Action:        req.Action,
		Severity:      req.Severity,
		Enabled:       true,
		CaseSensitive: req.CaseSensitive,
		WholeWord:     req.WholeWord,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	coll := s.db.Collection("moderation_rules")
	result, err := coll.InsertOne(ctx, rule)
	if err != nil {
		return nil, err
	}

	rule.ID = result.InsertedID.(primitive.ObjectID)
	return rule, nil
}

// GetRules retrieves moderation rules for a project
func (s *ModerationService) GetRules(ctx context.Context, projectID primitive.ObjectID) ([]models.ModerationRule, error) {
	coll := s.db.Collection("moderation_rules")

	cursor, err := coll.Find(ctx, bson.M{"project_id": projectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rules []models.ModerationRule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, err
	}

	return rules, nil
}

// GetLogs retrieves moderation logs
func (s *ModerationService) GetLogs(ctx context.Context, projectID primitive.ObjectID, page, limit int) ([]models.ModerationLog, int64, error) {
	coll := s.db.Collection("moderation_logs")

	filter := bson.M{"project_id": projectID}

	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{"created_at", -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var logs []models.ModerationLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// AddToWhitelist adds an entry to the whitelist
func (s *ModerationService) AddToWhitelist(ctx context.Context, projectID primitive.ObjectID, req *models.WhitelistRequest) (*models.Whitelist, error) {
	whitelist := &models.Whitelist{
		ProjectID:   projectID,
		Type:        req.Type,
		Value:       req.Value,
		Description: req.Description,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
	}

	coll := s.db.Collection("whitelists")
	result, err := coll.InsertOne(ctx, whitelist)
	if err != nil {
		return nil, err
	}

	whitelist.ID = result.InsertedID.(primitive.ObjectID)
	return whitelist, nil
}

// AddToBlacklist adds an entry to the blacklist
func (s *ModerationService) AddToBlacklist(ctx context.Context, projectID primitive.ObjectID, req *models.BlacklistRequest) (*models.Blacklist, error) {
	blacklist := &models.Blacklist{
		ProjectID: projectID,
		Type:      req.Type,
		Value:     req.Value,
		Reason:    req.Reason,
		Severity:  req.Severity,
		CreatedBy: req.CreatedBy,
		ExpiresAt: req.ExpiresAt,
		CreatedAt: time.Now(),
	}

	coll := s.db.Collection("blacklists")
	result, err := coll.InsertOne(ctx, blacklist)
	if err != nil {
		return nil, err
	}

	blacklist.ID = result.InsertedID.(primitive.ObjectID)
	return blacklist, nil
}

// IsWhitelisted checks if a value is whitelisted
func (s *ModerationService) IsWhitelisted(ctx context.Context, projectID primitive.ObjectID, entryType, value string) (bool, error) {
	coll := s.db.Collection("whitelists")

	count, err := coll.CountDocuments(ctx, bson.M{
		"project_id": projectID,
		"type":       entryType,
		"value":      value,
	})

	return count > 0, err
}

// IsBlacklisted checks if a value is blacklisted
func (s *ModerationService) IsBlacklisted(ctx context.Context, projectID primitive.ObjectID, entryType, value string) (bool, string) {
	coll := s.db.Collection("blacklists")

	var blacklist models.Blacklist
	err := coll.FindOne(ctx, bson.M{
		"project_id": projectID,
		"type":       entryType,
		"value":      value,
	}).Decode(&blacklist)

	if err != nil {
		return false, ""
	}

	// Check if expired
	if blacklist.ExpiresAt != nil && time.Now().After(*blacklist.ExpiresAt) {
		return false, ""
	}

	return true, blacklist.Reason
}

// GetStats retrieves moderation statistics
func (s *ModerationService) GetStats(ctx context.Context, projectID primitive.ObjectID, period string) (*models.ModerationStats, error) {
	coll := s.db.Collection("moderation_logs")

	// Calculate time range
	var startTime time.Time
	switch period {
	case "daily":
		startTime = time.Now().AddDate(0, 0, -1)
	case "weekly":
		startTime = time.Now().AddDate(0, 0, -7)
	case "monthly":
		startTime = time.Now().AddDate(0, -1, 0)
	default:
		startTime = time.Now().AddDate(0, 0, -7)
		period = "weekly"
	}

	filter := bson.M{
		"project_id": projectID,
		"created_at": bson.M{"$gte": startTime},
	}

	total, _ := coll.CountDocuments(ctx, filter)

	stats := &models.ModerationStats{
		ProjectID:     projectID,
		TotalAnalyzed: total,
		Period:        period,
	}

	// Count by action
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$action"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return stats, nil
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID    models.ModerationAction `bson:"_id"`
		Count int64                   `bson:"count"`
	}

	if err := cursor.All(ctx, &results); err == nil {
		for _, r := range results {
			switch r.ID {
			case models.ModerationActionFlag:
				stats.Flagged = r.Count
			case models.ModerationActionBlock:
				stats.Blocked = r.Count
			case models.ModerationActionWarn:
				stats.Warnings = r.Count
			case models.ModerationActionDelete:
				stats.Deleted = r.Count
			}
		}
	}

	return stats, nil
}

// LogAction logs a moderation action
func (s *ModerationService) LogAction(ctx context.Context, log *models.ModerationLog) error {
	log.CreatedAt = time.Now()

	coll := s.db.Collection("moderation_logs")
	result, err := coll.InsertOne(ctx, log)
	if err != nil {
		return err
	}

	log.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// Helper function for max
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
