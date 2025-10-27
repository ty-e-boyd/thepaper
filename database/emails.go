package database

import (
	"encoding/json"
	"fmt"
	"time"
)

// CreateEmailSent creates a new email sent record
func CreateEmailSent(subject string, totalArticlesAnalyzed, totalSources, recipientCount int) (*EmailSent, error) {
	email := &EmailSent{
		Subject:               subject,
		SentAt:                time.Now(),
		TotalArticlesAnalyzed: totalArticlesAnalyzed,
		TotalSources:          totalSources,
		RecipientCount:        recipientCount,
	}

	result := DB.Create(email)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create email record: %w", result.Error)
	}

	return email, nil
}

// CreateEmailArticle creates a record of an article included in an email
func CreateEmailArticle(emailID uint, url, title, source string, relevanceScore float64, category string, tags []string, summary string, publishedAt time.Time, position int) (*EmailArticle, error) {
	// Encode tags as JSON
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("failed to encode tags: %w", err)
	}

	article := &EmailArticle{
		EmailID:        emailID,
		ArticleURL:     url,
		ArticleTitle:   title,
		ArticleSource:  source,
		RelevanceScore: relevanceScore,
		Category:       category,
		Tags:           string(tagsJSON),
		Summary:        summary,
		PublishedAt:    publishedAt,
		Position:       position,
	}

	result := DB.Create(article)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create email article: %w", result.Error)
	}

	return article, nil
}

// CreateUserEmail records that an email was sent to a user
func CreateUserEmail(userID, emailID uint) (*UserEmail, error) {
	userEmail := &UserEmail{
		UserID:  userID,
		EmailID: emailID,
		SentAt:  time.Now(),
		Opened:  false,
	}

	result := DB.Create(userEmail)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user email record: %w", result.Error)
	}

	return userEmail, nil
}

// GetRecentEmailArticles returns articles sent in the last N days to prevent duplicates
func GetRecentEmailArticles(days int) ([]EmailArticle, error) {
	var articles []EmailArticle
	cutoff := time.Now().AddDate(0, 0, -days)

	result := DB.Where("created_at > ?", cutoff).Find(&articles)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get recent email articles: %w", result.Error)
	}

	return articles, nil
}

// GetRecentArticleURLs returns a map of article URLs sent in the last N days
func GetRecentArticleURLs(days int) (map[string]bool, error) {
	articles, err := GetRecentEmailArticles(days)
	if err != nil {
		return nil, err
	}

	urlMap := make(map[string]bool)
	for _, article := range articles {
		urlMap[article.ArticleURL] = true
	}

	return urlMap, nil
}

// GetEmailByID retrieves an email record by its ID
func GetEmailByID(emailID uint) (*EmailSent, error) {
	var email EmailSent
	result := DB.First(&email, emailID)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get email: %w", result.Error)
	}
	return &email, nil
}

// GetEmailArticles retrieves all articles for a specific email
func GetEmailArticles(emailID uint) ([]EmailArticle, error) {
	var articles []EmailArticle
	result := DB.Where("email_id = ?", emailID).Order("position").Find(&articles)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get email articles: %w", result.Error)
	}
	return articles, nil
}

// GetUserEmails retrieves all email records for a specific user
func GetUserEmails(userID uint) ([]UserEmail, error) {
	var userEmails []UserEmail
	result := DB.Where("user_id = ?", userID).Order("sent_at DESC").Find(&userEmails)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user emails: %w", result.Error)
	}
	return userEmails, nil
}

// MarkEmailOpened marks an email as opened by a user
func MarkEmailOpened(userEmailID uint) error {
	now := time.Now()
	result := DB.Model(&UserEmail{}).Where("id = ?", userEmailID).Updates(map[string]interface{}{
		"opened":    true,
		"opened_at": now,
	})
	if result.Error != nil {
		return fmt.Errorf("failed to mark email as opened: %w", result.Error)
	}
	return nil
}
