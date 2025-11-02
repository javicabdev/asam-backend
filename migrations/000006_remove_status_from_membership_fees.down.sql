-- Rollback: Add status column back to membership_fees table
-- Note: This rollback is provided for safety, but the status field
-- should not be used as it was causing incorrect behavior

-- Add the status column back
ALTER TABLE membership_fees ADD COLUMN status VARCHAR(20) DEFAULT 'pending';

-- Recreate the index
CREATE INDEX idx_membership_fees_status ON membership_fees(status);

-- Set all existing records to 'pending' status
UPDATE membership_fees SET status = 'pending' WHERE status IS NULL;
