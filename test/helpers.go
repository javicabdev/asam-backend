package test

import (
	"fmt"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// GenerateValidNumeroSocio Helper para crear un número de socio válido
func GenerateValidNumeroSocio(index int) string {
	// Asegurar que el índice tenga 4 dígitos
	return fmt.Sprintf("A%04d", index)
}

// CreateValidFamily Helper para crear una familia válida
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

// CreateValidMember crea y retorna un objeto Member con datos válidos para usar en pruebas.
// El miembro creado incluye todos los campos requeridos con valores de ejemplo realistas.
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
