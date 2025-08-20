-- Add member_id column to users table
ALTER TABLE users 
ADD COLUMN member_id INTEGER REFERENCES members(id) ON DELETE RESTRICT;

-- Create unique index to ensure 1:1 relationship
CREATE UNIQUE INDEX idx_users_member_id ON users(member_id) 
WHERE member_id IS NOT NULL;

-- Add constraint: if role = 'user', member_id cannot be NULL
ALTER TABLE users ADD CONSTRAINT chk_user_requires_member
CHECK (role != 'user' OR member_id IS NOT NULL);

-- Add constraint: if role = 'admin', member_id must be NULL
ALTER TABLE users ADD CONSTRAINT chk_admin_no_member
CHECK (role != 'admin' OR member_id IS NULL);

-- Add comment explaining the column
COMMENT ON COLUMN users.member_id IS 
'ID of the associated member. NULL for admin users, required for regular users';
