package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ty-e-boyd/thepaper/models"
)

// Load reads configuration from environment variables
func Load() (*models.Config, error) {
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

	sendgridKey := os.Getenv("SENDGRID_API_KEY")
	if sendgridKey == "" {
		return nil, fmt.Errorf("SENDGRID_API_KEY environment variable is required")
	}

	fromEmail := os.Getenv("FROM_EMAIL")
	if fromEmail == "" {
		return nil, fmt.Errorf("FROM_EMAIL environment variable is required")
	}

	// Optional: rate limit delay in milliseconds (default 200ms for paid tier)
	rateLimitMs := 200
	if rateLimitStr := os.Getenv("GEMINI_RATE_LIMIT_MS"); rateLimitStr != "" {
		parsed, err := strconv.Atoi(rateLimitStr)
		if err != nil {
			return nil, fmt.Errorf("GEMINI_RATE_LIMIT_MS must be a number: %w", err)
		}
		rateLimitMs = parsed
	}

	return &models.Config{
		GeminiAPIKey:    geminiKey,
		SendGridAPIKey:  sendgridKey,
		FromEmail:       fromEmail,
		GeminiRateLimit: time.Duration(rateLimitMs) * time.Millisecond,
	}, nil
}
