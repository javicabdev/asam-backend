package gql

import (
	"context"
	"errors"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/middleware"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gorm.io/gorm"
)

// CustomErrorPresenter transforma los errores en errores GraphQL
func CustomErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	// Si ya es un error GraphQL, retornarlo tal cual
	if gqlErr, ok := err.(*gqlerror.Error); ok {
		return gqlErr
	}

	// Obtener la ruta de la query/mutation actual
	path := graphql.GetPath(ctx)

	// Conversión de errores específicos
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Error de GORM "not found"
		return &gqlerror.Error{
			Path:    path,
			Message: "Resource not found",
			Extensions: map[string]interface{}{
				"code": appErrors.ErrNotFound,
			},
		}
	}

	// Mapeo de AppError
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		// Construir extensiones con el código y campos de error
		extensions := map[string]interface{}{
			"code": appErr.Code,
		}

		// Añadir campos de validación si existen
		if appErr.Fields != nil && len(appErr.Fields) > 0 {
			extensions["fields"] = appErr.Fields
		}

		return &gqlerror.Error{
			Path:       path,
			Message:    appErr.Message,
			Extensions: extensions,
		}
	}

	// Para errores no estructurados, crear un error interno genérico
	return &gqlerror.Error{
		Path:    path,
		Message: fmt.Sprintf("Internal error: %s", err.Error()),
		Extensions: map[string]interface{}{
			"code": appErrors.ErrInternalError,
		},
	}
}

// ErrorHandler devuelve una función middleware.ConfigErrorPresenter compatible
// para integrar con nuestro sistema refactorizado
func ErrorHandler() graphql.ErrorPresenterFunc {
	return func(ctx context.Context, err error) *gqlerror.Error {
		// Buscar errorHandler en el contexto
		if handler, ok := ctx.Value(middleware.ErrorHandlerKey{}).(*middleware.ErrorMiddleware); ok {
			return handler.HandleError(ctx, err)
		}

		// Fallback al CustomErrorPresenter si no hay un errorHandler en el contexto
		return CustomErrorPresenter(ctx, err)
	}
}

// RecoverFunc proporciona una función para recuperarse de pánicos en resolvers GraphQL
func RecoverFunc() graphql.RecoverFunc {
	return func(ctx context.Context, err interface{}) error {
		// Construir mensaje de error
		errMsg := "Internal server error"
		if err != nil {
			errMsg = fmt.Sprintf("GraphQL panic: %v", err)
		}

		// Crear un error estructurado
		appErr := appErrors.New(appErrors.ErrInternalError, errMsg)

		// Convertir a gqlerror.Error usando nuestro presenter
		return CustomErrorPresenter(ctx, appErr)
	}
}
