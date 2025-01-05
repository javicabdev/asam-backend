package gql

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	asamErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func CustomErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	var appErr *asamErrors.AppError
	if errors.As(err, &appErr) {
		// Mapeamos AppError -> gqlerror.Error
		return &gqlerror.Error{
			Message: appErr.Message,
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]any{
				"code":   appErr.Code,
				"fields": appErr.Fields,
			},
		}
	}
	// Si no es un AppError, fallback
	return graphql.DefaultErrorPresenter(ctx, err)
}
