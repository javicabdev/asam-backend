# Etapa 1: Construcción
FROM golang:1.24-alpine AS builder

# Instalar dependencias básicas y herramientas de seguridad
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
RUN go mod download && go mod verify

# Copiar el resto del código fuente
COPY . .

# Generar el código GraphQL solo si es necesario
# Verificar si los archivos generados ya existen
RUN if [ ! -f "internal/adapters/gql/generated/generated.go" ]; then \
        echo "=== Archivos generados no encontrados, ejecutando gqlgen ===" && \
        go install github.com/99designs/gqlgen@v0.17.73 && \
        mkdir -p internal/adapters/gql/generated internal/adapters/gql/model && \
        gqlgen generate; \
    else \
        echo "=== Archivos GraphQL ya generados, omitiendo generación ==="; \
    fi

# Build arguments
ARG VERSION=unknown
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

# Compilar la aplicación con información de versión
RUN go build -ldflags "-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildTime=${BUILD_TIME}" -o asam-backend ./cmd/api

# Etapa 2: Imagen final mínima
FROM alpine:3.19

# Argumentos para metadatos
ARG VERSION=unknown
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

# Metadatos de la imagen
LABEL maintainer="ASAM Backend Team" \
      description="ASAM Backend Service" \
      version="${VERSION}" \
      commit="${COMMIT}" \
      build_time="${BUILD_TIME}"

# Instalar certificados, timezone data y dumb-init para mejor manejo de señales
RUN apk --no-cache add ca-certificates tzdata dumb-init && \
    adduser -D -g '' -s /bin/false -h /nonexistent appuser && \
    mkdir -p /app/logs && \
    chown -R appuser:appuser /app

# Directorio de trabajo
WORKDIR /app

# Copiar el ejecutable y las migraciones
COPY --from=builder --chown=appuser:appuser /build/asam-backend .
COPY --from=builder --chown=appuser:appuser /build/migrations ./migrations

# Cambiar al usuario no privilegiado
USER appuser

# Puerto por defecto
EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health/live || exit 1

# Ejecutar la aplicación con dumb-init para mejor manejo de señales
ENTRYPOINT ["dumb-init", "--"]
CMD ["./asam-backend"]
