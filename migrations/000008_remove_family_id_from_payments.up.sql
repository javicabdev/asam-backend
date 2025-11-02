-- Migration: Remove family_id column from payments and cash_flows tables
-- Reason: The family_id field creates redundant data. Payments should always
-- be associated with a member (for families, use the origin member's ID).
--
-- Correct design:
-- - Individual member payment: payment.member_id → member.id
-- - Family member payment: payment.member_id → family.miembro_origen_id
-- - To get family payments: SELECT * FROM payments WHERE member_id = family.miembro_origen_id
--
-- Problems with family_id:
-- 1. Redundant: Family relationship already exists through Family.miembro_origen_id
-- 2. Inconsistent: Some payments have member_id, some have family_id, some both
-- 3. Complicates queries: Need to handle both member_id and family_id
-- 4. Data integrity: Can have conflicting member_id and family_id values

-- Step 1: Migrate existing data in payments table
-- Set member_id from family's origin member when payment only has family_id
DO $$
BEGIN
    -- Update payments that have family_id but no member_id
    UPDATE payments p
    SET member_id = f.miembro_origen_id
    FROM families f
    WHERE p.family_id = f.id
      AND p.member_id IS NULL
      AND f.miembro_origen_id IS NOT NULL;

    RAISE NOTICE 'Migrated family_id to member_id in payments table';
END $$;

-- Step 2: Drop family_id from payments table
ALTER TABLE payments DROP COLUMN IF EXISTS family_id;
RAISE NOTICE 'Dropped family_id column from payments table';

-- Step 3: Migrate existing data in cash_flows table
DO $$
BEGIN
    -- Update cash_flows that have family_id but no member_id
    -- Use the payment's member_id if payment exists, otherwise use family's origin member
    UPDATE cash_flows cf
    SET member_id = COALESCE(
        (SELECT member_id FROM payments p WHERE p.id = cf.payment_id),
        (SELECT miembro_origen_id FROM families f WHERE f.id = cf.family_id)
    )
    WHERE cf.family_id IS NOT NULL
      AND cf.member_id IS NULL;

    RAISE NOTICE 'Migrated family_id to member_id in cash_flows table';
END $$;

-- Step 4: Drop family_id from cash_flows table
ALTER TABLE cash_flows DROP COLUMN IF EXISTS family_id;
RAISE NOTICE 'Dropped family_id column from cash_flows table';

-- Step 5: Make member_id NOT NULL in payments (all payments must have a member)
ALTER TABLE payments ALTER COLUMN member_id SET NOT NULL;
RAISE NOTICE 'Made member_id NOT NULL in payments table';

-- Note: member_id in cash_flows can remain nullable for non-payment transactions
