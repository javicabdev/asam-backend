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

# Compilar la aplicación con optimizaciones
RUN go build -ldflags="-s -w -extldflags '-static'" -o asam-backend ./cmd/api

# Etapa 2: Imagen final mínima
FROM alpine:latest

# Instalar certificados, zona horaria y wget para healthcheck
RUN apk --no-cache add ca-certificates tzdata wget

# Crear un usuario no privilegiado
RUN adduser -D -g '' appuser

# Establecer la zona horaria
ENV TZ=Europe/Madrid

# Crear directorio para la aplicación y logs
WORKDIR /app
RUN mkdir -p /app/logs && chown -R appuser:appuser /app

# Copiar el ejecutable compilado desde la etapa de construcción
COPY --from=builder /build/asam-backend .

# Copiar archivos de migración si son necesarios para ejecutarse en tiempo de ejecución
COPY --from=builder /build/migrations ./migrations

# Copiar el script de entrada
COPY docker-entrypoint.sh /app/
RUN chmod +x /app/docker-entrypoint.sh

# Asignar propiedad de los archivos al usuario no privilegiado
RUN chown -R appuser:appuser /app

# Cambiar al usuario no privilegiado por seguridad
USER appuser

# Exponer el puerto (por defecto para Cloud Run)
EXPOSE 8080

# Variables de entorno para optimizar Go en producción
ENV GOMEMLIMIT=256MiB \
    GOMAXPROCS=2

# Healthcheck para verificar que el servidor está funcionando
HEALTHCHECK --interval=30s --timeout=3s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Comando para ejecutar la aplicación
ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["./asam-backend"]