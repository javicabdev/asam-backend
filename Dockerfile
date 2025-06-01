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

# Generar el código GraphQL antes de compilar
RUN go run ./cmd/generate/main.go

# Compilar la aplicación con información de versión
ARG VERSION=unknown
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
RUN go build -ldflags="-s -w \
    -X main.Version=${VERSION} \
    -X main.Commit=${COMMIT} \
    -X main.BuildTime=${BUILD_TIME}" \
    -o asam-backend ./cmd/api

# Etapa 2: Verificación de seguridad (opcional pero recomendado)
FROM aquasec/trivy:latest AS security
COPY --from=builder /build/go.mod /build/go.sum /
RUN trivy fs --no-progress --security-checks vuln --exit-code 0 /

# Etapa 3: Imagen final mínima
FROM alpine:3.19

# Metadatos de la imagen
LABEL maintainer="ASAM Backend Team" \
      description="ASAM Backend Service" \
      version="${VERSION}"

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