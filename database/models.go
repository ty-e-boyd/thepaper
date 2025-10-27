package database

import (
	"time"

	"gorm.io/gorm"
)

// User represents a subscriber to the newsletter
type User struct {
	ID               uint   `gorm:"primaryKey"`
	Email            string `gorm:"uniqueIndex;not null"`
	Name             string // Optional field, nullable
	Subscribed       bool   `gorm:"default:true"`
	UnsubscribeToken string `gorm:"uniqueIndex;not null"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

// Source represents an RSS feed source
type Source struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Category  string `gorm:"not null;index"`
	URL       string `gorm:"uniqueIndex;not null"`
	Active    bool   `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// EmailSent represents an email that was sent out
type EmailSent struct {
	ID                    uint      `gorm:"primaryKey"`
	Subject               string    `gorm:"not null"`
	SentAt                time.Time `gorm:"index"`
	TotalArticlesAnalyzed int
	TotalSources          int
	RecipientCount        int
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// EmailArticle represents an article that was included in an email
type EmailArticle struct {
	ID             uint   `gorm:"primaryKey"`
	EmailID        uint   `gorm:"not null;index"`
	ArticleURL     string `gorm:"not null;index"`
	ArticleTitle   string `gorm:"not null"`
	ArticleSource  string
	RelevanceScore float64 `gorm:"type:decimal(3,1)"`
	Category       string
	Tags           string    // JSON encoded array
	Summary        string    `gorm:"type:text"`
	PublishedAt    time.Time `gorm:"index"`
	Position       int       // Position in the email (1-8)
	CreatedAt      time.Time
	Email          EmailSent `gorm:"foreignKey:EmailID;constraint:OnDelete:CASCADE"`
}

// UserEmail represents the join table tracking which users received which emails
type UserEmail struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"not null;index"`
	EmailID   uint `gorm:"not null;index"`
	SentAt    time.Time
	Opened    bool `gorm:"default:false"`
	OpenedAt  *time.Time
	CreatedAt time.Time
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Email     EmailSent `gorm:"foreignKey:EmailID;constraint:OnDelete:CASCADE"`
}

// TableName overrides for GORM
func (User) TableName() string {
	return "users"
}

func (Source) TableName() string {
	return "sources"
}

func (EmailSent) TableName() string {
	return "emails_sent"
}

func (EmailArticle) TableName() string {
	return "email_articles"
}

func (UserEmail) TableName() string {
	return "user_emails"
}
