-- Fix NULL timestamps in existing user records
-- Run this script if you have users with NULL created_at or updated_at values
--
-- Usage:
--   psql -d thepaper -f scripts/fix_timestamps.sql
--
-- Or from psql prompt:
--   \i scripts/fix_timestamps.sql

-- Show users with NULL timestamps before fixing
SELECT
    id,
    email,
    created_at,
    updated_at,
    CASE
        WHEN created_at IS NULL THEN 'NEEDS FIX'
        ELSE 'OK'
    END as created_status,
    CASE
        WHEN updated_at IS NULL THEN 'NEEDS FIX'
        ELSE 'OK'
    END as updated_status
FROM users
WHERE created_at IS NULL OR updated_at IS NULL;

-- Count how many records need fixing
SELECT
    COUNT(*) as total_users,
    COUNT(*) FILTER (WHERE created_at IS NULL) as null_created_at,
    COUNT(*) FILTER (WHERE updated_at IS NULL) as null_updated_at
FROM users;

-- Fix NULL timestamps
-- Sets them to NOW() if NULL, otherwise keeps existing value
UPDATE users
SET
    created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, NOW())
WHERE created_at IS NULL OR updated_at IS NULL;

-- Show how many records were updated
SELECT
    COUNT(*) as records_fixed
FROM users
WHERE created_at IS NOT NULL AND updated_at IS NOT NULL;

-- Verify all records now have timestamps
SELECT
    COUNT(*) as total_users,
    COUNT(*) FILTER (WHERE created_at IS NOT NULL) as has_created_at,
    COUNT(*) FILTER (WHERE updated_at IS NOT NULL) as has_updated_at
FROM users;

-- Show sample of fixed records
SELECT
    id,
    email,
    subscribed,
    created_at,
    updated_at
FROM users
ORDER BY id
LIMIT 10;
