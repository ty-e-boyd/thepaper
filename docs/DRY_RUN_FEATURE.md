# Dry Run Feature

## Overview

The `--dry-run` flag allows you to preview what the newsletter would contain without actually sending any emails. Perfect for testing, debugging, or checking content before sending.

## Usage

```bash
# Normal send
./thepaper

# Dry run (no emails sent)
./thepaper --dry-run
```

## What Happens in Dry Run Mode

### ✅ Everything Normal EXCEPT Sending
1. ✅ Connects to database
2. ✅ Pulls all subscribed users
3. ✅ Fetches articles from all RSS sources
4. ✅ Filters out duplicates (last 30 days)
5. ✅ Uses Gemini AI to score articles
6. ✅ Selects top 8 articles
7. ✅ Generates AI summaries
8. ✅ Creates email record in database
9. ✅ Saves selected articles to database
10. ❌ **SKIPS** sending emails
11. ✅ Shows comprehensive summary

### What Gets Saved to Database

**Normal Mode:**
- Email campaign record
- Articles in email
- User sends (one per recipient)

**Dry Run Mode:**
- Email campaign record ✅
- Articles in email ✅
- User sends ❌ (NOT created)

This means dry run creates the email record but doesn't mark it as sent to any users.

## Example Output

```
2025/10/26 20:00:00 🔍 DRY RUN MODE - No emails will be sent
2025/10/26 20:00:00 Connecting to database...
2025/10/26 20:00:00 ✓ Database connection established
2025/10/26 20:00:00 Loading configuration...
2025/10/26 20:00:00 Found 3 subscribed user(s)
2025/10/26 20:00:00 Fetching articles from 81 feeds from database across 12 categories...
2025/10/26 20:00:05 Fetched 1247 articles
2025/10/26 20:00:05 Filtered to 892 articles from last 24 hours
2025/10/26 20:00:05 Filtered out 134 duplicate articles sent in the last 30 days

2025/10/26 20:00:05 Article distribution by source:
  Hacker News: 145 articles
  TechCrunch: 23 articles
  The Verge: 18 articles
  [...]

2025/10/26 20:00:05 Analyzing articles with Gemini AI...
2025/10/26 20:00:45 Selected and summarized 8 top articles
2025/10/26 20:00:45 ✓ Email record created (ID: 42)
2025/10/26 20:00:45 ✓ Saved 8 articles to database

============================================================
🔍 DRY RUN SUMMARY
============================================================
📊 Total articles fetched: 758
📰 Unique sources: 45
👥 Subscribed users: 3
⭐ Top articles selected: 8

📧 Would send to:
  • tyler@tylerevan.dev (Tyler)
  • john@example.com (John Doe)
  • sarah@example.com (Sarah Smith)

📝 Selected articles:
  1. [9.5] AI Breakthrough: New Language Model Outperforms GPT-4
     Source: TechCrunch | Category: AI & Machine Learning
  2. [9.2] Rust 1.75 Released with Major Performance Improvements
     Source: Rust Blog | Category: Programming Languages
  3. [9.0] How Netflix Scales Their Microservices Architecture
     Source: Netflix Tech Blog | Category: Engineering
  4. [8.8] Understanding WebAssembly: A Practical Guide
     Source: Smashing Magazine | Category: Web Development
  5. [8.7] PostgreSQL 16: What's New and What Matters
     Source: PostgreSQL Blog | Category: Databases
  6. [8.5] GitHub Copilot Workspace: AI-Powered Development
     Source: GitHub Blog | Category: Developer Tools
  7. [8.3] Building Resilient Systems: Lessons from AWS
     Source: AWS Architecture Blog | Category: Cloud
  8. [8.1] The State of JavaScript 2024: Survey Results
     Source: State of JS | Category: JavaScript

============================================================
✅ Dry run complete - no emails sent
============================================================
```

## Use Cases

### 1. Testing Configuration
```bash
# Did I set up the database correctly?
./thepaper --dry-run
```

### 2. Content Preview
```bash
# What articles would go out today?
./thepaper --dry-run
```

### 3. Debugging
```bash
# Why aren't my sources showing up?
./thepaper --dry-run
# Check the "Article distribution by source" section
```

### 4. Schedule Testing
```bash
# Test your cron job without spamming users
0 9 * * * cd /app && ./thepaper --dry-run >> dry-run.log
```

### 5. Before Adding Users
```bash
# Make sure everything works before adding real subscribers
./thepaper --dry-run
```

## Comparison

| Feature | Normal Mode | Dry Run Mode |
|---------|-------------|--------------|
| Fetch articles | ✅ Yes | ✅ Yes |
| Filter duplicates | ✅ Yes | ✅ Yes |
| AI analysis | ✅ Yes | ✅ Yes |
| AI summaries | ✅ Yes | ✅ Yes |
| Create email record | ✅ Yes | ✅ Yes |
| Save articles | ✅ Yes | ✅ Yes |
| Send emails | ✅ Yes | ❌ No |
| Create user_emails | ✅ Yes | ❌ No |
| SendGrid API call | ✅ Yes | ❌ No |
| Show summary | ✅ Yes | ✅ Yes (detailed) |

## Database Impact

**Dry Run Still Writes to Database:**
- Creates record in `emails_sent` table
- Creates records in `email_articles` table
- Does NOT create records in `user_emails` table

**Why?** 
- Tracks what articles were analyzed
- Maintains duplicate prevention history
- Allows you to see "what would have been sent"

**Query to see dry run emails:**
```sql
SELECT es.id, es.subject, es.sent_at, es.recipient_count
FROM emails_sent es
LEFT JOIN user_emails ue ON es.id = ue.email_id
WHERE ue.id IS NULL;
```

This shows emails that were recorded but never sent (dry runs).

## Tips

### Daily Testing
```bash
# Morning: Check what would be sent
./thepaper --dry-run

# Evening: Send the actual newsletter
./thepaper
```

### Cron Job Example
```bash
# Dry run at 8 AM (preview)
0 8 * * * cd /app && ./thepaper --dry-run >> /var/log/thepaper-preview.log

# Real send at 9 AM
0 9 * * * cd /app && ./thepaper >> /var/log/thepaper.log
```

### Check Content Quality
```bash
# Run dry-run and grep for scores
./thepaper --dry-run 2>&1 | grep -A 20 "Selected articles:"
```

## Help Text

```bash
./thepaper -h
# or
./thepaper --help

Output:
  Usage of ./thepaper:
    -dry-run
        Run without sending emails (preview mode)
```

## Exit Codes

Both modes use the same exit codes:
- `0` - Success
- `1` - Error (database connection, no users, no sources, etc.)

## Cost Considerations

### Normal Mode
- Gemini API calls: ~$0.01 per run (analyzing articles)
- SendGrid API calls: Free tier (up to 100 emails/day)

### Dry Run Mode
- Gemini API calls: Same (~$0.01 per run)
- SendGrid API calls: $0 (not called)

**Dry run still costs API credits** because it analyzes articles with Gemini. Only SendGrid is skipped.

## Summary

`--dry-run` is perfect for:
- ✅ Testing before going live
- ✅ Previewing daily content
- ✅ Debugging issues
- ✅ Checking configuration
- ✅ Avoiding accidental sends

It does everything except send emails, giving you full visibility into what would happen.

---
**Added:** October 26, 2025
**Cost:** Same as normal run (Gemini API only)
**Database:** Partial write (no user_emails records)
