package middleware

import (
	"context"
	"net/http"

	"gorm.io/gorm"
)

type transactionKey struct{}

type TransactionMiddleware struct {
	db   *gorm.DB
	next http.Handler
}

func NewTransactionMiddleware(db *gorm.DB) *TransactionMiddleware {
	return &TransactionMiddleware{
		db: db,
	}
}

func (m *TransactionMiddleware) Handler(next http.Handler) http.Handler {
	m.next = next
	return m
}

func (m *TransactionMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Solo iniciar transacción para mutations
	if r.Method != http.MethodPost {
		m.next.ServeHTTP(w, r)
		return
	}

	// Iniciar transacción
	tx := m.db.Begin()
	if tx.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Añadir transacción al contexto
	ctx := context.WithValue(r.Context(), transactionKey{}, tx)

	// Crear ResponseWriter personalizado para capturar el código de estado
	wrapped := NewResponseWriter(w)

	// Ejecutar el siguiente handler con el contexto modificado
	m.next.ServeHTTP(wrapped, r.WithContext(ctx))

	// Si hay error, hacer rollback
	if wrapped.statusCode >= 400 {
		tx.Rollback()
		return
	}

	// Si no hay errores, commit
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// ResponseWriter personalizado para capturar el código de estado
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, http.StatusOK}
}

func (w *ResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func GetTransaction(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(transactionKey{}).(*gorm.DB); ok {
		return tx
	}
	return nil
}
