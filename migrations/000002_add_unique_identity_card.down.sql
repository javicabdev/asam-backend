-- Rollback: Eliminar constraint UNIQUE y el índice del campo identity_card

-- Eliminar el índice
DROP INDEX IF EXISTS idx_members_identity_card;

-- Eliminar el constraint UNIQUE
ALTER TABLE members 
DROP CONSTRAINT IF EXISTS members_identity_card_unique;
