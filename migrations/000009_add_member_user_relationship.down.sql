-- Remove constraints first
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_user_requires_member;
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_admin_no_member;

-- Drop the unique index
DROP INDEX IF EXISTS idx_users_member_id;

-- Remove the member_id column
ALTER TABLE users DROP COLUMN IF EXISTS member_id;
