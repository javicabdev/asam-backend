package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/javicabdev/asam-backend/pkg/constants"
)

// clientInfoMiddleware captura información del cliente (IP, User-Agent, Device Name)
// y la añade al contexto de la petición antes de que otros manejadores la procesen.
func clientInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Extraer la IP real del cliente utilizando la lógica robusta.
		clientIP := getClientIP(r)

		// 2. Extraer el User-Agent.
		userAgent := r.Header.Get("User-Agent")

		// 3. Determinar el nombre del dispositivo.
		// Primero, intentamos obtenerlo de la cabecera "X-Device-Name" (si está presente).
		deviceName := r.Header.Get("X-Device-Name")
		if deviceName == "" {
			// Si no hay cabecera "X-Device-Name", lo inferimos del User-Agent.
			deviceName = getDeviceName(userAgent)
		}

		// 4. Añadir la información al contexto de la petición.
		ctx := r.Context()
		ctx = context.WithValue(ctx, constants.IPContextKey, clientIP)
		ctx = context.WithValue(ctx, constants.UserAgentContextKey, userAgent)
		ctx = context.WithValue(ctx, constants.DeviceNameContextKey, deviceName)

		// 5. Continuar con el manejador siguiente, pasando el contexto enriquecido.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getClientIP extrae la IP real del cliente, considerando cabeceras de proxy.
// Prioriza "X-Real-IP", luego "X-Forwarded-For" (tomando la primera IP si hay varias),
// y finalmente "RemoteAddr" (eliminando el puerto).
func getClientIP(r *http.Request) string {
	// Comprobar las cabeceras de proxy en orden de prioridad.
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// "X-Forwarded-For" puede contener múltiples IPs, tomamos la primera.
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Si no hay cabeceras de proxy, usar RemoteAddr.
	// RemoteAddr puede incluir el puerto, así que lo separamos.
	ip := r.RemoteAddr
	if colonIdx := strings.LastIndex(ip, ":"); colonIdx != -1 {
		ip = ip[:colonIdx]
	}

	return ip
}

// getDeviceName intenta determinar el tipo de dispositivo a partir de la cadena User-Agent.
func getDeviceName(userAgent string) string {
	ua := strings.ToLower(userAgent)

	// Detectar dispositivos móviles.
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") {
		if strings.Contains(ua, "tablet") || strings.Contains(ua, "kindle") {
			return "Android Tablet"
		}
		return "Android Mobile"
	}

	if strings.Contains(ua, "iphone") {
		return "iPhone"
	}

	if strings.Contains(ua, "ipad") {
		return "iPad"
	}

	// Detectar navegadores de escritorio.
	if strings.Contains(ua, "windows") {
		return "Windows Desktop"
	}

	if strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os") {
		return "Mac Desktop"
	}

	if strings.Contains(ua, "linux") {
		return "Linux Desktop"
	}

	// Valor por defecto si no se detecta nada específico.
	return "Web Browser"
}
