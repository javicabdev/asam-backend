# Etapa 1: Construcción
FROM golang:1.23-alpine AS builder

# Instalar dependencias básicas
RUN apk add --no-cache git ca-certificates tzdata

# Variables de entorno para la compilación
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Directorio de trabajo
WORKDIR /build

# Copiar los archivos de dependencias y descargarlas primero (para aprovechar la caché)
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto del código fuente
COPY . .

# Compilar la aplicación
RUN go build -ldflags="-s -w" -o asam-backend ./cmd/api

# Etapa 2: Imagen final mínima
FROM alpine:latest

# Instalar certificados y crear usuario
RUN apk --no-cache add ca-certificates && \
    adduser -D -g '' appuser

# Directorio de trabajo
WORKDIR /app

# Copiar el ejecutable y las migraciones
COPY --from=builder /build/asam-backend .
COPY --from=builder /build/migrations ./migrations

# Cambiar permisos
RUN chown -R appuser:appuser /app

# Cambiar al usuario no privilegiado
USER appuser

# Puerto por defecto
EXPOSE 8080

# Ejecutar la aplicación directamente
CMD ["./asam-backend"]