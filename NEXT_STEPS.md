# Next Steps - The Paper

## Database Implementation Plan

### Overview
Add persistent storage to track email history and manage user subscriptions. This will prevent duplicate articles and enable multi-user support.

---

## 1. Database Setup

### Technology Choice
- **PostgreSQL** (recommended) - Robust, good for relational data, excellent JSON support
- Alternative: **SQLite** (simpler, file-based, good for getting started)

### Schema Design

#### Tables Needed:

**1. `users`**
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    subscribed BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    unsubscribe_token VARCHAR(255) UNIQUE NOT NULL
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_subscribed ON users(subscribed);
CREATE INDEX idx_users_unsubscribe_token ON users(unsubscribe_token);
```

**2. `emails_sent`**
```sql
CREATE TABLE emails_sent (
    id SERIAL PRIMARY KEY,
    subject VARCHAR(500) NOT NULL,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    total_articles_analyzed INTEGER,
    total_sources INTEGER,
    recipient_count INTEGER
);

CREATE INDEX idx_emails_sent_sent_at ON emails_sent(sent_at);
```

**3. `email_articles`**
```sql
CREATE TABLE email_articles (
    id SERIAL PRIMARY KEY,
    email_id INTEGER REFERENCES emails_sent(id) ON DELETE CASCADE,
    article_url VARCHAR(1000) UNIQUE NOT NULL,
    article_title VARCHAR(500) NOT NULL,
    article_source VARCHAR(255),
    relevance_score DECIMAL(3,1),
    category VARCHAR(100),
    tags JSONB,
    summary TEXT,
    published_at TIMESTAMP,
    position INTEGER, -- Position in the email (1-8)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_email_articles_email_id ON email_articles(email_id);
CREATE INDEX idx_email_articles_url ON email_articles(article_url);
CREATE INDEX idx_email_articles_published_at ON email_articles(published_at);
```

**4. `user_emails`** (join table for tracking who received what)
```sql
CREATE TABLE user_emails (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    email_id INTEGER REFERENCES emails_sent(id) ON DELETE CASCADE,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    opened BOOLEAN DEFAULT false,
    opened_at TIMESTAMP,
    UNIQUE(user_id, email_id)
);

CREATE INDEX idx_user_emails_user_id ON user_emails(user_id);
CREATE INDEX idx_user_emails_email_id ON user_emails(email_id);
```

---

## 2. Implementation Tasks

### Phase 1: Database Integration
- [ ] Add database driver dependency (`github.com/lib/pq` for PostgreSQL)
- [ ] Create database configuration (add `DATABASE_URL` to .env)
- [ ] Create migration tool or SQL scripts for schema creation
- [ ] Add database connection package (`database/db.go`)
- [ ] Create models/repositories for each table

### Phase 2: Email History Tracking
- [ ] Before sending email, save email record to `emails_sent`
- [ ] Save selected articles to `email_articles` with foreign key to email
- [ ] Query `email_articles` to check if article URL was sent in last 30 days
- [ ] Filter out previously sent articles during selection process
- [ ] Add fallback logic if not enough new articles (expand time window or lower threshold)

### Phase 3: Multi-User Support
- [ ] Update `config/config.go` to support multiple recipients
- [ ] Modify email sending loop to iterate through subscribed users
- [ ] Record each send in `user_emails` table
- [ ] Add personalization (user's name in greeting)
- [ ] Implement unsubscribe token generation

### Phase 4: Subscription Management
- [ ] Create HTTP server with web endpoints
- [ ] Add `/subscribe` endpoint (form to add email)
- [ ] Add `/unsubscribe/:token` endpoint (one-click unsubscribe)
- [ ] Add `/resubscribe/:token` endpoint (opt back in)
- [ ] Update email template to include unsubscribe link in footer
- [ ] Add double opt-in confirmation email (optional but recommended)

---

## 3. Code Structure Changes

### New Packages to Create:

```
thepaper/
├── database/
│   ├── db.go              # Database connection & initialization
│   ├── migrations.go      # Schema migrations
│   └── repositories/
│       ├── users.go       # User CRUD operations
│       ├── emails.go      # Email history operations
│       └── articles.go    # Article tracking operations
├── web/
│   ├── server.go          # HTTP server setup
│   ├── handlers/
│   │   ├── subscribe.go   # Subscription handlers
│   │   └── unsubscribe.go # Unsubscription handlers
│   └── templates/
│       ├── subscribe.html
│       └── unsubscribe.html
└── utils/
    └── tokens.go          # Unsubscribe token generation
```

---

## 4. Updated Application Flow

### Current Flow:
```
1. Fetch articles from RSS feeds
2. Score articles with AI
3. Select top 8 articles
4. Summarize selected articles
5. Build HTML email
6. Send to single recipient
```

### New Flow:
```
1. Fetch articles from RSS feeds
2. Filter out articles sent in last 30 days (check database)
3. Score articles with AI
4. Select top 8 articles (with diversity + deduplication)
5. Summarize selected articles
6. Save email record to database
7. Save selected articles to database
8. Get all subscribed users from database
9. For each user:
   a. Build personalized HTML email
   b. Send via SendGrid
   c. Record send in user_emails table
10. Log summary (users reached, articles featured)
```

---

## 5. Configuration Updates

### New Environment Variables:
```bash
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/thepaper
# or for SQLite
DATABASE_PATH=./thepaper.db

# Web Server (for subscription management)
WEB_SERVER_PORT=8080
WEB_SERVER_HOST=0.0.0.0

# Application
DUPLICATE_WINDOW_DAYS=30  # How far back to check for duplicate articles
```

---

## 6. Migration Strategy

### Step 1: Add Database (Non-Breaking)
- Add database without removing single-user functionality
- Track emails being sent to existing single user
- Test thoroughly

### Step 2: Multi-User Support
- Keep `TO_EMAIL` as fallback
- Add user import script to seed initial users
- Switch to database-driven recipient list

### Step 3: Web Interface
- Launch subscription management endpoints
- Update email template with unsubscribe link
- Announce to users

---

## 7. Additional Features to Consider

### Analytics & Insights
- [ ] Track open rates (SendGrid webhooks)
- [ ] Track click-through rates on articles
- [ ] Generate weekly/monthly reports
- [ ] Dashboard showing most popular articles/sources

### Advanced Features
- [ ] User preferences (categories to include/exclude)
- [ ] Frequency preference (daily, weekly digest)
- [ ] Time zone support (send at optimal local time)
- [ ] A/B testing different article selections
- [ ] RSS feed for users who prefer readers
- [ ] Slack/Discord integration

### Admin Tools
- [ ] Admin dashboard to view stats
- [ ] Manually trigger email sends
- [ ] Preview next email before sending
- [ ] Blacklist/whitelist specific articles or sources

---

## 8. Testing Considerations

- [ ] Test with empty database (first run)
- [ ] Test with all articles previously sent (fallback behavior)
- [ ] Test unsubscribe flow end-to-end
- [ ] Test with multiple users (performance with 100, 1000, 10000 users)
- [ ] Test database connection failures (graceful degradation)
- [ ] Test SendGrid rate limits with many users

---

## 9. Deployment Considerations

### Database Hosting Options:
- **Heroku Postgres** (easy, free tier available)
- **Railway** (modern, generous free tier)
- **Supabase** (PostgreSQL + additional features)
- **AWS RDS** (scalable, production-ready)
- **Neon** (serverless Postgres)

### Application Hosting:
- Keep existing cron job setup
- Add web server for subscription management (can be separate dyno/service)
- Consider using background job queue for email sending (if user count grows large)

---

## 10. Legal/Compliance

### Email Compliance (CAN-SPAM Act, GDPR)
- [ ] Add physical mailing address to email footer
- [ ] Include clear unsubscribe link in every email
- [ ] Honor unsubscribe requests immediately
- [ ] Add privacy policy page
- [ ] Implement data export (GDPR right to data portability)
- [ ] Implement data deletion (GDPR right to be forgotten)

---

## Priority Order

**High Priority:**
1. Database setup with schema
2. Email history tracking (prevent duplicate articles)
3. Multi-user support (send to multiple recipients)
4. Unsubscribe functionality

**Medium Priority:**
5. Subscription web interface
6. User preferences
7. Open/click tracking

**Low Priority:**
8. Analytics dashboard
9. Advanced features (A/B testing, etc.)
10. Admin tools

---

## Getting Started

1. Choose database (recommend PostgreSQL)
2. Set up local database
3. Create schema using SQL scripts above
4. Add `database` package with connection logic
5. Update `main.go` to check for duplicate articles
6. Test with existing single-user setup
7. Gradually add multi-user support

Good luck! 🚀