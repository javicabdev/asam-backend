# Estructura de Pruebas

## Estructura de Directorios

Para que la cobertura de código funcione correctamente, los tests deben reflejar la misma estructura de directorios que el código que están probando. Por eso, hemos reorganizado los tests siguiendo esta convención:

```
test/
  internal/               <- Refleja el paquete internal del código fuente
    domain/               <- Refleja el paquete domain del código fuente
      services/           <- Tests para servicios
      models/             <- Tests para modelos
    adapters/             <- Refleja el paquete adapters del código fuente
      gql/                <- Tests para GraphQL
      middleware/         <- Tests para middleware
```

## Ejecutando Tests con Cobertura

Para ejecutar los tests con cobertura, utiliza el siguiente comando:

```bash
go test ./test/... -coverprofile=coverage.out -coverpkg=./...
```

Esto ejecutará todas las pruebas y generará un archivo de cobertura que incluye todos los paquetes del proyecto.

Para generar un reporte HTML de cobertura:

```bash
go tool cover -html=coverage.out -o coverage.html
```

También puedes usar el script `run_tests_with_coverage.sh` incluido en la raíz del proyecto.

## Mantenimiento de los Tests

Al agregar nuevos tests:

1. Asegúrate de colocarlos en el directorio correspondiente dentro de `test/` que refleje la estructura del código que estás probando
2. Mantén el mismo nombre de paquete que el código original (o añade el sufijo `_test`)
3. Asegúrate de que los mocks están configurados correctamente para ejecutar el código real

Si los tests siguen reportando cobertura de 0%, verifica:

1. Que el paquete corresponda correctamente con el código bajo prueba
2. Que los mocks estén configurados para llamar a las implementaciones reales
3. Que estés usando el flag `-coverpkg=./...` para incluir todos los paquetes
