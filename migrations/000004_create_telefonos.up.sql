CREATE TABLE telefonos (
                           telefono_id SERIAL PRIMARY KEY,
                           numero_telefono VARCHAR(20) NOT NULL,

    -- Patrón polimórfico
                           contactable_id INTEGER NOT NULL,
                           contactable_type VARCHAR(50) NOT NULL,

                           created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                           updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

                           CONSTRAINT uk_telefono_numero UNIQUE (numero_telefono)
);

-- Trigger para actualizar updated_at
CREATE TRIGGER update_telefonos_updated_at
    BEFORE UPDATE ON telefonos
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Índices
CREATE INDEX idx_telefonos_contactable ON telefonos(contactable_id, contactable_type);
CREATE INDEX idx_telefonos_numero ON telefonos(numero_telefono);