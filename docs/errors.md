# Catálogo de Errores

Este documento describe los **códigos de error** definidos en el sistema ASAM y explica su uso a lo largo de la aplicación. Además, proporciona lineamientos para el manejo de errores y la forma en la que se exponen a través de la API GraphQL.

---

## 1. Convenciones Generales

- El sistema de errores se basa en un tipo `AppError` definido en `pkg/errors/errors.go`.
- Cada `AppError` contiene:
    - **Code**: Un string (definido como `ErrorCode`) que identifica la clase de error (p. ej. `VALIDATION_FAILED`).
    - **Message**: Texto que describe el error al usuario.
    - **Fields** (opcional): Un mapa clave->valor para detallar validaciones de campos o información adicional.
    - **Cause** (opcional): Un `error` que explica la causa raíz interna (para logging y depuración).

### Respuesta en GraphQL

Cuando ocurre un error en un resolver, si el sistema detecta un `*AppError`, se convierte en la respuesta con la forma:

```json
{
  "errors": [
    {
      "message": "Descripción del error",
      "path": ["nombreDelResolver"],
      "extensions": {
        "code": "CÓDIGO_DE_ERROR",
        "fields": {
          // Datos extra si corresponde
        }
      }
    }
  ],
  "data": null
}
```

- `extensions.code` coincide con `AppError.Code`.
- `extensions.fields` se rellena, en caso de validaciones o datos específicos de error.

## 2. Códigos de Error

A continuación se listan los códigos de error principales definidos en pkg/errors/errors.go, junto con su significado, mensaje sugerido y acciones recomendadas.

### 2.1 Errores de Validación

| Código              | Descripción                                                                    | Mensaje de Usuario                                    | Acciones Recomendadas                                   |
|---------------------|--------------------------------------------------------------------------------|-------------------------------------------------------|---------------------------------------------------------|
| `VALIDATION_FAILED` | Reglas de validación incumplidas (formato, rangos, campos obligatorios, etc.). | "Datos de entrada no válidos" o descripción detallada | Verificar los campos retornados en `fields` y corregir. |


Ejemplo: Campo amount <= 0 al registrar un pago.

```json
{
  "errors": [
    {
      "message": "El campo amount no puede ser <= 0",
      "extensions": {
        "code": "VALIDATION_FAILED",
        "fields": {
          "amount": "must be greater than 0"
        }
      }
    }
  ]
}
```

### 2.2 Errores de Negocio

| Código               | Descripción                                                                         | Mensaje de Usuario                           | Acciones Recomendadas                                                                 |
|----------------------|-------------------------------------------------------------------------------------|----------------------------------------------|---------------------------------------------------------------------------------------|
| `NOT_FOUND`          | Recurso no encontrado (member, family, payment, etc.).                              | "Recurso no encontrado" o "member not found" | Verificar el ID o referencia proporcionada, asegurarse de que exista.                 |
| `INVALID_OPERATION`  | Operación no permitida por las reglas de negocio.                                   | "Operación no válida"                        | Revisar la secuencia o el estado previo antes de invocar la operación.                |
| `INSUFFICIENT_FUNDS` | No hay fondos suficientes para completar una operación financiera.                  | "Fondos insuficientes"                       | Aportar más fondos o comprobar que el balance sea correcto.                           |
| `DUPLICATE_ENTRY`    | Intento de crear algo que ya existe (ej. dos miembros con el mismo `numero_socio`). | "El recurso ya existe"                       | Usar un ID único o verificar si ya estaba creado.                                     |
| `INVALID_STATUS`     | Estado no válido al actualizar un recurso (p. ej., pasar de "cancelado" a "paid").  | "Estado no válido"                           | Revisar la lógica de estados o la transición permitida.                               |
| `INVALID_DATE`       | Fecha no válida o inconsistente (ej. fecha de baja anterior a la de alta).          | "Fecha no válida"                            | Ajustar las fechas según la regla de negocio.                                         |
| `INVALID_AMOUNT`     | Monto no válido (ej. negativo cuando se espera positivo).                           | "Monto no válido"                            | Revisar el monto ingresado y corregir.                                                |

### 2.3 Errores de Sistema

| Código             | Descripción                                                               | Mensaje de Usuario               | Acciones Recomendadas                                                 |
|--------------------|---------------------------------------------------------------------------|----------------------------------|-----------------------------------------------------------------------|
| `DATABASE_ERROR`   | Error interno de base de datos (fallo de conexión, SQL, etc.)             | "Error en la base de datos"      | Contactar soporte, revisar logs de la DB, verificar la configuración. |
| `INTERNAL_ERROR`   | Error genérico de servidor para casos no contemplados en otras categorías | "Error interno del servidor"     | Contactar soporte, revisar logs y trazar la causa raíz.               |
| `NETWORK_ERROR`    | Fallos de comunicación externa (servicios remotos, red).                  | "Fallo de conexión o red"        | Revisar configuración de red, estado de servicios externos, etc.      |

### 2.4 Errores de Autenticación

| Código          | Descripción                                                | Mensaje de Usuario    | Acciones Recomendadas                                          |
|-----------------|------------------------------------------------------------|-----------------------|----------------------------------------------------------------|
| `UNAUTHORIZED`  | Faltan credenciales o token no provisto.                   | "No autenticado"      | Iniciar sesión, reenviar token o credenciales correctos.       |
| `FORBIDDEN`     | El usuario no tiene permisos para la operación solicitada. | "Operación prohibida" | Usar una cuenta con permisos suficientes o elevar privilegios. |
| `INVALID_TOKEN` | Token inválido o caducado (JWT, OAuth, etc.).              | "Token inválido"      | Renovar token, reautenticarse.                                 |

## 3. Ejemplo de Error en GraphQL

Cuando ocurre un error (por ejemplo, `VALIDATION_FAILED` en un campo), el cliente recibe algo como:

```json
{
  "errors": [
    {
      "message": "El número de socio es requerido",
      "path": ["createMember"],
      "extensions": {
        "code": "VALIDATION_FAILED",
        "fields": {
          "numero_socio": "campo requerido"
        }
      }
    }
  ],
  "data": null
}
```

* `extensions.code` se mapea desde `AppError.Code`.
* `extensions.fields` describe detalles de validación.

## 4. Mapeo de Errores en gqlgen (ErrorPresenter)

En `internal/adapters/gql/handlerChain.go` (o similar), se configura un `ErrorPresenter` para que cada `AppError` se muestre con `extensions.code`, `extensions.fields`, etc. Esto asegura la consistencia en todas las operaciones GraphQL.

## 5. Guía de Uso y Recomendaciones

1. **Crear** un `*AppError` cuando surja un error de validación o negocio en la capa de dominio.
2. **No** envolver varias veces el mismo error. Si ya tenemos un `VALIDATION_FAILED`, propagarlo sin sobrescribirlo.
3. **Exponer** datos relevantes en `fields` cuando sea un error de validación (ej. campo, requerimiento, etc.).
4. **Evitar** poner información sensible en `Message` o `fields`.
5. **Documentar** nuevos códigos de error en este archivo cada vez que se agreguen en `pkg/errors/errors.go`.

## 6. Futuras Extensiones

* **Internacionalización (i18n)**: Mapear los `Code` a mensajes en distintos idiomas.
* **Logging**: `appErr.Cause` u otros campos pueden usarse para logs de servidor o para rastrear errores más en detalle.
* **HTTP**: En caso de REST, se pueden mapear a códigos HTTP específicos (`VALIDATION_FAILED` → `400`, `NOT_FOUND` → `404`, etc.), aunque GraphQL no funciona por status codes, sino por contenido.

## 7. Ejemplos de Implementación

1. **Validación de Fechas**
    * `ErrInvalidDate`: Mensaje `"La fecha de alta no puede ser futura"`, `fields = {"fecha_alta": "no puede ser futura"}`.
2. **Recurso No Encontrado**
   * `NOT_FOUND`: Si `GetMemberByID` retorna `nil`, crear `NewNotFoundError("member")`.
   * En GraphQL: `"code": "NOT_FOUND", "message": "member not found"`.
3. **Operaciones de Pago**
   * `INVALID_AMOUNT` si `payment.Amount <= 0`.
   * `INSUFFICIENT_FUNDS` si la cuenta no cubre el saldo requerido.