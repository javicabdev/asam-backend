# Convención de Numeración de Membresías ASAM

## Resumen

El sistema ASAM utiliza una convención específica para la numeración de membresías que distingue entre miembros individuales y familiares:

- **Prefijo 'A'**: Miembros FAMILIARES (asociados a una entidad Family)
- **Prefijo 'B'**: Miembros INDIVIDUALES

## Formato

El formato completo es: `[A|B]XXXXX`

Donde:
- El primer carácter es la letra 'A' o 'B' (mayúscula)
- Seguido de al menos 5 dígitos numéricos
- Ejemplos válidos: `A00001`, `B00001`, `A99999`, `B12345`

## Reglas de Negocio

### Miembros Individuales (Prefijo B)
- Son miembros independientes sin familia asociada
- El campo `MembershipType` debe ser `"individual"`
- No requieren una entidad `Family` relacionada
- Ejemplos: `B00001`, `B99001`, `B99002`

### Miembros Familiares (Prefijo A)
- Son miembros que pertenecen a una unidad familiar
- El campo `MembershipType` debe ser `"familiar"`
- DEBEN tener una entidad `Family` asociada
- La familia debe tener al menos 2 miembros (esposo y esposa)
- Ejemplos: `A00001`, `A00002`, `A99001`

## Implementación en el Código

### Modelos
- `Member.MembershipNumber`: Campo que almacena el número de membresía
- `Family.NumeroSocio`: Campo que identifica a la familia (también usa prefijo A)

### Seeds de Desarrollo
Los datos de prueba en el entorno de desarrollo siguen esta convención:
- `B99001` - `B99005`: Miembros individuales de prueba
- No se crean familias de prueba para mantener simplicidad

### Validación
El sistema debe validar que:
1. El formato del número sea correcto
2. El prefijo corresponda con el tipo de membresía
3. Para miembros con prefijo A, exista una familia asociada

## Histórico de Cambios

- **2024-01**: Implementación inicial con convención correcta
- **2024-12**: Corrección de seeds de desarrollo para usar prefijos correctos

## Notas Importantes

⚠️ **IMPORTANTE**: Esta convención es crítica para el correcto funcionamiento del sistema. Cualquier cambio debe ser coordinado con todas las partes del sistema que dependan de esta numeración.
