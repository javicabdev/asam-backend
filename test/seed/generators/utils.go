package generators

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Utility functions for data generation

// Common Spanish first names
var firstNamesMale = []string{
	"Antonio", "Manuel", "José", "Francisco", "David", "Juan", "Carlos", "Jesús",
	"Miguel", "Alejandro", "Rafael", "Pedro", "Fernando", "Pablo", "Luis", "Alberto",
	"Sergio", "Jorge", "Alberto", "Álvaro", "Diego", "Adrián", "Raúl", "Iván",
	"Mamadou", "Abdoulaye", "Ibrahima", "Ousmane", "Cheikh", "Moussa", "Modou", "Boubacar",
}

var firstNamesFemale = []string{
	"María", "Carmen", "Josefa", "Isabel", "Ana", "Laura", "Dolores", "Pilar",
	"Cristina", "Lucía", "Mercedes", "Teresa", "Rosa", "Paula", "Raquel", "Manuela",
	"Sara", "Marta", "Sofía", "Ainhoa", "Aitana", "Alba", "Nerea", "Claudia",
	"Fatou", "Aminata", "Aissatou", "Mariama", "Adama", "Binta", "Khady", "Rama",
}

// Common Spanish last names
var lastNames = []string{
	"García", "Rodríguez", "González", "Fernández", "López", "Martínez", "Sánchez", "Pérez",
	"Gómez", "Martín", "Jiménez", "Ruiz", "Hernández", "Díaz", "Moreno", "Álvarez",
	"Romero", "Alonso", "Gutiérrez", "Navarro", "Torres", "Domínguez", "Vázquez", "Ramos",
	"Diop", "Ndiaye", "Fall", "Mbaye", "Gueye", "Sow", "Diallo", "Thiam", "Ba", "Cissé",
}

// Common Spanish street names
var streetNames = []string{
	"Calle Mayor", "Avenida de la Constitución", "Calle del Sol", "Calle de la Luna",
	"Calle Real", "Plaza de España", "Calle Nueva", "Avenida de la Libertad",
	"Calle del Carmen", "Calle San Francisco", "Calle Alameda", "Calle del Río",
	"Paseo de la Castellana", "Calle Gran Vía", "Avenida Diagonal", "Calle Princesa",
	"Rambla Catalunya", "Passeig de Gràcia", "Carrer de Sants", "Carrer del Carme",
}

// Spanish cities
var cities = []string{
	"Barcelona", "Badalona", "Sabadell", "Terrassa", "Mataró", "Santa Coloma de Gramenet",
	"Cornellà de Llobregat", "Sant Boi de Llobregat", "Rubí", "Manresa", "Vilafranca del Penedès",
}

// Spanish provinces
var provinces = []string{
	"Barcelona", "Tarragona", "Lleida", "Girona",
}

// Helper to select random item from a slice
func randomItem(items []string, r *rand.Rand) string {
	return items[r.Intn(len(items))]
}

// GenerateRandomName generates a random full name
func GenerateRandomName(r *rand.Rand, gender string) (firstName string, lastName string) {
	if gender == "male" {
		firstName = randomItem(firstNamesMale, r)
	} else {
		firstName = randomItem(firstNamesFemale, r)
	}
	lastName = randomItem(lastNames, r)

	// 30% chance to have compound last name
	if r.Float64() < 0.3 {
		lastName = lastName + " " + randomItem(lastNames, r)
	}

	return firstName, lastName
}

// GenerateRandomAddress generates a random address
func GenerateRandomAddress(r *rand.Rand) (address string, postalCode string, city string, province string) {
	street := randomItem(streetNames, r)
	num := r.Intn(150) + 1
	floor := ""

	// 60% chance to have floor info
	if r.Float64() < 0.6 {
		floorNum := r.Intn(8) + 1
		doorOptions := []string{"A", "B", "C", "1", "2", "3"}
		door := randomItem(doorOptions, r)
		floor = fmt.Sprintf(", %dº %s", floorNum, door)
	}

	address = fmt.Sprintf("%s, %d%s", street, num, floor)
	postalCode = fmt.Sprintf("%d", 8000+r.Intn(100)) // Barcelona postal codes 08001-08099
	city = randomItem(cities, r)
	province = randomItem(provinces, r)

	return address, postalCode, city, province
}

// GenerateRandomDNI generates a random Spanish ID
func GenerateRandomDNI(r *rand.Rand) string {
	numPart := r.Intn(90000000) + 10000000 // 8 digits

	// Letters for Spanish DNI (excluding I, O, U)
	dniLetters := "TRWAGMYFPDXBNJZSQVHLCKE"

	// Calculate check letter
	checkIndex := numPart % 23
	checkLetter := string(dniLetters[checkIndex])

	return fmt.Sprintf("%d%s", numPart, checkLetter)
}

// GenerateRandomNIE generates a random foreign ID (NIE)
func GenerateRandomNIE(r *rand.Rand) string {
	prefixes := []string{"X", "Y", "Z"}
	prefix := randomItem(prefixes, r)

	numPart := r.Intn(9000000) + 1000000 // 7 digits

	// Letters for NIE (excluding I, O, U)
	nieLetters := "TRWAGMYFPDXBNJZSQVHLCKE"

	// Calculate check letter - special calculation for NIE
	var checkIndex int
	switch prefix {
	case "X":
		checkIndex = (0*10000000 + numPart) % 23
	case "Y":
		checkIndex = (1*10000000 + numPart) % 23
	case "Z":
		checkIndex = (2*10000000 + numPart) % 23
	}

	checkLetter := string(nieLetters[checkIndex])

	return fmt.Sprintf("%s%d%s", prefix, numPart, checkLetter)
}

// GenerateRandomEmail generates a random email address based on name
func GenerateRandomEmail(r *rand.Rand, firstName, lastName string) string {
	domains := []string{"gmail.com", "hotmail.com", "yahoo.es", "outlook.com", "protonmail.com"}
	domain := randomItem(domains, r)

	// Normalize names for email
	firstName = strings.ToLower(firstName)
	firstName = strings.ReplaceAll(firstName, " ", "")
	lastName = strings.ToLower(lastName)
	lastName = strings.ReplaceAll(lastName, " ", "")

	// Remove accents
	replacements := map[string]string{
		"á": "a", "é": "e", "í": "i", "ó": "o", "ú": "u", "ü": "u",
		"ñ": "n", "ç": "c", "à": "a", "è": "e", "ì": "i", "ò": "o", "ù": "u",
	}

	for old, new := range replacements {
		firstName = strings.ReplaceAll(firstName, old, new)
		lastName = strings.ReplaceAll(lastName, old, new)
	}

	// Email generation types
	emailTypes := []int{0, 1, 2, 3, 4}
	emailType := emailTypes[r.Intn(len(emailTypes))]

	var email string
	switch emailType {
	case 0:
		email = fmt.Sprintf("%s.%s@%s", firstName, lastName, domain)
	case 1:
		email = fmt.Sprintf("%s%s@%s", firstName, lastName, domain)
	case 2:
		email = fmt.Sprintf("%s.%s%d@%s", firstName, lastName, r.Intn(99)+1, domain)
	case 3:
		email = fmt.Sprintf("%s_%s@%s", firstName, lastName, domain)
	case 4:
		email = fmt.Sprintf("%s%d@%s", firstName, r.Intn(999)+1, domain)
	}

	return email
}

// GenerateRandomPhone generates a random Spanish phone number
func GenerateRandomPhone(r *rand.Rand) string {
	prefixes := []string{"6", "7", "9"}
	prefix := randomItem(prefixes, r)

	number := r.Intn(90000000) + 10000000 // 8 digits
	return fmt.Sprintf("%s%d", prefix, number)
}

// GenerateRandomProfession generates a random profession
func GenerateRandomProfession(r *rand.Rand) string {
	professions := []string{
		"Professor/a", "Médico/a", "Abogado/a", "Ingeniero/a", "Comerciante",
		"Administrativo/a", "Cocinero/a", "Conductor/a", "Electricista", "Empresario/a",
		"Fontanero/a", "Funcionario/a", "Mecánico/a", "Obrero/a", "Peluquero/a",
		"Autónomo/a", "Dependiente/a", "Estudiante", "Jubilado/a", "Desempleado/a",
	}

	return randomItem(professions, r)
}

// GenerateRandomDate generates a random date within a range
func GenerateRandomDate(r *rand.Rand, start, end time.Time) time.Time {
	delta := end.Sub(start)
	deltaDays := int(delta.Hours() / 24)
	if deltaDays <= 0 {
		return start
	}

	days := r.Intn(deltaDays)
	return start.AddDate(0, 0, days)
}

// GenerateRandomAmount generates a random amount within a range with 2 decimal places
func GenerateRandomAmount(r *rand.Rand, minVal, maxVal float64) float64 {
	amount := minVal + r.Float64()*(maxVal-minVal)
	// Round to 2 decimal places
	return float64(int(amount*100)) / 100
}

// GenerateRandomMembershipNumber generates a random membership number
func GenerateRandomMembershipNumber(_ *rand.Rand, prefix string, count int) string {
	return fmt.Sprintf("%s-%d", prefix, count)
}
