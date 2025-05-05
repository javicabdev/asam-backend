package test

import (
	"fmt"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"time"
)

// Helper para crear un número de socio válido
func GenerateValidNumeroSocio(index int) string {
	// Asegurar que el índice tenga 4 dígitos
	return fmt.Sprintf("A%04d", index)
}

// Helper para crear datos válidos de familia para tests
func CreateValidFamilyInput() model.CreateFamilyInput {
	return model.CreateFamilyInput{
		NumeroSocio:              GenerateValidNumeroSocio(1),
		EsposoNombre:             "Juan",
		EsposoApellidos:          "García",
		EsposaNombre:             "María",
		EsposaApellidos:          "López",
		EsposoDocumentoIdentidad: StringPtr("12345678A"),
		EsposaDocumentoIdentidad: StringPtr("87654321B"),
		EsposoCorreoElectronico:  StringPtr("juan@test.com"),
		EsposaCorreoElectronico:  StringPtr("maria@test.com"),
	}
}

// Helper para crear datos válidos de miembro para tests
func CreateValidMemberInput() model.CreateMemberInput {
	return model.CreateMemberInput{
		NumeroSocio:     GenerateValidNumeroSocio(2),
		TipoMembresia:   model.MembershipTypeIndividual,
		Nombre:          "Pedro",
		Apellidos:       "Martínez",
		CalleNumeroPiso: "Calle Test 1",
		CodigoPostal:    "08001",
		Poblacion:       "Barcelona",
		Provincia:       StringPtr("Barcelona"),
		Pais:            StringPtr("España"),
	}
}

// Helper para crear una familia válida
func CreateValidFamily() *models.Family {
	return &models.Family{
		NumeroSocio:              "A1234", // Cumple con la validación de formato
		EsposoNombre:             "Pedro",
		EsposoApellidos:          "López",
		EsposaNombre:             "María",
		EsposaApellidos:          "García",
		EsposoFechaNacimiento:    TimePtr(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)),
		EsposaFechaNacimiento:    TimePtr(time.Date(1985, 1, 1, 0, 0, 0, 0, time.UTC)),
		EsposoDocumentoIdentidad: "12345678A",         // Documento de identidad válido
		EsposaDocumentoIdentidad: "87654321B",         // Documento de identidad válido
		EsposoCorreoElectronico:  "pedro@example.com", // Correo válido
		EsposaCorreoElectronico:  "maria@example.com", // Correo válido
	}
}

func CreateValidMember() *models.Member {
	email := "test@example.com"
	return &models.Member{
		MembershipNumber: "B0001",
		MembershipType:   models.TipoMembresiaPIndividual,
		Name:             "Juan",
		Surnames:         "García",
		Address:          "Calle Test 1, 1º",
		Postcode:         "08224",
		City:             "Terrassa",
		Province:         "Barcelona",
		Country:          "España",
		State:            models.EstadoActivo,
		RegistrationDate: time.Now().Add(-24 * time.Hour),
		Nationality:      "Senegal",
		Email:            &email,
	}
}
