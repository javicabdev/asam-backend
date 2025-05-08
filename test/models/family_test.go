package models_test

import (
	"strings"
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test"

	"github.com/stretchr/testify/assert"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

func TestFamilyValidation_BasicFields(t *testing.T) {
	tests := []struct {
		name    string
		family  *models.Family
		wantErr bool
	}{
		{
			name: "valid family",
			family: &models.Family{
				NumeroSocio:              "F1234",
				EsposoNombre:             "Juan",
				EsposoApellidos:          "García",
				EsposaNombre:             "María",
				EsposaApellidos:          "López",
				EsposoDocumentoIdentidad: "12345678A",
				EsposaDocumentoIdentidad: "87654321B",
			},
			wantErr: false,
		},
		{
			name: "missing numero socio",
			family: &models.Family{
				EsposoNombre:    "Juan",
				EsposoApellidos: "García",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.family.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFamilyValidation_Conyuges(t *testing.T) {
	tests := []struct {
		name    string
		family  *models.Family
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid family data",
			family: &models.Family{
				NumeroSocio:              test.GenerateValidNumeroSocio(1),
				EsposoNombre:             "Juan",
				EsposoApellidos:          "García",
				EsposaNombre:             "María",
				EsposaApellidos:          "López",
				EsposoDocumentoIdentidad: "12345678A",
				EsposaDocumentoIdentidad: "87654321B",
			},
			wantErr: false,
		},
		{
			name: "missing spouse names",
			family: &models.Family{
				NumeroSocio:              test.GenerateValidNumeroSocio(2),
				EsposoDocumentoIdentidad: "12345678A",
				EsposaDocumentoIdentidad: "87654321B",
			},
			wantErr: true,
			errMsg:  "se requiere información de al menos un cónyuge",
		},
		{
			name: "only husband data",
			family: &models.Family{
				NumeroSocio:              test.GenerateValidNumeroSocio(3),
				EsposoNombre:             "Juan",
				EsposoApellidos:          "García",
				EsposoDocumentoIdentidad: "12345678A",
				EsposaDocumentoIdentidad: "87654321B",
			},
			wantErr: true,
			errMsg:  "se requiere información de al menos un cónyuge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.family.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errMsg))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Tests de validaciones de documentos de identidad
func TestFamilyValidation_DocumentIDs(t *testing.T) {
	family := test.CreateValidFamily()

	// Caso válido
	assert.NoError(t, family.Validate())

	// Caso: Documento de identidad del esposo inválido
	family.EsposoDocumentoIdentidad = "INVALID_DOC"
	err := family.Validate()
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err), "Debería ser un error de validación")

	// Verificar que el campo específico tiene error
	fields := errors.GetFields(err)
	assert.NotNil(t, fields, "Debería contener campos con error")
	assert.Contains(t, fields, "esposoDocumentoIdentidad", "Debería tener error en el documento del esposo")

	// Restaurar documento válido
	family.EsposoDocumentoIdentidad = "12345678A"

	// Caso: Documento de identidad de la esposa inválido
	family.EsposaDocumentoIdentidad = "INVALID_DOC"
	err = family.Validate()
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err), "Debería ser un error de validación")

	// Verificar que el campo específico tiene error
	fields = errors.GetFields(err)
	assert.NotNil(t, fields, "Debería contener campos con error")
	assert.Contains(t, fields, "esposaDocumentoIdentidad", "Debería tener error en el documento de la esposa")
}

// Tests de validaciones de información de contacto
func TestFamilyValidation_ContactInfo(t *testing.T) {
	family := test.CreateValidFamily()

	// Caso válido
	assert.NoError(t, family.Validate())

	// Caso: Correo electrónico inválido
	family.EsposoCorreoElectronico = "invalid-email"
	err := family.Validate()
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err), "Debería ser un error de validación")

	// Verificar que el campo específico tiene error
	fields := errors.GetFields(err)
	assert.NotNil(t, fields, "Debería contener campos con error")
	assert.Contains(t, fields, "esposoCorreoElectronico", "Debería tener error en el correo del esposo")

	// Restaurar correo válido
	family.EsposoCorreoElectronico = "pedro.lopez@example.com"

	// Caso: Correo electrónico de la esposa inválido
	family.EsposaCorreoElectronico = "invalid-email"
	err = family.Validate()
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err), "Debería ser un error de validación")

	// Verificar que el campo específico tiene error
	fields = errors.GetFields(err)
	assert.NotNil(t, fields, "Debería contener campos con error")
	assert.Contains(t, fields, "esposaCorreoElectronico", "Debería tener error en el correo de la esposa")
}

// Tests de validaciones de fechas
func TestFamilyValidation_Dates(t *testing.T) {
	family := test.CreateValidFamily()

	// Caso válido
	assert.NoError(t, family.Validate())

	// Caso: Fecha de nacimiento futura
	family.EsposoFechaNacimiento = test.TimePtr(time.Now().Add(24 * time.Hour))
	err := family.Validate()
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err), "Debería ser un error de validación")

	// Verificar que el campo específico tiene error
	fields := errors.GetFields(err)
	assert.NotNil(t, fields, "Debería contener campos con error")
	assert.Contains(t, fields, "esposoFechaNacimiento", "Debería tener error en la fecha de nacimiento")
}

// Tests de lógica de negocio
func TestFamily_NombreCompletoEsposo(t *testing.T) {
	family := test.CreateValidFamily()

	// Validar nombre completo del esposo
	expected := "Pedro López"
	actual := family.NombreCompletoEsposo()
	assert.Equal(t, expected, actual)
}

func TestFamily_NombreCompletoEsposa(t *testing.T) {
	family := test.CreateValidFamily()

	// Validar nombre completo de la esposa
	expected := "María García"
	actual := family.NombreCompletoEsposa()
	assert.Equal(t, expected, actual)
}

// Tests de hooks de GORM
func TestFamily_BeforeCreate(t *testing.T) {
	family := test.CreateValidFamily()

	// Caso válido
	assert.NoError(t, family.BeforeCreate(nil))

	// Caso inválido - numero socio vacío
	family.NumeroSocio = ""
	assert.Error(t, family.BeforeCreate(nil))

	// Restauramos numero socio y probamos documento identidad vacío
	family.NumeroSocio = "A1234"
	family.EsposoDocumentoIdentidad = ""
	err := family.BeforeCreate(nil)
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err), "Debería ser un error de validación")

	// Verificar que el campo específico tiene error
	fields := errors.GetFields(err)
	assert.NotNil(t, fields, "Debería contener campos con error")
	assert.Contains(t, fields, "esposoDocumentoIdentidad", "Debería tener error en el documento de identidad del esposo")
}

func TestFamily_BeforeUpdate(t *testing.T) {
	// Caso 1: Familia válida
	family := &models.Family{
		NumeroSocio:              test.GenerateValidNumeroSocio(1),
		EsposoNombre:             "Juan",
		EsposoApellidos:          "García",
		EsposaNombre:             "María",
		EsposaApellidos:          "López",
		EsposoDocumentoIdentidad: "12345678A",
		EsposaDocumentoIdentidad: "87654321B",
	}
	assert.NoError(t, family.BeforeUpdate(nil))

	// Caso 2: Número de socio inválido (debe generar error)
	family.NumeroSocio = "invalid"
	err := family.BeforeUpdate(nil)
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err), "Debería ser un error de validación")

	// Verificar que el campo específico tiene error
	fields := errors.GetFields(err)
	assert.NotNil(t, fields, "Debería contener campos con error")
	assert.Contains(t, fields, "numeroSocio", "Debería tener error en el número de socio")

	// Caso 3: Sin cónyuges (debe generar error)
	family = &models.Family{
		NumeroSocio:              test.GenerateValidNumeroSocio(2),
		EsposoDocumentoIdentidad: "12345678A",
		EsposaDocumentoIdentidad: "87654321B",
	}
	err = family.BeforeUpdate(nil)
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err), "Debería ser un error de validación")

	// Verificar los campos con error para cónyuges
	fields = errors.GetFields(err)
	assert.NotNil(t, fields, "Debería contener campos con error")
	// Debe tener al menos un error relacionado con los cónyuges
	conyugeFields := []string{"esposoNombre", "esposoApellidos", "esposaNombre", "esposaApellidos"}
	hasConyugeError := false
	for _, field := range conyugeFields {
		if _, exists := fields[field]; exists {
			hasConyugeError = true
			break
		}
	}
	assert.True(t, hasConyugeError, "Debería tener al menos un error relacionado con los cónyuges")
}
