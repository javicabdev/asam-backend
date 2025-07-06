# Problema de Verificación de Email Múltiple

## Descripción del Problema

Después de verificar exitosamente un email, cuando el usuario navega al dashboard, el sistema intenta verificar el token nuevamente, lo que resulta en el error:

```
"INVALID_TOKEN: verification token has already been used"
```

## Causa Raíz

El problema está en el frontend, no en el backend. Posibles causas:

1. **Token almacenado en estado/localStorage**: El token se está guardando y reenviando automáticamente
2. **Router/Guards**: Algún guard o middleware está causando múltiples verificaciones
3. **Componente que se re-renderiza**: Un componente está llamando a la mutación `verifyEmail` múltiples veces

## Soluciones Implementadas en el Backend

### 1. Manejo Mejorado de Tokens Usados

He actualizado el resolver `VerifyEmail` para manejar mejor el caso cuando un token ya fue usado:

```go
// Si el error es "token already used", devolvemos éxito con mensaje apropiado
if appErr.Code == errors.ErrInvalidToken && appErr.Message == "verification token has already been used" {
    msg := "Email is already verified"
    return &model.MutationResponse{
        Success: true,
        Message: &msg,
    }, nil
}
```

Esto evita que el frontend muestre un error cuando el email ya está verificado.

### 2. Scripts de Utilidad

Se han creado scripts para diagnosticar y solucionar problemas:

- `check-email-verification-status.ps1`: Verifica el estado de verificación de un usuario
- `manually-verify-user.ps1`: Marca manualmente un usuario como verificado

## Recomendaciones para el Frontend

### 1. Limpiar el Token Después de Usarlo

```javascript
// Después de verificar exitosamente
const { data } = await verifyEmail({ variables: { token } });
if (data.verifyEmail.success) {
    // Limpiar el token del estado/localStorage
    localStorage.removeItem('verificationToken');
    // Limpiar query params
    router.replace('/dashboard');
}
```

### 2. Verificar Estado Antes de Intentar Verificar

```javascript
// Antes de llamar verifyEmail
const currentUser = await getCurrentUser();
if (currentUser.emailVerified) {
    // Ya está verificado, ir directo al dashboard
    router.push('/dashboard');
    return;
}
```

### 3. Evitar Múltiples Llamadas

```javascript
// Usar un flag para evitar llamadas múltiples
const [isVerifying, setIsVerifying] = useState(false);

const handleVerification = async (token) => {
    if (isVerifying) return;
    
    setIsVerifying(true);
    try {
        await verifyEmail({ variables: { token } });
    } finally {
        setIsVerifying(false);
    }
};
```

### 4. Manejo de Rutas

```javascript
// En el router/guard
if (route.path === '/verify-email' && user.emailVerified) {
    // Redirigir si ya está verificado
    return redirect('/dashboard');
}
```

## Flujo Correcto de Verificación

1. Usuario recibe email con link: `/verify-email?token=XXX`
2. Frontend extrae el token y llama a `verifyEmail` mutation
3. Backend verifica el token y marca el usuario como verificado
4. Frontend recibe respuesta exitosa
5. Frontend limpia el token y redirige al dashboard
6. Dashboard verifica que el usuario está verificado y muestra el contenido

## Testing

Para probar que todo funciona correctamente:

1. Ejecutar el script de verificación de estado:
   ```powershell
   .\scripts\check-email-verification-status.ps1
   ```

2. Si el usuario no está verificado, usar el script manual:
   ```powershell
   .\scripts\manually-verify-user.ps1
   ```

3. Verificar que el frontend no está guardando/reenviando el token

## Conclusión

El backend ahora maneja graciosamente los intentos de verificación múltiple, pero el frontend debe ser actualizado para evitar estos intentos innecesarios y mejorar la experiencia del usuario.
