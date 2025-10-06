package models

import "time"

// Article represents a single article from an RSS feed
type Article struct {
	Title       string
	Description string
	Link        string
	Published   time.Time
	Source      string // The feed source name
	Content     string // Full content if available
}

// Config holds application configuration
type Config struct {
	GeminiAPIKey   string
	SendGridAPIKey string
	FromEmail      string
	ToEmail        string
}

// AnalyzedArticle wraps an Article with AI analysis results
type AnalyzedArticle struct {
	Article
	RelevanceScore float64
	Summary        string
	Selected       bool
}
