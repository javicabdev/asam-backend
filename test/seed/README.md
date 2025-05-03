# Sistema de Generación de Datos de Prueba (Seeding)

Este módulo implementa un sistema completo para generar datos de prueba que permiten poblar la base de datos con información realista para desarrollo y testing.

## Características

- Generación de datos realistas y consistentes
- Varios conjuntos de datos predefinidos
- Escenarios específicos para casos de prueba
- Control de dependencias entre entidades
- Limpieza sencilla de la base de datos
- Opciones configurables para la generación
- Generación basada en semilla aleatoria (reproducible)

## Estructura

El sistema está organizado de la siguiente manera:

```
test/
└── seed/
    ├── seeder.go         # Coordinador principal de seeding
    ├── generators/
    │   ├── utils.go            # Utilidades comunes
    │   ├── member_generator.go # Generador de miembros
    │   ├── family_generator.go # Generador de familias
    │   ├── family_member_generator.go # Generador de familiares
    │   ├── payment_generator.go # Generador de pagos
    │   └── cashflow_generator.go # Generador de movimientos de caja
    └── data/
        ├── dataset.go    # Fábrica de datasets
        ├── minimal.go    # Dataset mínimo
        ├── full.go       # Dataset completo
        └── scenarios.go  # Escenarios específicos
```

## Uso

### Desde línea de comandos

El sistema incluye un comando para ejecutar el seeding desde la línea de comandos:

```bash
# Seed con dataset mínimo (por defecto)
go run cmd/seed/main.go

# Seed con dataset completo
go run cmd/seed/main.go -type=full

# Seed con escenario específico
go run cmd/seed/main.go -type=scenario -scenario=payment_overdue

# Solo limpiar la base de datos
go run cmd/seed/main.go -clean

# Seed personalizado con cantidades específicas
go run cmd/seed/main.go -type=custom -members=50 -families=20 -payments=100

# Usar una semilla específica para reproducibilidad
go run cmd/seed/main.go -seed=12345
```

### Desde código

También puedes usar el sistema programáticamente:

```go
package main

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	
	"github.com/babacar/asam/asam-backend/test/seed"
	"github.com/babacar/asam/asam-backend/test/seed/data"
)

func main() {
	// Conectar a la base de datos
	db, err := sqlx.Connect("postgres", "conexión_a_db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	// Crear seeder
	seeder := seed.NewSeeder(db)
	
	// Crear contexto
	ctx := context.Background()
	
	// Opción 1: Usar un dataset predefinido
	err = seeder.SeedMinimalDataset(ctx)
	if err != nil {
		log.Fatal(err)
	}
	
	// Opción 2: Seeding personalizado
	err = seeder.Clean(ctx) // Limpiar BD primero
	if err != nil {
		log.Fatal(err)
	}
	
	err = seeder.SeedMiembros(ctx)
	if err != nil {
		log.Fatal(err)
	}
	
	err = seeder.SeedFamilias(ctx)
	if err != nil {
		log.Fatal(err)
	}
	
	// ... etc.
}
```

## Datasets Predefinidos

### Dataset Mínimo

- 10 miembros (5 individuales, 5 familiares)
- 3 familias
- 6 familiares (2 por familia)
- 15 pagos de cuotas
- 20 movimientos de caja

Ideal para pruebas básicas y desarrollo local rápido.

### Dataset Completo

- 70 miembros (50 individuales, 20 familiares)
- 20 familias
- 60 familiares
- 150 pagos de cuotas
- 200 movimientos de caja

Ideal para pruebas más completas y testing de rendimiento.

### Escenarios Específicos

- **payment_overdue**: Miembros con pagos pendientes
- **membership_expired**: Miembros con membresías expiradas
- **large_family**: Familias con muchos miembros
- **financial_emergency**: Casos de emergencias financieras

## Personalización

Puedes ajustar varios parámetros de generación:

- Semilla aleatoria para reproducibilidad
- Número de entidades a generar
- Niveles de concurrencia
- Verbosidad de logs

## Tips y Mejores Prácticas

1. **Limpieza previa**: Siempre ejecuta una limpieza antes de seed para evitar conflictos.
2. **Semilla fija**: Para pruebas reproducibles, usa una semilla fija.
3. **Entornos aislados**: Usa bases de datos separadas para desarrollo y testing.
4. **Automatización**: Integra el seeding en tus flujos de CI/CD.
