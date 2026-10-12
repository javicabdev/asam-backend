# Etapa 1: Construcción
FROM golang:1.26.9-alpine@sha256:3082400e369fa24d5fc60bca20edab3f6d604e0c5a690ec66b295eff4dd87ade AS builder

# Dejar constancia en el log del compilador con el que se construye la imagen
RUN go version

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
        go install github.com/99designs/gqlgen@v0.17.95 && \
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
# Stable Alpine, not edge (the development branch): 3.24 ships postgresql17-client
# (17.11), which pg_dump needs to back up the PostgreSQL 17 server. If a future
# Alpine drops postgresql17, the CI Docker Build fails on the apk add below.
FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

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

# Instalar certificados, timezone data, dumb-init y postgresql17-client para backups
RUN apk --no-cache add ca-certificates tzdata dumb-init postgresql17-client && \
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
