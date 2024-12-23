CREATE TABLE familias (
                          familia_id SERIAL PRIMARY KEY,
                          numero_socio VARCHAR(50) UNIQUE NOT NULL,
                          miembro_origen_id INTEGER REFERENCES miembros(miembro_id),
                          esposo_nombre VARCHAR(100) NOT NULL,
                          esposo_apellidos VARCHAR(100) NOT NULL,
                          esposa_nombre VARCHAR(100) NOT NULL,
                          esposa_apellidos VARCHAR(100) NOT NULL,
                          esposo_fecha_nacimiento DATE,
                          esposo_documento_identidad VARCHAR(50),
                          esposo_correo_electronico VARCHAR(100),
                          esposa_fecha_nacimiento DATE,
                          esposa_documento_identidad VARCHAR(50),
                          esposa_correo_electronico VARCHAR(100),
                          created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Trigger para actualizar updated_at
CREATE TRIGGER update_familias_updated_at
    BEFORE UPDATE ON familias
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Índices
CREATE INDEX idx_familias_numero_socio ON familias(numero_socio);
CREATE INDEX idx_familias_miembro_origen ON familias(miembro_origen_id);
CREATE INDEX idx_familias_documentos ON familias(esposo_documento_identidad, esposa_documento_identidad);