# AGENTS.md - AI Agent Guidelines for The Paper

## Project Overview

**The Paper** is a daily programming news aggregator that:
- Fetches articles from 81+ RSS feeds (tech/programming sources)
- Uses Google Gemini AI to score and select top articles
- Generates AI-powered summaries
- Sends curated HTML email digest via SendGrid
- Stores everything in PostgreSQL database
- Supports multiple subscribers with duplicate prevention

**Tech Stack:** Go 1.24+, PostgreSQL, GORM, SendGrid API, Google Gemini API

---

## Core Principles

### 1. Database-First Architecture
- **All data comes from the database** - no hardcoded fallbacks
- Users, RSS sources, email history - everything in PostgreSQL
- Application will **exit with error** if database unavailable or empty
- This is intentional (fail-fast approach)

### 2. 100% Database-Driven
- Recipients: From `users` table (NO `TO_EMAIL` env var)
- Sources: From `sources` table (NO hardcoded feeds at runtime)
- Duplicate prevention: Via `email_articles` table (30-day window)
- Email tracking: All sends recorded in database

### 3. Fail-Fast Error Handling
- If database connection fails → EXIT
- If no subscribed users → EXIT
- If no active sources → EXIT
- Clear, actionable error messages with solutions

---

## Project Structure

```
thepaper/
├── main.go                      # Entry point (send flow orchestration)
├── config/config.go             # Environment variable loading
├── models/types.go              # Data structures (Article, Config, etc.)
├── database/                    # GORM database layer
│   ├── db.go                    # Connection, migrations
│   ├── models.go                # GORM models (User, Source, EmailSent, etc.)
│   ├── users.go                 # User CRUD operations
│   ├── sources.go               # RSS source management
│   └── emails.go                # Email tracking & duplicate prevention
├── feeds/
│   ├── sources.go               # GetAllFeeds() pulls from DB
│   └── fetcher.go               # RSS feed fetching logic
├── ai/
│   └── analyzer.go              # Gemini AI integration
├── email/
│   ├── builder.go               # HTML email generation
│   └── sender.go                # SendGrid integration
├── scripts/                     # Utility scripts
│   ├── seed_sources.go          # Populate sources table
│   └── add_user.go              # Add subscribers
└── docs/                        # Documentation
    ├── DATABASE_SETUP.md        # Complete DB guide
    ├── DRY_RUN_FEATURE.md       # Dry-run documentation
    └── *.md                     # Various guides
```

---

## Database Schema

### Core Tables (5)

1. **users** - Newsletter subscribers
   - Fields: email (unique), name, subscribed, unsubscribe_token
   - Indexes: email, unsubscribe_token

2. **sources** - RSS feed sources
   - Fields: name, category, url (unique), active
   - Indexes: url, category

3. **emails_sent** - Email campaign records
   - Fields: subject, sent_at, total_articles_analyzed, recipient_count

4. **email_articles** - Articles in each email (for duplicate tracking)
   - Fields: email_id (FK), article_url, article_title, relevance_score, position
   - Indexes: email_id, article_url (for duplicate checking)

5. **user_emails** - Join table (who received what)
   - Fields: user_id (FK), email_id (FK), sent_at, opened
   - Purpose: Individual send tracking, future analytics

**Foreign Keys:** All use CASCADE delete for data integrity

---

## Key Functions by Package

### database/users.go
- `CreateUser(email, name)` - Add subscriber with unsubscribe token
- `GetAllSubscribedUsers()` - Get recipients (used by main.go)
- `GetUserByEmail(email)` - Check if user exists
- `UpdateUserSubscription(userID, subscribed)` - For unsubscribe

### database/sources.go
- `GetAllActiveSources()` - Get RSS feeds (used by feeds/sources.go)
- `CreateSource(name, category, url, active)` - Add feed
- `GetSourceByURL(url)` - Check for duplicates

### database/emails.go
- `CreateEmailSent()` - Record email campaign
- `CreateEmailArticle()` - Save article to email
- `CreateUserEmail()` - Record send per user
- `GetRecentArticleURLs(days)` - Duplicate prevention (30 days)

### feeds/sources.go
- `GetAllFeeds()` - Queries DB for active sources, returns URLs
- **NO FALLBACKS** - Exits if DB query fails

### main.go Flow
1. Connect to DB, run migrations
2. Get subscribed users (exit if none)
3. Get RSS sources from DB (exit if none)
4. Fetch articles from all sources
5. Filter duplicates (last 30 days via DB)
6. AI analysis with Gemini (score articles)
7. Select top 8 articles
8. Generate AI summaries
9. Create email record in DB
10. Save articles to DB
11. Send to each user (or skip if `--dry-run`)
12. Record each send in DB

---

## Environment Variables

### Required
```bash
DB_CONNECT_STRING=postgresql://user:pass@host:5432/thepaper
GEMINI_API_KEY=your-gemini-api-key
SENDGRID_API_KEY=your-sendgrid-api-key
FROM_EMAIL=sender@example.com
```

### Optional
```bash
GEMINI_RATE_LIMIT_MS=200  # Rate limiting (default: 200ms)
```

### Removed (No Longer Used)
- ~~`TO_EMAIL`~~ - Recipients now from database

---

## Command-Line Flags

### `--dry-run`
Preview newsletter without sending emails.

**Usage:**
```bash
./thepaper --dry-run
```

**Behavior:**
- ✅ Fetches articles
- ✅ Analyzes with AI
- ✅ Creates email record in DB
- ✅ Saves articles to DB
- ❌ Does NOT send emails
- ❌ Does NOT create user_emails records
- ✅ Shows comprehensive summary

**Use Cases:** Testing, debugging, content preview, cron testing

---

## Coding Guidelines

### 1. Database Operations
```go
// Always check errors
users, err := database.GetAllSubscribedUsers()
if err != nil {
    log.Fatalf("Failed to get users: %v", err)
}

// Check for empty results
if len(users) == 0 {
    log.Fatalf("No subscribed users found")
}
```

### 2. GORM Best Practices
- Use proper struct tags: `gorm:"uniqueIndex;not null"`
- Always specify foreign keys with constraints
- Use soft deletes: `gorm.DeletedAt`
- Index all WHERE/JOIN columns
- Use transactions for multi-step operations (if needed)

### 3. Error Messages
**Bad:** "Error getting data"
**Good:** "Failed to get subscribed users: connection refused. Check DB_CONNECT_STRING in .env"

Include:
- What failed
- Why it failed (error details)
- How to fix it (if known)

### 4. Logging
```go
// Startup
log.Println("Connecting to database...")
log.Println("✓ Database connection established")

// Progress
log.Printf("Fetched %d articles", len(articles))

// Errors
log.Fatalf("Failed to connect: %v", err)
```

Use emojis for clarity: ✓ ✗ ⊘ 🔍 📊 👥 📧

### 5. No Hardcoded Data
**Never hardcode:**
- Email recipients (use database)
- RSS feed URLs at runtime (use database)
- Configuration values (use env vars)

**Exception:** `feeds/sources.go` has hardcoded map ONLY for seeding scripts

---

## Testing Guidelines

### Before Making Changes
1. Read relevant documentation in `docs/`
2. Understand current database schema
3. Check existing error handling patterns
4. Test with `--dry-run` first

### After Making Changes
1. Run `go build -o thepaper` (must compile)
2. Test with `--dry-run` flag
3. Check database records were created
4. Verify error messages are clear
5. Update documentation if behavior changed

### Database Testing
```bash
# Seed sources (first time)
cd scripts && go run seed_sources.go

# Add test user
cd scripts && go run add_user.go

# Dry run
cd .. && ./thepaper --dry-run

# Real send (only after dry-run success)
./thepaper
```

---

## Common Tasks

### Adding a New RSS Source
**Option 1: Via Script**
1. Edit `feeds/sources.go` (add to FeedSources map)
2. Run `cd scripts && go run seed_sources.go`

**Option 2: Via SQL**
```sql
INSERT INTO sources (name, category, url, active, created_at, updated_at)
VALUES ('Feed Name', 'Category', 'https://feed.url/rss', true, NOW(), NOW());
```

### Adding a New User
**Option 1: Via Script**
1. Edit `scripts/add_user.go` (change email/name)
2. Run `cd scripts && go run add_user.go`

**Option 2: Via SQL**
```sql
INSERT INTO users (email, name, subscribed, unsubscribe_token, created_at, updated_at)
VALUES ('user@example.com', 'Name', true, encode(gen_random_bytes(32), 'hex'), NOW(), NOW());
```

### Checking Recent Sends
```sql
-- View recent email campaigns
SELECT id, subject, sent_at, recipient_count 
FROM emails_sent 
ORDER BY sent_at DESC 
LIMIT 10;

-- View articles in a specific email
SELECT article_title, relevance_score, position
FROM email_articles
WHERE email_id = 42
ORDER BY position;

-- Check for duplicate article
SELECT ea.article_url, es.sent_at
FROM email_articles ea
JOIN emails_sent es ON ea.email_id = es.id
WHERE ea.article_url = 'https://example.com/article'
ORDER BY es.sent_at DESC;
```

---

## What NOT to Do

❌ **Don't add fallbacks to hardcoded data**
- No "if DB fails, use this array" logic
- Fail-fast is intentional

❌ **Don't add TO_EMAIL back**
- It was removed intentionally
- All recipients from database

❌ **Don't skip error checking**
- Every database call must check errors
- Every empty result must be handled

❌ **Don't create circular dependencies**
- main.go depends on database, feeds, ai, email
- Those packages should NOT depend on main

❌ **Don't modify database schema without GORM models**
- Always update `database/models.go` first
- Let AutoMigrate handle SQL

❌ **Don't send test emails to real users**
- Use `--dry-run` for testing
- Or create test user with your email

---

## Future Features (Not Yet Implemented)

These functions exist in database package but aren't used yet:

### Web Interface (Phase 4)
- `GetUserByToken()` - For unsubscribe page
- `UpdateUserSubscription()` - Handle unsubscribe

### Admin Dashboard
- `GetAllSources()` - View all sources
- `UpdateSourceActive()` - Enable/disable feeds
- `DeleteSource()` - Remove feeds

### Analytics
- `GetEmailByID()` - View campaign details
- `GetEmailArticles()` - Articles in campaign
- `GetUserEmails()` - User email history
- `MarkEmailOpened()` - Track opens

**Don't remove these!** They're ready for future implementation.

---

## Debugging Tips

### "No subscribed users found"
```bash
# Check users table
psql $DB_CONNECT_STRING -c "SELECT email, subscribed FROM users;"

# Add user if needed
cd scripts && go run add_user.go
```

### "No active sources found"
```bash
# Check sources table
psql $DB_CONNECT_STRING -c "SELECT count(*) FROM sources WHERE active = true;"

# Seed sources if needed
cd scripts && go run seed_sources.go
```

### "Failed to connect to database"
1. Check `DB_CONNECT_STRING` in .env
2. Verify PostgreSQL is running
3. Test connection: `psql $DB_CONNECT_STRING`
4. Check firewall if remote database

### Articles not being deduplicated
```sql
-- Check recent article URLs
SELECT article_url, created_at 
FROM email_articles 
WHERE created_at > NOW() - INTERVAL '30 days'
ORDER BY created_at DESC;
```

### Dry run emails in database
```sql
-- Find emails recorded but never sent
SELECT es.id, es.subject, es.sent_at
FROM emails_sent es
LEFT JOIN user_emails ue ON es.id = ue.email_id
WHERE ue.id IS NULL;
```

---

## Dependencies

### Main Dependencies
```go
github.com/joho/godotenv        // .env file loading
github.com/mmcdole/gofeed       // RSS parsing
github.com/sendgrid/sendgrid-go // Email sending
google.golang.org/genai         // Gemini AI
gorm.io/gorm                    // ORM
gorm.io/driver/postgres         // PostgreSQL driver
```

### Adding New Dependencies
```bash
go get package-name
go mod tidy
```

**Always:** Test that project still builds after adding dependencies.

---

## Documentation Files

When making changes, update relevant docs:

- **README.md** - Main project overview, setup, usage
- **docs/DATABASE_SETUP.md** - Database schema, setup guide
- **docs/DRY_RUN_FEATURE.md** - Dry-run flag documentation
- **docs/NEXT_STEPS.md** - Roadmap, completed features
- **docs/DATABASE_FUNCTION_USAGE.md** - Function usage analysis
- **scripts/README.md** - Script usage guide

If behavior changes significantly, update docs in the same commit.

---

## Code Style

### Naming Conventions
- Functions: `CamelCase` for exported, `camelCase` for private
- Variables: `camelCase` always
- Constants: `ALL_CAPS` or `CamelCase` (Go style)
- Files: `snake_case.go`

### Logging Format
```go
log.Println("Starting process...")           // Action starting
log.Printf("Processed %d items", count)      // Progress update
log.Println("✓ Process complete")            // Success
log.Fatalf("Failed to process: %v", err)     // Fatal error
log.Printf("Warning: %v", err)               // Non-fatal warning
```

### Comment Style
```go
// GetAllSubscribedUsers returns all users where subscribed = true
func GetAllSubscribedUsers() ([]User, error) {
    // ...
}
```

- Document all exported functions
- Explain "why" not "what" for complex logic
- Keep comments up-to-date with code

---

## Security Considerations

1. **Environment Variables**
   - Never commit `.env` file
   - Never log API keys or passwords
   - Use strong passwords for database

2. **Database**
   - Use prepared statements (GORM handles this)
   - Validate user input before SQL queries
   - Use SSL for remote database connections

3. **Unsubscribe Tokens**
   - Generated with `crypto/rand` (32 bytes)
   - Stored as hex string (64 characters)
   - Unique per user

4. **Email Sending**
   - Validate email addresses before adding users
   - Honor unsubscribe requests immediately
   - Include unsubscribe link in emails (future)

---

## Performance Notes

1. **RSS Fetching**
   - Concurrent fetching (existing implementation)
   - Timeout per feed to avoid hanging

2. **Database**
   - All common queries have indexes
   - Use `GetRecentArticleURLs()` (returns map) not `GetRecentEmailArticles()` (returns full structs)
   - Connection pooling handled by GORM automatically

3. **AI Analysis**
   - Rate limiting configurable via `GEMINI_RATE_LIMIT_MS`
   - Default: 200ms between requests
   - Consider batch scoring in future

4. **Email Sending**
   - Currently sends sequentially to each user
   - SendGrid handles rate limiting
   - Consider batch API in future for many users

---

## Version History

- **October 2025** - Initial database integration (Phase 1-3)
  - PostgreSQL with GORM
  - Multi-user support
  - Duplicate prevention (30 days)
  - Removed TO_EMAIL, removed hardcoded fallbacks
  - Added --dry-run flag

- **Earlier** - Original single-user implementation
  - Hardcoded TO_EMAIL recipient
  - Hardcoded RSS sources
  - No duplicate tracking

---

## Support & Resources

- **Documentation:** See `docs/` directory
- **Scripts:** See `scripts/README.md`
- **Database:** See `docs/DATABASE_SETUP.md`
- **Dry Run:** See `docs/DRY_RUN_FEATURE.md`

---

## Quick Reference Commands

```bash
# Build
go build -o thepaper

# Run normally
./thepaper

# Dry run (test mode)
./thepaper --dry-run

# Help
./thepaper -h

# Seed sources
cd scripts && go run seed_sources.go

# Add user
cd scripts && go run add_user.go

# Test database connection
psql $DB_CONNECT_STRING
```

---

## Final Notes for AI Agents

1. **Always test with `--dry-run` first** before making changes that affect email sending
2. **Respect the database-first architecture** - no hardcoded fallbacks
3. **Fail-fast is intentional** - don't add fallback logic without discussion
4. **Documentation is critical** - update docs when changing behavior
5. **Error messages should be actionable** - tell users how to fix issues
6. **Security matters** - never log credentials, validate input
7. **The codebase is small and focused** - keep it that way

**When in doubt:** Check existing patterns in the codebase. The project has consistent patterns for database access, error handling, and logging. Follow those patterns.

---

**Last Updated:** October 26, 2025
**Project Status:** Production-ready for multi-user newsletter sends
**Phase:** 3.5 complete (database-driven with dry-run), Phase 4 pending (web interface)