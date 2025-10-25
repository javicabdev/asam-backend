-- Añadir constraint UNIQUE al campo identity_card de members
-- Esto evita que se registren dos miembros con el mismo documento de identidad

-- Primero, eliminar cualquier duplicado existente (si los hay)
-- Mantener solo el registro más antiguo de cada DNI duplicado
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

-- Añadir el constraint UNIQUE
ALTER TABLE members 
ADD CONSTRAINT members_identity_card_unique 
UNIQUE (identity_card);

-- Crear índice para mejorar performance de búsquedas por DNI
CREATE INDEX idx_members_identity_card 
ON members(identity_card) 
WHERE identity_card IS NOT NULL;
