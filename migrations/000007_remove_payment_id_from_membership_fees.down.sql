-- Rollback: Add payment_id column back to membership_fees table
-- Note: This rollback is provided for safety, but the payment_id field
-- should not be used as it creates an incorrect relationship design

-- Add the payment_id column back
ALTER TABLE membership_fees ADD COLUMN IF NOT EXISTS payment_id INTEGER;

-- Recreate the foreign key constraint
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'membership_fees_payment_id_fkey'
    ) THEN
        ALTER TABLE membership_fees
        ADD CONSTRAINT membership_fees_payment_id_fkey
        FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL;
        RAISE NOTICE 'Created foreign key constraint membership_fees_payment_id_fkey';
    ELSE
        RAISE NOTICE 'Foreign key constraint membership_fees_payment_id_fkey already exists';
    END IF;
END $$;

-- Note: No data restoration because this field was never populated in production
