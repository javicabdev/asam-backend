-- Rollback: Add family_id column back to payments and cash_flows tables
-- Note: This rollback is provided for safety, but family_id should not be used
-- as it creates redundant and potentially inconsistent data

-- Step 1: Make member_id nullable again in payments
ALTER TABLE payments ALTER COLUMN member_id DROP NOT NULL;
RAISE NOTICE 'Made member_id nullable in payments table';

-- Step 2: Add family_id back to cash_flows
ALTER TABLE cash_flows ADD COLUMN IF NOT EXISTS family_id INTEGER;
RAISE NOTICE 'Added family_id column to cash_flows table';

-- Step 3: Create index on cash_flows.family_id
CREATE INDEX IF NOT EXISTS idx_cash_flows_family_id ON cash_flows(family_id);
RAISE NOTICE 'Created index on cash_flows.family_id';

-- Step 4: Add family_id back to payments
ALTER TABLE payments ADD COLUMN IF NOT EXISTS family_id INTEGER;
RAISE NOTICE 'Added family_id column to payments table';

-- Step 5: Restore foreign key constraints
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'payments_family_id_fkey'
    ) THEN
        ALTER TABLE payments
        ADD CONSTRAINT payments_family_id_fkey
        FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE SET NULL;
        RAISE NOTICE 'Created foreign key constraint payments_family_id_fkey';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'cash_flows_family_id_fkey'
    ) THEN
        ALTER TABLE cash_flows
        ADD CONSTRAINT cash_flows_family_id_fkey
        FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE SET NULL;
        RAISE NOTICE 'Created foreign key constraint cash_flows_family_id_fkey';
    END IF;
END $$;

-- Note: Data cannot be fully restored as we don't know which payments/cashflows
-- originally had family_id populated
