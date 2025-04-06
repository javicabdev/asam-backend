# Script de Pruebas para ASAM Backend

Este script facilita la ejecución de pruebas en entornos Docker controlados para el proyecto ASAM backend.

## Descripción

`test.ps1` automatiza la configuración, ejecución y limpieza del entorno de pruebas utilizando Docker Compose. El script permite:

- Construir imágenes Docker con o sin caché
- Ejecutar pruebas específicas por módulo o todas las pruebas
- Conservar el entorno después de las pruebas si es necesario
- Utilizar Docker Compose con Bake para mejorar el rendimiento de construcción

## Parámetros

| Parámetro   | Tipo    | Descripción                                                |
|-------------|---------|-----------------------------------------------------------|
| `-Build`    | switch  | Reconstruye las imágenes Docker sin usar caché             |
| `-NoCleanup`| switch  | Evita eliminar los contenedores al finalizar               |
| `-Module`   | string  | Especifica el módulo a probar (vacío = todos los módulos)  |
| `-UseBake`  | switch  | Habilita Docker Compose con Bake para mejor rendimiento    |

## Ejemplos de Uso

### Ejecutar todas las pruebas con imágenes existentes
```powershell
.\test.ps1
```

### Reconstruir imágenes y ejecutar todas las pruebas
```powershell
.\test.ps1 -Build
```

### Ejecutar pruebas solo para un módulo específico
```powershell
.\test.ps1 -Module "auth"
```

### Reconstruir imágenes y ejecutar pruebas para un módulo
```powershell
.\test.ps1 -Build -Module "users"
```

### Mantener el entorno de pruebas después de la ejecución
```powershell
.\test.ps1 -NoCleanup
```

### Utilizar Bake para construcción más rápida
```powershell
.\test.ps1 -Build -UseBake
```

### Combinación de opciones
```powershell
.\test.ps1 -Build -UseBake -Module "api" -NoCleanup
```

## Notas sobre Bake

La opción `-UseBake` utiliza Docker Compose con Buildx Bake como motor de construcción subyacente, lo que proporciona:

- Construcción de imágenes en paralelo más eficiente
- Mejor utilización del caché de construcción
- Compatibilidad completa con Buildkit
- Soporte para construcción multi-plataforma

Se recomienda usar esta opción en proyectos con múltiples servicios o cuando se necesite optimizar el tiempo de construcción.

## Flujo de Ejecución

1. Configuración del entorno de pruebas
2. Limpieza de entornos previos (si existen)
3. Construcción de imágenes (si `-Build` está activado)
4. Inicio del contenedor de PostgreSQL
5. Verificación de la conexión a la base de datos
6. Ejecución de las pruebas
7. Limpieza del entorno (a menos que `-NoCleanup` esté activado)

## Requisitos

- Docker Desktop
- Docker Compose
- PowerShell 5.0 o superior