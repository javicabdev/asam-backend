-- Añadir constraint UNIQUE al campo identity_card de members
-- Esto evita que se registren dos miembros con el mismo documento de identidad
-- IDEMPOTENT: Safe to run multiple times

-- Primero, eliminar cualquier duplicado existente (si los hay)
-- Mantener solo el registro más antiguo de cada DNI duplicado
-- Esta operación es idempotente: solo elimina duplicados si existen
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM members
        WHERE identity_card IS NOT NULL
        GROUP BY identity_card
        HAVING COUNT(*) > 1
    ) THEN
        WITH duplicates AS (
            SELECT
                identity_card,
                MIN(id) as keep_id
            FROM members
            WHERE identity_card IS NOT NULL
            GROUP BY identity_card
            HAVING COUNT(*) > 1
        )
        DELETE FROM members m
        WHERE m.identity_card IN (SELECT identity_card FROM duplicates)
          AND m.id NOT IN (SELECT keep_id FROM duplicates);

        RAISE NOTICE 'Cleaned up duplicate identity cards';
    ELSE
        RAISE NOTICE 'No duplicate identity cards found';
    END IF;
END $$;

-- Añadir el constraint UNIQUE (idempotente)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'members_identity_card_unique'
    ) THEN
        ALTER TABLE members 
        ADD CONSTRAINT members_identity_card_unique 
        UNIQUE (identity_card);
    END IF;
END $$;

-- Crear índice para mejorar performance de búsquedas por DNI (idempotente)
CREATE INDEX IF NOT EXISTS idx_members_identity_card 
ON members(identity_card) 
WHERE identity_card IS NOT NULL;
