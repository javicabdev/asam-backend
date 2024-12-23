CREATE TABLE familiares (
                            familiar_id SERIAL PRIMARY KEY,
                            familia_id INTEGER NOT NULL REFERENCES familias(familia_id) ON DELETE CASCADE,
                            nombre VARCHAR(100) NOT NULL,
                            dni_nie VARCHAR(50),
                            fecha_nacimiento DATE,
                            correo_electronico VARCHAR(100),
                            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Trigger para actualizar updated_at
CREATE TRIGGER update_familiares_updated_at
    BEFORE UPDATE ON familiares
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Índices
CREATE INDEX idx_familiares_familia ON familiares(familia_id);
CREATE INDEX idx_familiares_dni ON familiares(dni_nie);