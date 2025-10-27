# Database Setup Guide

This guide covers the database integration that was added to The Paper newsletter application.

## Overview

The Paper now uses PostgreSQL with GORM (Go ORM) to:
- Store and manage RSS feed sources
- Track subscribed users
- Prevent duplicate articles from being sent
- Record email send history
- Support multiple subscribers

## Prerequisites

- PostgreSQL database (local or hosted)
- Go 1.24+ with GORM dependencies installed
- `DB_CONNECT_STRING` environment variable set

## Environment Configuration

Add the following to your `.env` file:

```bash
DB_CONNECT_STRING=postgresql://username:password@host:port/database_name

# Example for local PostgreSQL:
DB_CONNECT_STRING=postgresql://postgres:password@localhost:5432/thepaper

# Example for hosted database (Railway, Heroku, etc):
DB_CONNECT_STRING=postgresql://user:pass@example.com:5432/dbname?sslmode=require

# Note: Application is 100% database-driven
# - Recipients come from the users table
# - RSS sources come from the sources table
# - No hardcoded fallbacks - database is required
```

## Database Schema

The application uses 5 main tables:

### 1. `users`
Stores newsletter subscribers and their subscription status.

| Column | Type | Description |
|--------|------|-------------|
| id | bigserial | Primary key |
| email | text | Unique email address |
| name | text | User's name (optional) |
| subscribed | boolean | Subscription status (default: true) |
| unsubscribe_token | text | Unique token for one-click unsubscribe |
| created_at | timestamptz | Account creation timestamp |
| updated_at | timestamptz | Last update timestamp |
| deleted_at | timestamptz | Soft delete timestamp (nullable) |

**Indexes:**
- Unique index on `email`
- Unique index on `unsubscribe_token`
- Index on `deleted_at` (for soft deletes)

### 2. `sources`
Stores RSS feed sources organized by category.

| Column | Type | Description |
|--------|------|-------------|
| id | bigserial | Primary key |
| name | text | Source name/description |
| category | text | Category (e.g., "Hacker News", "General Tech News") |
| url | text | RSS feed URL |
| active | boolean | Whether source is active (default: true) |
| created_at | timestamptz | Creation timestamp |
| updated_at | timestamptz | Last update timestamp |
| deleted_at | timestamptz | Soft delete timestamp (nullable) |

**Indexes:**
- Unique index on `url`
- Index on `category`
- Index on `deleted_at`

### 3. `emails_sent`
Records each email campaign that was sent out.

| Column | Type | Description |
|--------|------|-------------|
| id | bigserial | Primary key |
| subject | text | Email subject line |
| sent_at | timestamptz | When email was sent |
| total_articles_analyzed | bigint | Number of articles analyzed |
| total_sources | bigint | Number of unique sources |
| recipient_count | bigint | Number of recipients |
| created_at | timestamptz | Record creation timestamp |
| updated_at | timestamptz | Last update timestamp |

**Indexes:**
- Index on `sent_at`

### 4. `email_articles`
Tracks which articles were included in each email (for duplicate prevention).

| Column | Type | Description |
|--------|------|-------------|
| id | bigserial | Primary key |
| email_id | bigint | Foreign key to `emails_sent` |
| article_url | text | Article URL |
| article_title | text | Article title |
| article_source | text | Source name |
| relevance_score | decimal(3,1) | AI relevance score (0.0-10.0) |
| category | text | Article category |
| tags | text | JSON-encoded tag array |
| summary | text | AI-generated summary |
| published_at | timestamptz | Article publication date |
| position | bigint | Position in email (1-8) |
| created_at | timestamptz | Record creation timestamp |

**Indexes:**
- Index on `email_id`
- Index on `article_url` (for duplicate checking)
- Index on `published_at`

**Foreign Keys:**
- `email_id` references `emails_sent(id)` with CASCADE delete

### 5. `user_emails`
Join table tracking which users received which emails.

| Column | Type | Description |
|--------|------|-------------|
| id | bigserial | Primary key |
| user_id | bigint | Foreign key to `users` |
| email_id | bigint | Foreign key to `emails_sent` |
| sent_at | timestamptz | When email was sent to this user |
| opened | boolean | Whether email was opened (default: false) |
| opened_at | timestamptz | When email was opened (nullable) |
| created_at | timestamptz | Record creation timestamp |

**Indexes:**
- Index on `user_id`
- Index on `email_id`

**Foreign Keys:**
- `user_id` references `users(id)` with CASCADE delete
- `email_id` references `emails_sent(id)` with CASCADE delete

## Initial Setup

### Step 1: Create Database

Create a PostgreSQL database for The Paper:

```sql
CREATE DATABASE thepaper;
```

Or use a hosted database service like:
- Railway
- Supabase
- Heroku Postgres
- AWS RDS
- Neon

### Step 2: Configure Environment

Set your required environment variables in `.env`:

```bash
# Required
DB_CONNECT_STRING=postgresql://user:password@localhost:5432/thepaper
GEMINI_API_KEY=your-gemini-api-key
SENDGRID_API_KEY=your-sendgrid-api-key
FROM_EMAIL=noreply@yourdomain.com

# Optional
GEMINI_RATE_LIMIT_MS=200
```

**Important:** 
- `TO_EMAIL` is no longer required. All recipients are pulled from the `users` table.
- The application will **exit with an error** if no subscribed users or active sources are found in the database.
- Run `cd scripts && go run seed_sources.go` to populate sources before first run.

### Step 3: Run Migrations

Migrations run automatically when the application starts. GORM's AutoMigrate will:
- Create tables if they don't exist
- Add missing columns
- Create indexes
- Set up foreign key constraints

To manually trigger migrations, run any script or the main application:

```bash
cd scripts
go run seed_sources.go
```

### Step 4: Seed RSS Sources

Populate the `sources` table with all RSS feeds:

```bash
cd scripts
go run seed_sources.go
```

This will:
- Connect to the database
- Run migrations
- Import all 81 RSS feeds from `feeds/sources.go`
- Skip any sources that already exist

### Step 5: Add Users

Add your first subscriber:

```bash
cd scripts
go run add_user.go
```

Default user: `tyler@tylerevan.dev` (modify script to add different users)

To add more users, edit `scripts/add_user.go`:

```go
email := "newuser@example.com"
name := "User Name"
```

Then run the script again.

## How It Works

### Application Flow

1. **Startup:**
   - Connects to PostgreSQL
   - Runs GORM AutoMigrate (creates/updates schema)
   - Loads configuration

2. **Source Management:**
   - `feeds.GetAllFeeds()` queries `sources` table for active feeds
   - **No fallbacks** - exits if database query fails or no sources found
   - Fetches articles from all active sources

3. **Duplicate Prevention:**
   - Queries `email_articles` table for articles sent in last 30 days
   - Filters out any article URLs that were recently sent
   - Only analyzes new articles with AI

4. **Email Campaign:**
   - AI selects top 8 articles
   - Creates `emails_sent` record
   - Creates `email_articles` records (one per selected article)
   - Queries `users` table for subscribed users
   - Sends email to each user
   - Creates `user_emails` record for each send

### Duplicate Prevention

Articles are deduplicated based on URL over a 30-day window:

```go
// Get articles sent in last 30 days
recentArticleURLs, err := database.GetRecentArticleURLs(30)

// Filter out duplicates
for _, article := range articles {
    if !recentArticleURLs[article.Link] {
        newArticles = append(newArticles, article)
    }
}
```

### Multi-User Support

The application now sends to all subscribed users:

```go
// Get all subscribed users
users, err := database.GetAllSubscribedUsers()

// Send to each user
for _, user := range users {
    sender.Send(cfg.FromEmail, user.Email, subject, htmlContent)
    database.CreateUserEmail(user.ID, emailRecord.ID)
}
```

## Database Functions

### User Management

```go
// Create a new user
user, err := database.CreateUser("email@example.com", "Name")

// Get all subscribed users
users, err := database.GetAllSubscribedUsers()

// Get user by email
user, err := database.GetUserByEmail("email@example.com")

// Update subscription status
err := database.UpdateUserSubscription(userID, false) // unsubscribe
```

### Source Management

```go
// Create a new source
source, err := database.CreateSource("Source Name", "Category", "https://feed.url/rss", true)

// Get all active sources
sources, err := database.GetAllActiveSources()

// Get sources by category
sources, err := database.GetSourcesByCategory("Hacker News")

// Deactivate a source
err := database.UpdateSourceActive(sourceID, false)
```

### Email Tracking

```go
// Create email record
email, err := database.CreateEmailSent(subject, totalArticles, totalSources, recipientCount)

// Save article to email
article, err := database.CreateEmailArticle(emailID, url, title, source, score, category, tags, summary, publishedAt, position)

// Record that user received email
userEmail, err := database.CreateUserEmail(userID, emailID)

// Check for recent articles (duplicate prevention)
recentURLs, err := database.GetRecentArticleURLs(30) // last 30 days
```

## Maintenance

### Adding New RSS Sources

Option 1: Via Database

```sql
INSERT INTO sources (name, category, url, active, created_at, updated_at)
VALUES ('Source Name', 'Category', 'https://feed.url/rss', true, NOW(), NOW());
```

Option 2: Via Code

1. Add to `feeds/sources.go`
2. Run `cd scripts && go run seed_sources.go`

### Managing Users

```sql
-- List all users
SELECT id, email, name, subscribed, created_at FROM users;

-- Unsubscribe a user
UPDATE users SET subscribed = false WHERE email = 'user@example.com';

-- Resubscribe a user
UPDATE users SET subscribed = true WHERE email = 'user@example.com';

-- Delete a user (soft delete)
UPDATE users SET deleted_at = NOW() WHERE email = 'user@example.com';
```

### Viewing Email History

```sql
-- Recent emails sent
SELECT id, subject, sent_at, recipient_count, total_articles_analyzed
FROM emails_sent
ORDER BY sent_at DESC
LIMIT 10;

-- Articles in a specific email
SELECT article_title, article_source, relevance_score, position
FROM email_articles
WHERE email_id = 1
ORDER BY position;

-- Check if article was sent recently
SELECT ea.article_url, ea.article_title, es.sent_at
FROM email_articles ea
JOIN emails_sent es ON ea.email_id = es.id
WHERE ea.article_url = 'https://example.com/article'
ORDER BY es.sent_at DESC;
```

## Troubleshooting

### Connection Issues

**Error: "Failed to connect to database"**

Check:
1. PostgreSQL is running
2. `DB_CONNECT_STRING` is correct
3. Database exists
4. User has proper permissions
5. Firewall allows connection (if remote)

**Error: "No active sources found in database"**

Solution:
```bash
cd scripts
go run seed_sources.go
```

**Error: "No subscribed users found, exiting"**

Solution:
```bash
cd scripts
# Edit add_user.go with your email/name
go run add_user.go
```

**Error: "SSL is required"**

Add `?sslmode=require` to connection string:
```bash
DB_CONNECT_STRING=postgresql://user:pass@host:5432/db?sslmode=require
```

### Migration Issues

**Error: "permission denied for schema public"**

Grant permissions:
```sql
GRANT ALL ON SCHEMA public TO your_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO your_user;
```

**Tables not being created:**

Check GORM logs in application output. Ensure AutoMigrate is being called.

### Performance

**Slow queries:**

All common queries have indexes. Check:
```sql
-- Verify indexes exist
\di

-- Analyze query performance
EXPLAIN ANALYZE SELECT * FROM email_articles WHERE article_url = 'url';
```

## Security Best Practices

1. **Never commit `.env` file** - Contains sensitive database credentials
2. **Use strong passwords** - Especially for hosted databases
3. **Enable SSL** - Use `?sslmode=require` for remote connections
4. **Limit permissions** - Database user should only have necessary permissions
5. **Regular backups** - Use pg_dump or your hosting provider's backup feature
6. **Secure unsubscribe tokens** - Already using 32-byte random tokens

## Backup & Restore

### Backup

```bash
# Backup entire database
pg_dump -h localhost -U postgres -d thepaper > backup.sql

# Backup specific tables
pg_dump -h localhost -U postgres -d thepaper -t users -t sources > partial_backup.sql
```

### Restore

```bash
# Restore from backup
psql -h localhost -U postgres -d thepaper < backup.sql
```

## Important Notes

**100% Database-Driven:**
- The application **requires** an active database connection
- No hardcoded fallbacks for sources or users
- Application will exit if database is unavailable or empty
- This ensures data integrity and prevents incomplete sends

**First-Time Setup Checklist:**
1. ✅ Database created and accessible
2. ✅ `DB_CONNECT_STRING` set in `.env`
3. ✅ Sources seeded (`cd scripts && go run seed_sources.go`)
4. ✅ At least one user added (`cd scripts && go run add_user.go`)
5. ✅ Ready to send!

## Next Steps

- [ ] Build web interface for subscription management (`/subscribe`, `/unsubscribe`)
- [ ] Add email open tracking (webhook from SendGrid)
- [ ] Implement user preferences (categories, frequency)
- [ ] Create admin dashboard
- [ ] Add analytics and reporting

## Resources

- [GORM Documentation](https://gorm.io/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [SendGrid Documentation](https://docs.sendgrid.com/)