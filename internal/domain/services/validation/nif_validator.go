package validation

import (
	"regexp"
	"strconv"
	"strings"
)

// ValidarNIF valida un DNI o NIE español después de normalizar la entrada
func ValidarNIF(nif string) bool {
	// Si el campo está vacío, considerarlo válido (es opcional)
	if nif == "" {
		return true
	}

	// Normalizar: eliminar todos los espacios y guiones
	normalized := strings.ReplaceAll(nif, " ", "")
	normalized = strings.ReplaceAll(normalized, "-", "")

	// Convertir a mayúsculas para hacer la validación insensible a mayúsculas/minúsculas
	normalized = strings.ToUpper(normalized)

	// Verificar longitud: debe ser exactamente 9 caracteres
	if len(normalized) != 9 {
		return false
	}

	// Letras de control según el algoritmo oficial
	const letrasControl = "TRWAGMYFPDXBNJZSQVHLCKE"

	// Determinar si es DNI o NIE y extraer el número y la letra de control
	var numero int
	var letraProporcionada string

	// Verificar si es un NIE (empieza con X, Y o Z)
	if normalized[0] == 'X' || normalized[0] == 'Y' || normalized[0] == 'Z' {
		// Es un NIE
		// Verificar que los siguientes 7 caracteres son dígitos
		digitosNIE := normalized[1:8]
		if !esNumerico(digitosNIE) {
			return false
		}

		// Convertir la letra inicial a número según el algoritmo
		var valorInicial int
		switch normalized[0] {
		case 'X':
			valorInicial = 0
		case 'Y':
			valorInicial = 1
		case 'Z':
			valorInicial = 2
		default:
			return false
		}

		// Construir el número completo para el cálculo
		numeroStr := strconv.Itoa(valorInicial) + digitosNIE
		var err error
		numero, err = strconv.Atoi(numeroStr)
		if err != nil {
			return false
		}

		letraProporcionada = string(normalized[8])
	} else {
		// Intentar como DNI
		// Los primeros 8 caracteres deben ser números
		digitosDNI := normalized[0:8]
		if !esNumerico(digitosDNI) {
			return false
		}

		var err error
		numero, err = strconv.Atoi(digitosDNI)
		if err != nil {
			return false
		}

		letraProporcionada = string(normalized[8])
	}

	// Verificar que el último carácter es una letra
	if !esLetra(letraProporcionada) {
		return false
	}

	// Calcular la letra de control correcta
	resto := numero % 23
	letraCorrecta := string(letrasControl[resto])

	// Comparar la letra proporcionada con la calculada
	return letraProporcionada == letraCorrecta
}

// NormalizarNIF normaliza un NIF eliminando espacios y guiones y convirtiéndolo a mayúsculas
func NormalizarNIF(nif string) string {
	normalized := strings.ReplaceAll(nif, " ", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	return strings.ToUpper(normalized)
}

// NormalizeIdentityDocument normaliza cualquier documento de identidad (DNI, NIE, pasaporte, etc.)
// eliminando espacios y guiones y convirtiéndolo a mayúsculas
func NormalizeIdentityDocument(document string) string {
	normalized := strings.ReplaceAll(document, " ", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	return strings.ToUpper(normalized)
}

// esNumerico verifica si una cadena contiene solo dígitos
func esNumerico(s string) bool {
	match, _ := regexp.MatchString("^[0-9]+$", s)
	return match
}

// esLetra verifica si una cadena es una sola letra
func esLetra(s string) bool {
	match, _ := regexp.MatchString("^[A-Z]$", s)
	return match
}

// ValidarPasaporteSenegal valida un pasaporte senegalés usando los 10 caracteres de la MRZ
// (posiciones 1-10 de la línea 2) con el algoritmo de validación 7-3-1
func ValidarPasaporteSenegal(mrz string) bool {
	// Si el campo está vacío, considerarlo válido (es opcional)
	if mrz == "" {
		return true
	}

	// Normalizar: eliminar espacios y convertir a mayúsculas
	normalized := strings.ReplaceAll(mrz, " ", "")
	normalized = strings.ToUpper(normalized)

	// Verificar longitud: debe ser exactamente 10 caracteres
	if len(normalized) != 10 {
		return false
	}

	// Verificar formato: 9 caracteres de datos + 1 dígito de control
	// Los 9 caracteres de datos pueden ser A-Z, 0-9 o <
	numberPart := normalized[0:9]
	checkDigitStr := normalized[9:10]

	// Validar formato de la parte de datos usando regex
	matched, _ := regexp.MatchString("^[A-Z0-9<]{9}$", numberPart)
	if !matched {
		return false
	}

	// Validar que el último carácter es un dígito
	if !esNumerico(checkDigitStr) {
		return false
	}

	// Convertir el dígito de control a entero
	checkDigit, err := strconv.Atoi(checkDigitStr)
	if err != nil {
		return false
	}

	// Aplicar algoritmo 7-3-1
	calculatedCheckDigit := calcularDigitoControl731(numberPart)

	// Comparar el dígito calculado con el proporcionado
	return calculatedCheckDigit == checkDigit
}

// calcularDigitoControl731 implementa el algoritmo de validación 7-3-1 para MRZ
func calcularDigitoControl731(data string) int {
	// Pesos que se repiten: 7, 3, 1, 7, 3, 1, ...
	weights := []int{7, 3, 1}
	sum := 0

	for i, char := range data {
		// Convertir carácter a valor numérico
		value := charToMRZValue(char)

		// Obtener el peso correspondiente (rotación 7-3-1)
		weight := weights[i%3]

		// Multiplicar y sumar
		sum += value * weight
	}

	// El dígito de control es el residuo módulo 10
	return sum % 10
}

// charToMRZValue convierte un carácter MRZ a su valor numérico
// Dígitos 0-9 → 0-9
// Letras A-Z → 10-35
// '<' → 0
func charToMRZValue(char rune) int {
	if char >= '0' && char <= '9' {
		return int(char - '0')
	}
	if char >= 'A' && char <= 'Z' {
		return int(char-'A') + 10
	}
	if char == '<' {
		return 0
	}
	// En caso de carácter no válido, retornar 0
	return 0
}

// ExtraerNumeroPasaporteSenegal extrae el número de pasaporte (9 caracteres)
// de los 10 caracteres de la MRZ, eliminando el dígito de control
func ExtraerNumeroPasaporteSenegal(mrz string) string {
	// Normalizar
	normalized := strings.ReplaceAll(mrz, " ", "")
	normalized = strings.ToUpper(normalized)

	// Verificar longitud
	if len(normalized) != 10 {
		return normalized // Retornar tal cual si no tiene el formato esperado
	}

	// Retornar solo los 9 primeros caracteres (sin el dígito de control)
	return normalized[0:9]
}
