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
