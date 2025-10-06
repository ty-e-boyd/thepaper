package config

import (
	"fmt"
	"os"

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

	toEmail := os.Getenv("TO_EMAIL")
	if toEmail == "" {
		return nil, fmt.Errorf("TO_EMAIL environment variable is required")
	}

	return &models.Config{
		GeminiAPIKey:   geminiKey,
		SendGridAPIKey: sendgridKey,
		FromEmail:      fromEmail,
		ToEmail:        toEmail,
	}, nil
}
