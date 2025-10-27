# Dry Run Feature - Quick Summary

## What Was Added

A `--dry-run` flag that allows you to preview the newsletter without sending emails.

## Usage

```bash
# Normal send
./thepaper

# Preview mode (no emails sent)
./thepaper --dry-run
```

## What It Does

### ✅ Everything Normal
- Fetches articles from all RSS sources
- Filters duplicates (last 30 days)
- Uses Gemini AI to analyze and score articles
- Selects top 8 articles
- Generates AI summaries
- Creates email record in database
- Saves articles to database

### ❌ Skips Sending
- Does NOT send emails via SendGrid
- Does NOT create user_emails records
- Shows comprehensive summary instead

## Example Output

```
🔍 DRY RUN MODE - No emails will be sent
...
============================================================
🔍 DRY RUN SUMMARY
============================================================
📊 Total articles fetched: 758
📰 Unique sources: 45
👥 Subscribed users: 3
⭐ Top articles selected: 8

📧 Would send to:
  • tyler@tylerevan.dev (Tyler)
  • john@example.com (John)
  • sarah@example.com (Sarah)

📝 Selected articles:
  1. [9.5] AI Breakthrough: New Language Model...
     Source: TechCrunch | Category: AI & Machine Learning
  2. [9.2] Rust 1.75 Released with Major Performance...
     Source: Rust Blog | Category: Programming Languages
  [...]

============================================================
✅ Dry run complete - no emails sent
============================================================
```

## Why Use It?

✅ **Test configuration** - Verify database, sources, and users
✅ **Preview content** - See what articles would be sent
✅ **Debug issues** - Check source distribution and article selection
✅ **Avoid accidents** - Test cron jobs without spamming users
✅ **Check quality** - Review article scores before sending

## Database Impact

**Writes:**
- ✅ Creates `emails_sent` record
- ✅ Creates `email_articles` records

**Doesn't Write:**
- ❌ No `user_emails` records (not marked as sent)

This maintains duplicate tracking while avoiding actual sends.

## Code Changes

**File:** `main.go`

**Added:**
- `flag` package import
- `--dry-run` flag parsing
- Dry run mode check at startup
- Dry run summary output before sending
- Early return to skip sending

**Lines:** ~40 lines added

## Documentation

- ✅ `docs/DRY_RUN_FEATURE.md` - Complete documentation
- ✅ `README.md` - Usage examples
- ✅ `docs/COMPLETED_WORK.md` - Feature documented

## Testing

```bash
# Build
go build -o thepaper
✅ Success

# Test help
./thepaper -h
✅ Shows dry-run flag

# Test dry run
./thepaper --dry-run
✅ Shows summary, no emails sent
```

## Cost

**Dry Run Mode:**
- Gemini API: ~$0.01 (analyzes articles)
- SendGrid API: $0 (not called)

Same cost as normal run except SendGrid.

## Perfect For

- Morning preview, evening send workflow
- Testing new RSS sources
- Verifying user list
- Debugging article selection
- Cron job testing

---
**Added:** October 26, 2025
**Build Status:** ✅ Tested and working
**Documentation:** ✅ Complete
