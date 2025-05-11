package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/middleware"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/constants"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test"
)

// setupMockAuth configura un servicio de autenticación mock para pruebas
func setupMockAuth() (*test.MockAuthService, *test.MockLogger) {
	authService := new(test.MockAuthService)
	logger := new(test.MockLogger)
	return authService, logger
}

// TestAuthMiddleware_NoToken prueba que las solicitudes sin token que requieren autenticación devuelven un error 401
func TestAuthMiddleware_NoToken(t *testing.T) {
	// Configurar mocks
	authService, logger := setupMockAuth()
	authMiddleware := middleware.AuthMiddleware(authService, logger)

	// Crear un handler de siguiente nivel que nunca debería ser llamado
	nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// Esta función no debería ser llamada
		t.Error("Se llamó al handler siguiente cuando debió haber fallado con 401")
	})

	// Crear el request para una operación protegida (no es login, refresh, etc.)
	query := `{"operationName":"GetMember","query":"query GetMember($id: ID!) { getMember(id: $id) { nombre apellidos } }"}`
	req := httptest.NewRequest("POST", "/graphql", strings.NewReader(query))
	req.Header.Set("Content-Type", "application/json")

	// Configurar GetBody para que la solicitud pueda ser leída múltiples veces
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(query)), nil
	}

	// Registrar la respuesta
	w := httptest.NewRecorder()

	// Ejecutar el middleware
	handlerToTest := authMiddleware(nextHandler)
	handlerToTest.ServeHTTP(w, req)

	// Verificar que se devolvió un código 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Verificar que el cuerpo de la respuesta contiene un error de GraphQL
	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "La respuesta debería ser JSON válido")

	// Imprimir la respuesta completa para depurar
	fmt.Printf("\nRespuesta completa en NoToken: %+v\n", response)
	fmt.Printf("\nCuerpo de respuesta raw en NoToken: %s\n", w.Body.String())

	// Verificar que hay errores en la respuesta
	errs, ok := response["errors"].([]any)
	assert.True(t, ok, "La respuesta debería contener errores")
	assert.NotEmpty(t, errs, "La lista de errores no debería estar vacía")

	// Verificar que el error tiene el código correcto
	firstError := errs[0].(map[string]any)
	assert.Equal(t, "UNAUTHORIZED: Unauthorized access to the resource", firstError["message"])
	assert.Equal(t, "UNAUTHORIZED", firstError["extensions"].(map[string]any)["code"])
}

// TestAuthMiddleware_PublicOperations prueba que las operaciones públicas pasan sin token
func TestAuthMiddleware_PublicOperations(t *testing.T) {
	// Configurar mocks
	authService, logger := setupMockAuth()
	authMiddleware := middleware.AuthMiddleware(authService, logger)

	// Crear un handler de siguiente nivel que verifica ser llamado
	nextCalled := false
	nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		nextCalled = true
	})

	// Array de operaciones públicas a probar
	publicOps := []struct {
		name        string
		query       string
		operation   string
		contentType string
	}{
		{
			name:        "Login Mutation",
			query:       `{"operationName":"login","query":"mutation login($input: LoginInput!) { login(input: $input) { accessToken } }"}`,
			operation:   "login",
			contentType: "application/json",
		},
		{
			name:        "Refresh Token Mutation",
			query:       `{"operationName":"refreshToken","query":"mutation refreshToken($input: RefreshTokenInput!) { refreshToken(input: $input) { accessToken } }"}`,
			operation:   "refreshToken",
			contentType: "application/json",
		},
		{
			name:        "Introspection Query",
			query:       `{"operationName":"IntrospectionQuery","query":"query IntrospectionQuery { __schema { types { name } } }"}`,
			operation:   "IntrospectionQuery",
			contentType: "application/json",
		},
	}

	for _, tc := range publicOps {
		t.Run(tc.name, func(t *testing.T) {
			// Resetear la bandera
			nextCalled = false

			// Crear la solicitud
			req := httptest.NewRequest("POST", "/graphql", strings.NewReader(tc.query))
			req.Header.Set("Content-Type", tc.contentType)

			// Configurar GetBody para que la solicitud pueda ser leída múltiples veces
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader(tc.query)), nil
			}

			// Registrar la respuesta
			w := httptest.NewRecorder()

			// Ejecutar el middleware
			handlerToTest := authMiddleware(nextHandler)
			handlerToTest.ServeHTTP(w, req)

			// Verificar que se llamó al siguiente handler
			assert.True(t, nextCalled, "El siguiente handler debería haber sido llamado para %s", tc.name)

			// Verificar que no se devolvió un error
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// TestAuthMiddleware_InvalidTokenFormat prueba que los tokens con formato inválido devuelven un error 401
func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	// Configurar mocks
	authService, logger := setupMockAuth()
	authMiddleware := middleware.AuthMiddleware(authService, logger)

	// Crear un handler de siguiente nivel que nunca debería ser llamado
	nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// Esta función no debería ser llamada
		t.Error("Se llamó al handler siguiente cuando debió haber fallado con 401")
	})

	// No es necesario configurar el mock para token vacío, ya que el middleware debería rechazarlo por formato inválido

	// Array de formatos de token inválidos a probar
	invalidTokens := []struct {
		name       string
		authHeader string
		errorMsg   string
	}{
		{
			name:       "Token sin Bearer",
			authHeader: "token123",
			errorMsg:   "UNAUTHORIZED: Unauthorized access to the resource",
		},
		{
			name:       "Token con formato incorrecto",
			authHeader: "Bearer token token",
			errorMsg:   "UNAUTHORIZED: Unauthorized access to the resource",
		},
		{
			name:       "Token vacío",
			authHeader: "Bearer ",
			errorMsg:   "UNAUTHORIZED: Unauthorized access to the resource",
		},
	}

	for _, tc := range invalidTokens {
		t.Run(tc.name, func(t *testing.T) {
			// Crear la solicitud con una operación protegida
			query := `{"operationName":"GetMember","query":"query GetMember($id: ID!) { getMember(id: $id) { nombre apellidos } }"}`
			req := httptest.NewRequest("POST", "/graphql", strings.NewReader(query))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", tc.authHeader)

			// Configurar GetBody para que la solicitud pueda ser leída múltiples veces
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader(query)), nil
			}

			// Registrar la respuesta
			w := httptest.NewRecorder()

			// Ejecutar el middleware
			handlerToTest := authMiddleware(nextHandler)
			handlerToTest.ServeHTTP(w, req)

			// Verificar que se devolvió un código 401
			assert.Equal(t, http.StatusUnauthorized, w.Code)

			// Verificar que el cuerpo de la respuesta contiene un error de GraphQL
			var response map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err, "La respuesta debería ser JSON válido")

			// Imprimir la respuesta completa para depurar
			fmt.Printf("\nRespuesta completa en InvalidTokenFormat (%s): %+v\n", tc.name, response)
			fmt.Printf("\nCuerpo de respuesta raw: %s\n", w.Body.String())

			// Verificar que hay errores en la respuesta
			errs, ok := response["errors"].([]any)
			assert.True(t, ok, "La respuesta debería contener errores")
			assert.NotEmpty(t, errs, "La lista de errores no debería estar vacía")

			// Verificar que el error tiene el código correcto
			firstError := errs[0].(map[string]any)
			assert.Equal(t, tc.errorMsg, firstError["message"])
			assert.Equal(t, "UNAUTHORIZED", firstError["extensions"].(map[string]any)["code"])
		})
	}
}

// TestAuthMiddleware_InvalidToken prueba que los tokens inválidos devuelven un error 401
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	// Configurar mocks
	authService, logger := setupMockAuth()
	authMiddleware := middleware.AuthMiddleware(authService, logger)

	// Crear un handler de siguiente nivel que nunca debería ser llamado
	nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// Esta función no debería ser llamada
		t.Error("Se llamó al handler siguiente cuando debió haber fallado con 401")
	})

	// Configurar que el servicio de autenticación devuelva un error
	token := "validTokenFormat"
	authService.On("ValidateToken", mock.Anything, token).Return(nil, appErrors.NewBusinessError(appErrors.ErrInvalidToken, "Token expirado"))

	// Crear la solicitud con una operación protegida
	query := `{"operationName":"GetMember","query":"query GetMember($id: ID!) { getMember(id: $id) { nombre apellidos } }"}`
	req := httptest.NewRequest("POST", "/graphql", strings.NewReader(query))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Configurar GetBody para que la solicitud pueda ser leída múltiples veces
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(query)), nil
	}

	// Registrar la respuesta
	w := httptest.NewRecorder()

	// Ejecutar el middleware
	handlerToTest := authMiddleware(nextHandler)
	handlerToTest.ServeHTTP(w, req)

	// Verificar que se devolvió un código 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Verificar que el cuerpo de la respuesta contiene un error de GraphQL
	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "La respuesta debería ser JSON válido")

	// Imprimir la respuesta completa para depurar
	fmt.Printf("\nRespuesta completa en InvalidToken: %+v\n", response)
	fmt.Printf("\nCuerpo de respuesta raw: %s\n", w.Body.String())

	// Verificar que hay errores en la respuesta
	errs, ok := response["errors"].([]any)
	assert.True(t, ok, "La respuesta debería contener errores")
	assert.NotEmpty(t, errs, "La lista de errores no debería estar vacía")

	// Verificar que el error tiene el código correcto
	firstError := errs[0].(map[string]any)
	assert.Equal(t, "Token expirado", firstError["message"])
	assert.Equal(t, "UNAUTHORIZED", firstError["extensions"].(map[string]any)["code"])
}

// TestAuthMiddleware_ValidToken prueba que un token válido permite acceder a operaciones protegidas
func TestAuthMiddleware_ValidToken(t *testing.T) {
	// Configurar mocks
	authService, logger := setupMockAuth()
	authMiddleware := middleware.AuthMiddleware(authService, logger)

	// Crear un usuario mock válido
	mockUser := &models.User{
		Model:    gorm.Model{ID: 1},
		Username: "testuser",
		Role:     models.RoleAdmin,
		IsActive: true,
	}

	// Crear un handler de siguiente nivel que verifica la información del contexto
	var contextUser *models.User
	var isAuthorized bool
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verificar que la información del usuario se ha añadido al contexto
		ctx := r.Context()
		contextUser = ctx.Value(constants.UserContextKey).(*models.User)
		isAuthorized = ctx.Value(constants.AuthorizedContextKey).(bool)

		// Responder OK
		w.WriteHeader(http.StatusOK)
	})

	// Configurar que el servicio de autenticación devuelva un usuario válido
	token := "validToken"
	authService.On("ValidateToken", mock.Anything, token).Return(mockUser, nil)

	// Crear la solicitud con una operación protegida
	query := `{"operationName":"GetMember","query":"query GetMember($id: ID!) { getMember(id: $id) { nombre apellidos } }"}`
	req := httptest.NewRequest("POST", "/graphql", strings.NewReader(query))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Configurar GetBody para que la solicitud pueda ser leída múltiples veces
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(query)), nil
	}

	// Registrar la respuesta
	w := httptest.NewRecorder()

	// Ejecutar el middleware
	handlerToTest := authMiddleware(nextHandler)
	handlerToTest.ServeHTTP(w, req)

	// Verificar que se devolvió un código 200
	assert.Equal(t, http.StatusOK, w.Code)

	// Verificar que el siguiente handler recibió la información correcta en el contexto
	assert.NotNil(t, contextUser, "El usuario debería estar en el contexto")
	assert.Equal(t, mockUser.ID, contextUser.ID, "El ID del usuario en el contexto debería ser correcto")
	assert.Equal(t, mockUser.Username, contextUser.Username, "El username del usuario en el contexto debería ser correcto")
	assert.Equal(t, mockUser.Role, contextUser.Role, "El rol del usuario en el contexto debería ser correcto")
	assert.True(t, isAuthorized, "isAuthorized debería ser true en el contexto")
}

// TestAuthMiddleware_ServerError prueba que los errores inesperados del servidor devuelven 401
func TestAuthMiddleware_ServerError(t *testing.T) {
	// Configurar mocks
	authService, logger := setupMockAuth()
	authMiddleware := middleware.AuthMiddleware(authService, logger)

	// Crear un handler de siguiente nivel que nunca debería ser llamado
	nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// Esta función no debería ser llamada
		t.Error("Se llamó al handler siguiente cuando debió haber fallado con 401")
	})

	// Configurar que el servicio de autenticación devuelva un error inesperado
	token := "validTokenFormat"
	authService.On("ValidateToken", mock.Anything, token).Return(nil, errors.New("error inesperado del servidor"))

	// Crear la solicitud con una operación protegida
	query := `{"operationName":"GetMember","query":"query GetMember($id: ID!) { getMember(id: $id) { nombre apellidos } }"}`
	req := httptest.NewRequest("POST", "/graphql", strings.NewReader(query))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Configurar GetBody para que la solicitud pueda ser leída múltiples veces
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(query)), nil
	}

	// Registrar la respuesta
	w := httptest.NewRecorder()

	// Ejecutar el middleware
	handlerToTest := authMiddleware(nextHandler)
	handlerToTest.ServeHTTP(w, req)

	// Verificar que se devolvió un código 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Verificar que el cuerpo de la respuesta contiene un error de GraphQL
	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "La respuesta debería ser JSON válido")

	// Imprimir la respuesta completa para depurar
	fmt.Printf("\nRespuesta completa en ServerError: %+v\n", response)
	fmt.Printf("\nCuerpo de respuesta raw: %s\n", w.Body.String())

	// Verificar que hay errores en la respuesta
	errs, ok := response["errors"].([]any)
	assert.True(t, ok, "La respuesta debería contener errores")
	assert.NotEmpty(t, errs, "La lista de errores no debería estar vacía")

	// Verificar el mensaje de error predeterminado para errores no específicos
	firstError := errs[0].(map[string]any)
	assert.Equal(t, "Token inválido o expirado", firstError["message"])
}
