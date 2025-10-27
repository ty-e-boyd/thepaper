# Newsletter Subscription Setup Guide

This guide explains how the subscribe/unsubscribe functionality works across the **thepaper** newsletter system and the **tyler-portfolio** website.

## Overview

The subscription system is split across two projects:

1. **thepaper** - The newsletter engine that sends daily digests
2. **tyler-portfolio** - The public-facing subscription interface

Both projects share the same PostgreSQL database and `users` table.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         User Flow                            │
└─────────────────────────────────────────────────────────────┘

1. User visits portfolio.com/subscribe
2. Enters email → POST /api/subscribe
3. User record created in shared database
4. thepaper reads subscribers from database
5. Sends email with personalized unsubscribe link
6. User clicks unsubscribe → portfolio.com/unsubscribe?token=xxx
7. Subscription status updated in database
```

## Database Schema

Both projects use the same `users` table:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,        -- Required field
    name VARCHAR(255),                         -- Optional field
    subscribed BOOLEAN DEFAULT true,
    unsubscribe_token VARCHAR(64) UNIQUE NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

**Key Points:**
- Only `email` is required
- `name` field is optional (nullable)
- Each user gets a unique 64-character unsubscribe token
- `subscribed` tracks active/inactive subscriptions

## Setup Instructions

### 1. Database Configuration

Both projects must connect to the same PostgreSQL database.

**In thepaper/.env:**
```bash
DB_CONNECT_STRING=postgresql://user:password@localhost:5432/thepaper?sslmode=disable
PORTFOLIO_URL=http://localhost:4040  # URL where portfolio is hosted
```

**In tyler-portfolio/.env:**
```bash
DB_CONNECT_STRING=postgresql://user:password@localhost:5432/thepaper?sslmode=disable
```

**Production Example:**
```bash
# thepaper/.env
DB_CONNECT_STRING=postgresql://user:password@prod-db.example.com:5432/thepaper?sslmode=require
PORTFOLIO_URL=https://tyboyd.dev

# tyler-portfolio/.env
DB_CONNECT_STRING=postgresql://user:password@prod-db.example.com:5432/thepaper?sslmode=require
```

### 2. Start the Portfolio Server

The portfolio must be running to handle subscriptions and unsubscribes:

```bash
cd go_projects/tyler-portfolio
go run main.go
```

Server starts on `http://localhost:4040`

### 3. Test the Subscription Flow

**Subscribe a new user:**

1. Visit `http://localhost:4040/subscribe`
2. Enter an email address
3. Click "Subscribe"

**Verify in database:**
```sql
SELECT email, subscribed, unsubscribe_token, created_at 
FROM users 
WHERE email = 'test@example.com';
```

**Add subscriber via thepaper script:**
```bash
cd thepaper/scripts
go run add_user.go
# Edit the script first to set email and name
```

### 4. Test the Newsletter Send

With subscribers in the database, run thepaper:

```bash
cd thepaper
go run main.go --dry-run  # Preview without sending
go run main.go            # Actually send emails
```

The emails will include an unsubscribe link:
```
https://tyboyd.dev/unsubscribe?token=abc123...
```

### 5. Test the Unsubscribe Flow

**From email link:**
1. Click the unsubscribe link in a test email
2. User is unsubscribed and shown confirmation page

**Manual test:**
```bash
# Get token from database
psql -d thepaper -c "SELECT unsubscribe_token FROM users WHERE email = 'test@example.com';"

# Visit unsubscribe URL
curl "http://localhost:4040/unsubscribe?token=YOUR_TOKEN_HERE"
```

**Verify unsubscription:**
```sql
SELECT email, subscribed FROM users WHERE email = 'test@example.com';
-- Should show subscribed = false
```

## API Reference

### Subscribe Endpoint

**POST** `/api/subscribe`

Request:
```json
{
  "email": "user@example.com"
}
```

Responses:

**Success (new user):**
```json
{
  "success": true,
  "message": "Successfully subscribed! You'll receive The Paper daily digest in your inbox."
}
```

**Already subscribed:**
```json
{
  "success": true,
  "message": "You're already subscribed to The Paper!"
}
```

**Resubscribed:**
```json
{
  "success": true,
  "message": "Welcome back! You've been resubscribed to The Paper."
}
```

**Error:**
```json
{
  "success": false,
  "error": "Email is required"
}
```

### Unsubscribe Endpoint (GET)

**GET** `/unsubscribe?token=<unsubscribe_token>`

- Validates token
- Updates `subscribed` to `false`
- Displays confirmation page with resubscribe option

### Unsubscribe Endpoint (POST)

**POST** `/api/unsubscribe`

Request:
```json
{
  "token": "unsubscribe_token_here"
}
```

Response:
```json
{
  "success": true,
  "message": "You have been successfully unsubscribed"
}
```

## Email Template Integration

The thepaper email builder now includes personalized unsubscribe links.

**Changes made:**

1. `email/builder.go` - New function `BuildHTMLWithToken()`
2. `main.go` - Passes user's unsubscribe token when building emails
3. Email footer includes unsubscribe link using `PORTFOLIO_URL`

**Example footer in email:**
```html
<div class="footer">
  <p>You're receiving this because you subscribed to The Paper daily digest.</p>
  <p>Curated and summarized by AI | Powered by Gemini</p>
  <p><a href="https://tyboyd.dev/unsubscribe?token=abc123...">
    Unsubscribe from this newsletter
  </a></p>
</div>
```

## Security Features

### Unsubscribe Tokens

- Generated using `crypto/rand` (cryptographically secure)
- 64 characters (32 bytes hex-encoded)
- Unique per user
- Never exposed in logs or API responses

### Token Generation Code

```go
func generateUnsubscribeToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}
```

### Database Security

- Unique constraints on email and token
- Soft deletes with `deleted_at` timestamp
- Index on `subscribed` for query performance
- SSL connections in production

## Common Issues

### Issue: Unsubscribe link shows localhost

**Problem:** Email shows `http://localhost:4040/unsubscribe?token=...`

**Solution:** Set `PORTFOLIO_URL` in thepaper/.env:
```bash
PORTFOLIO_URL=https://your-production-domain.com
```

### Issue: "DB_CONNECT_STRING environment variable is required"

**Problem:** Missing database configuration

**Solution:** Create `.env` file with:
```bash
DB_CONNECT_STRING=postgresql://user:password@localhost:5432/thepaper
```

### Issue: User subscribes but thepaper doesn't send to them

**Problem:** User has `subscribed = false` in database

**Solution:** Update subscription status:
```sql
UPDATE users SET subscribed = true WHERE email = 'user@example.com';
```

Or use the subscribe page to resubscribe.

### Issue: created_at and updated_at are NULL for users

**Problem:** Existing users have NULL timestamps

**Cause:** The User model was missing `CreatedAt` and `UpdatedAt` fields initially. This has been fixed, but existing records may still have NULL values.

**Solution 1 - Run the fix script:**
```bash
cd thepaper
psql -d thepaper -f scripts/fix_timestamps.sql
```

**Solution 2 - Manual SQL:**
```sql
UPDATE users 
SET 
    created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, NOW())
WHERE created_at IS NULL OR updated_at IS NULL;
```

**Verify the fix:**
```sql
SELECT email, created_at, updated_at FROM users WHERE created_at IS NULL OR updated_at IS NULL;
-- Should return no rows
```

**Prevention:** The User struct now includes proper timestamp fields. New users created after this fix will automatically have timestamps populated by GORM.

### Issue: Can't connect to database from portfolio

**Problem:** Portfolio and thepaper use different databases

**Solution:** Both must use the SAME database. Check connection strings match.

## Testing Checklist

- [ ] Portfolio starts without errors on port 4040
- [ ] Can access `/subscribe` page
- [ ] Can submit email and see success message
- [ ] User appears in database with `subscribed = true`
- [ ] User has unique unsubscribe token
- [ ] **User has non-NULL `created_at` and `updated_at` timestamps**
- [ ] thepaper dry-run shows the new subscriber
- [ ] Send test email (thepaper sends with unsubscribe link)
- [ ] Unsubscribe link points to correct domain
- [ ] Clicking unsubscribe shows confirmation page
- [ ] User's `subscribed` field updates to `false`
- [ ] **User's `updated_at` timestamp updates on unsubscribe**
- [ ] Can resubscribe from `/subscribe` page
- [ ] Resubscribed user gets `subscribed = true` again

## Production Deployment

### Portfolio (tyler-portfolio)

1. Build the application:
```bash
go build -o portfolio
```

2. Set production environment variables:
```bash
export DB_CONNECT_STRING="postgresql://user:pass@prod-db:5432/thepaper?sslmode=require"
```

3. Run with a process manager (systemd, supervisor, etc.)

4. Put behind reverse proxy (nginx, caddy) with SSL

### Newsletter (thepaper)

1. Update `.env` with production values:
```bash
DB_CONNECT_STRING=postgresql://user:pass@prod-db:5432/thepaper?sslmode=require
PORTFOLIO_URL=https://tyboyd.dev
GEMINI_API_KEY=your_prod_key
SENDGRID_API_KEY=your_prod_key
FROM_EMAIL=newsletter@tyboyd.dev
```

2. Schedule with cron:
```bash
# Run daily at 8 AM
0 8 * * * cd /path/to/thepaper && ./thepaper >> /var/log/thepaper.log 2>&1
```

### Database Backup

Regular backups are critical since both apps share the database:

```bash
# Daily backup
pg_dump -h prod-db.example.com -U user thepaper > backup-$(date +%Y%m%d).sql

# Restore if needed
psql -h prod-db.example.com -U user thepaper < backup-20240101.sql
```

## Monitoring

### Metrics to Track

1. **Subscription Rate**: New users per day
2. **Unsubscribe Rate**: Users unsubscribing per day
3. **Active Subscribers**: `SELECT COUNT(*) FROM users WHERE subscribed = true`
4. **Total Users**: `SELECT COUNT(*) FROM users`
5. **Resubscribe Rate**: Users who unsubscribe then resubscribe

### SQL Queries

**Daily subscription stats:**
```sql
SELECT 
    DATE(created_at) as date,
    COUNT(*) as new_subscribers
FROM users
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

**Unsubscribe rate:**
```sql
SELECT 
    COUNT(*) as total_users,
    SUM(CASE WHEN subscribed THEN 1 ELSE 0 END) as active,
    SUM(CASE WHEN NOT subscribed THEN 1 ELSE 0 END) as unsubscribed,
    ROUND(100.0 * SUM(CASE WHEN NOT subscribed THEN 1 ELSE 0 END) / COUNT(*), 2) as unsubscribe_rate
FROM users;
```

## Support

For issues or questions:

1. Check this documentation
2. Review logs in both applications
3. Verify database connectivity
4. Check environment variables
5. Review the code in `main.go` (both projects)

## Summary

The subscription system provides:

✅ Simple email-only subscription (no name required)  
✅ Secure unsubscribe tokens  
✅ One-click unsubscribe from emails  
✅ Resubscribe capability  
✅ Shared database architecture  
✅ Clean separation of concerns  
✅ Production-ready with proper error handling  

Both applications work together seamlessly to provide a complete newsletter subscription experience.