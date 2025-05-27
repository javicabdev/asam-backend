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

# Instalar certificados y zona horaria
RUN apk --no-cache add ca-certificates tzdata

# Crear un usuario no privilegiado
RUN adduser -D -g '' appuser

# Establecer la zona horaria
ENV TZ=Europe/Madrid

# Crear directorio para la aplicación
WORKDIR /app

# Copiar el ejecutable compilado desde la etapa de construcción
COPY --from=builder /build/asam-backend .

# Copiar archivos de migración si son necesarios para ejecutarse en tiempo de ejecución
COPY --from=builder /build/migrations ./migrations

# Asignar propiedad de los archivos al usuario no privilegiado
RUN chown -R appuser:appuser /app

# Cambiar al usuario no privilegiado por seguridad
USER appuser

# Exponer el puerto (por defecto para Cloud Run)
EXPOSE 8080

# Variables de entorno para optimizar Go en producción
ENV GOMEMLIMIT=256MiB \
    GOMAXPROCS=2

# Comando para ejecutar la aplicación
CMD ["./asam-backend"]