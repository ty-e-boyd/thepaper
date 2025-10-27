# The Paper

A daily programming news aggregator that fetches articles from RSS feeds, uses Gemini AI to select the most relevant content, and sends a curated HTML email digest via SendGrid.

## Features

- 📊 **Database-driven**: PostgreSQL with GORM for persistent storage
- 🔄 **Multi-user support**: Send to multiple subscribers from database
- 🚫 **Duplicate prevention**: Tracks sent articles (30-day window)
- 📡 **RSS aggregation**: Fetches from 81+ tech/programming feeds
- 🤖 **AI-powered curation**: Gemini AI scores and selects top articles
- ✍️ **Smart summaries**: AI-generated summaries for each article
- 📧 **Beautiful emails**: HTML digest via SendGrid
- 🎯 **Flexible sources**: Manage RSS feeds in database

## Setup

### 1. Prerequisites

- Go 1.24+
- PostgreSQL database (local or hosted)
- Google Gemini API key
- SendGrid API key

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Environment

Copy `.env.example` to `.env` and set these variables:

```bash
# Database
DB_CONNECT_STRING=postgresql://user:password@localhost:5432/thepaper

# AI
GEMINI_API_KEY=your_gemini_api_key
GEMINI_RATE_LIMIT_MS=200  # Optional: rate limit in milliseconds

# Email
SENDGRID_API_KEY=your_sendgrid_api_key
FROM_EMAIL=sender@example.com
```

### 4. Initialize Database

```bash
# Seed RSS feed sources (81 feeds)
cd scripts
go run seed_sources.go

# Add your first subscriber
go run add_user.go  # Edit script to customize email/name
```

See [DATABASE_SETUP.md](DATABASE_SETUP.md) for detailed database documentation.

## Usage

### Send Newsletter

```bash
# Run directly
go run main.go

# Or build and run
go build -o thepaper
./thepaper

# Dry run mode (preview without sending)
./thepaper --dry-run
```

**Dry Run Mode:**
Use `--dry-run` to preview what would be sent without actually sending emails:
- Fetches articles from all sources
- Analyzes and selects top articles
- Shows summary of articles, sources, and recipients
- Does NOT send any emails
- Perfect for testing configuration or checking what content will be sent

The application will:
1. Connect to database and run migrations
2. Fetch all subscribed users
3. Pull RSS sources from database
4. Fetch and filter articles (removes duplicates from last 30 days)
5. Use Gemini AI to select top 8 articles
6. Generate summaries
7. Send personalized email to each subscriber
8. Track send history in database

### Manage Database

```bash
# Add RSS sources to database
cd scripts && go run seed_sources.go

# Add a new subscriber
cd scripts && go run add_user.go
```

See [scripts/README.md](scripts/README.md) for more utility scripts.

## Configuration

- **Top articles**: Default is 8, modify `topArticlesCount` in `main.go`
- **Duplicate window**: Default is 30 days, modify in `main.go` (line 91)
- **Rate limiting**: Set `GEMINI_RATE_LIMIT_MS` in `.env`
- **Dry run**: Use `--dry-run` flag to preview without sending

## RSS Feeds

The application pulls RSS feeds **exclusively from the database**. 81 feeds are organized by category:

- **General Tech News**: TechCrunch, The Verge, Ars Technica, Wired, etc.
- **Hacker News**: frontpage, newest, show HN, ask HN
- **Programming & Development**: DEV.to, freeCodeCamp, Stack Overflow Blog
- **Web Development**: Smashing Magazine, CSS-Tricks, Codrops
- **Software Architecture**: Martin Fowler, DZone, Clean Coder
- **Language-Specific**: Rust, Go, Python, Node.js, Kotlin, Swift blogs
- **DevOps & Cloud**: AWS, GCP, Kubernetes, Docker, HashiCorp
- **FAANG Companies**: Netflix, Uber, Airbnb, Stripe, Meta, LinkedIn
- **Other Tech**: Dropbox, Cloudflare, Discord, Figma, Mozilla
- **Reddit**: r/programming, r/webdev, r/javascript, r/devops, etc.

### Managing Sources

**Important:** Sources must be in the database. The application will **not run** without sources.

**Add via SQL:**
```sql
INSERT INTO sources (name, category, url, active, created_at, updated_at)
VALUES ('Feed Name', 'Category', 'https://feed.url/rss', true, NOW(), NOW());
```

**Add via code:**
1. Edit `feeds/sources.go` (FeedSources map - used only for seeding)
2. Run `cd scripts && go run seed_sources.go`

**Deactivate a source:**
```sql
UPDATE sources SET active = false WHERE url = 'https://feed.url/rss';
```

**Note:** If no active sources exist in the database, the application will exit with an error.

## Project Structure

```
thepaper/
├── main.go                  # Entry point and orchestration
├── models/
│   └── types.go             # Data structures
├── config/
│   └── config.go            # Environment configuration
├── database/                # 🆕 Database layer (GORM)
│   ├── db.go                # Connection and migrations
│   ├── models.go            # GORM models (users, sources, emails)
│   ├── users.go             # User management functions
│   ├── sources.go           # RSS source management
│   └── emails.go            # Email tracking functions
├── feeds/
│   ├── sources.go           # RSS feed URLs (seeds database)
│   └── fetcher.go           # RSS feed fetching
├── ai/
│   └── analyzer.go          # Gemini-powered analysis
├── email/
│   ├── builder.go           # HTML email generation
│   └── sender.go            # SendGrid integration
└── scripts/                 # 🆕 Utility scripts
    ├── seed_sources.go      # Import RSS feeds to database
    ├── add_user.go          # Add subscribers
    └── README.md            # Script documentation
```

## Documentation

- **[DATABASE_SETUP.md](DATABASE_SETUP.md)** - Complete database setup guide
- **[NEXT_STEPS.md](NEXT_STEPS.md)** - Roadmap and completed features
- **[scripts/README.md](scripts/README.md)** - Utility script documentation

## How It Works

1. **Database Connection**: Connects to PostgreSQL and runs migrations
2. **User Management**: Fetches all subscribed users from database (exits if none found)
3. **Source Management**: Pulls active RSS feeds from database (exits if none found)
4. **Article Fetching**: Concurrently fetches from all sources
5. **Duplicate Prevention**: Filters articles sent in last 30 days
6. **AI Analysis**: Gemini scores articles for relevance (0-10)
7. **Selection**: Picks top 8 articles with diversity across sources
8. **Summarization**: AI generates concise summaries
9. **Email Generation**: Builds HTML digest with summaries and links
10. **Multi-User Send**: Sends to all subscribers, tracks in database (skipped in `--dry-run` mode)

**Note:** The application is 100% database-driven. It will not fall back to hardcoded sources or recipients.

**Dry Run Mode:** Use `./thepaper --dry-run` to preview the newsletter without sending emails. Perfect for testing!

## Database Schema

Five main tables:
- **users**: Subscribers with unsubscribe tokens
- **sources**: RSS feed sources by category
- **emails_sent**: Email campaign records
- **email_articles**: Articles included in each email (duplicate tracking)
- **user_emails**: Join table tracking who received what

See [DATABASE_SETUP.md](DATABASE_SETUP.md) for full schema details.

## Contributing

To add new features:
1. Database changes: Update models in `database/models.go`
2. Run application to trigger AutoMigrate
3. Add repository functions as needed
4. Update documentation

## License

MIT
```
