-- Restrict cash_flows.operation_type to the values of models.OperationType
-- (internal/domain/models/operation_type.go, AllOperationTypes). Until now the
-- column accepted any VARCHAR(20), so rows written with raw SQL (as the
-- since-removed test/seed package did) could store types the app and GraphQL
-- layer can't map.
-- IDEMPOTENT: Safe to run multiple times
--
-- The constraint is added NOT VALID, so it applies to every INSERT/UPDATE from
-- now on without checking existing rows, and is then validated only if every
-- existing row already complies. Before 2025-11-02 the domain used lowercase
-- types (ingreso_cuota, gasto_corriente, entrega_fondo, otros_ingresos) and no
-- migration converted them; a plain ADD CONSTRAINT would fail on such rows and
-- leave the migration dirty. If any exist, the constraint stays NOT VALID and a
-- WARNING is raised. Reclassify those rows, then run:
--   ALTER TABLE cash_flows VALIDATE CONSTRAINT cash_flows_operation_type_check;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'cash_flows_operation_type_check'
          AND conrelid = 'cash_flows'::regclass
    ) THEN
        ALTER TABLE cash_flows
            ADD CONSTRAINT cash_flows_operation_type_check CHECK (operation_type IN (
                -- INGRESOS (amount > 0)
                'INGRESO_CUOTA', 'INGRESO_DONACION', 'INGRESO_OTRO',
                -- GASTOS (amount < 0)
                'GASTO_REPATRIACION', 'GASTO_ADMINISTRATIVO', 'GASTO_BANCARIO',
                'GASTO_AYUDA', 'GASTO_OTRO'
            )) NOT VALID;

        RAISE NOTICE 'Added cash_flows_operation_type_check (NOT VALID)';
    ELSE
        RAISE NOTICE 'cash_flows_operation_type_check already exists, skipping';
    END IF;

    -- No-op if already validated. The inner block is a subtransaction, so a
    -- failed validation rolls back only itself, not the constraint added above.
    BEGIN
        ALTER TABLE cash_flows VALIDATE CONSTRAINT cash_flows_operation_type_check;
    EXCEPTION WHEN check_violation THEN
        RAISE WARNING 'cash_flows has rows whose operation_type is not a models.OperationType: '
            'cash_flows_operation_type_check stays NOT VALID (enforced for new writes only)';
    END;
END $$;

COMMENT ON COLUMN cash_flows.operation_type IS 'models.OperationType: INGRESO_* (amount > 0) or GASTO_* (amount < 0). Enforced by cash_flows_operation_type_check';
