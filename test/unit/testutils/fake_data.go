package testutils

import (
	"fmt"
	"math/rand"
)

// Fake personal data for the test builders, moved from the removed
// test/seed/generators package.

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

var lastNames = []string{
	"García", "Rodríguez", "González", "Fernández", "López", "Martínez", "Sánchez", "Pérez",
	"Gómez", "Martín", "Jiménez", "Ruiz", "Hernández", "Díaz", "Moreno", "Álvarez",
	"Romero", "Alonso", "Gutiérrez", "Navarro", "Torres", "Domínguez", "Vázquez", "Ramos",
	"Diop", "Ndiaye", "Fall", "Mbaye", "Gueye", "Sow", "Diallo", "Thiam", "Ba", "Cissé",
}

var streetNames = []string{
	"Calle Mayor", "Avenida de la Constitución", "Calle del Sol", "Calle de la Luna",
	"Calle Real", "Plaza de España", "Calle Nueva", "Avenida de la Libertad",
	"Calle del Carmen", "Calle San Francisco", "Calle Alameda", "Calle del Río",
	"Paseo de la Castellana", "Calle Gran Vía", "Avenida Diagonal", "Calle Princesa",
	"Rambla Catalunya", "Passeig de Gràcia", "Carrer de Sants", "Carrer del Carme",
}

var cities = []string{
	"Barcelona", "Badalona", "Sabadell", "Terrassa", "Mataró", "Santa Coloma de Gramenet",
	"Cornellà de Llobregat", "Sant Boi de Llobregat", "Rubí", "Manresa", "Vilafranca del Penedès",
}

func randomItem(items []string, r *rand.Rand) string {
	return items[r.Intn(len(items))]
}

// randomName returns a random first name for the gender ("male" or anything
// else for female) and a last name, compound 30% of the time.
func randomName(r *rand.Rand, gender string) (firstName string, lastName string) {
	if gender == "male" {
		firstName = randomItem(firstNamesMale, r)
	} else {
		firstName = randomItem(firstNamesFemale, r)
	}
	lastName = randomItem(lastNames, r)

	if r.Float64() < 0.3 {
		lastName = lastName + " " + randomItem(lastNames, r)
	}

	return firstName, lastName
}

// randomAddress returns a random Barcelona-area street address, postal code
// and city.
func randomAddress(r *rand.Rand) (address string, postalCode string, city string) {
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

	return address, postalCode, city
}
