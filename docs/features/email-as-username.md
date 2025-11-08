# Email as Username Feature

## Overview

The ASAM backend supports using email addresses as usernames. This provides flexibility for user authentication and allows users to log in with their email addresses instead of traditional usernames.

## Features

### 1. Username Validation
- Supports both traditional usernames and email addresses
- Traditional usernames: 3-100 characters, alphanumeric with `_`, `-`, `.`
- Email usernames: Must be valid email format (basic validation), max 100 characters total
- Validation automatically detects format based on presence of `@` symbol

### 2. Email Validation
- Must contain `@` symbol
- Basic email format validation using regex pattern
- Pattern: `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
- Automatically normalized to lowercase for storage and comparison

### 3. Utility Functions

Located in `pkg/utils/email.go`:

- `IsEmail(str string) bool` - Checks if a string is a valid email using regex
- `NormalizeEmail(email string) string` - Converts to lowercase and trims whitespace
- `ExtractUsernameFromEmail(email string) string` - Extracts the local part (before @)
- `ObfuscateEmail(email string) string` - Partially hides email for privacy (e.g., `j***e@example.com`)

**Note:** These utility functions are available for general use, but the user service uses its own internal validation methods.

## Database Schema

### Current Schema
The `users` table already supports email addresses as usernames:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,  -- Supports both usernames and emails
    email VARCHAR(255) NOT NULL UNIQUE,     -- Separate email field
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    ...
);
```

**No migration required** - The initial schema already allocates sufficient space (255 characters) for email addresses.

## Implementation Details

### Username Validation Flow

Implemented in `internal/domain/services/user_service.go`:

```go
// validateUsername (line 750) - Main validation entry point
func (s *userService) validateUsername(username string) error {
    // 1. Trim whitespace and check length (3-100 characters)
    // 2. Detect if username contains '@'
    // 3. If contains '@': validate as email
    // 4. Otherwise: validate as regular username
}

// validateEmail (line 802) - Email format validation
func (s *userService) validateEmail(email string) error {
    // Uses regex pattern for basic email validation
    // Normalizes to lowercase before storage
}

// validateRegularUsername (line 784) - Username validation
func (s *userService) validateRegularUsername(username string) error {
    // Allows: a-z, A-Z, 0-9, underscore, hyphen, dot
}
```

### Login Flow with Dual Support

Implemented in `internal/domain/services/auth_service.go`:

```go
// validateCredentials (line 145) - Dual login support
func (s *authService) validateCredentials(ctx, usernameOrEmail, password) error {
    // 1. Try to find user by username field
    user := userRepo.FindByUsername(usernameOrEmail)

    // 2. If not found, try to find by email field
    if user == nil {
        user = userRepo.FindByEmail(usernameOrEmail)
    }

    // 3. Validate password if user found
    // 4. Return error if user not found or password invalid
}
```

**Benefits:**
- Users can login with whatever they remember (username or email)
- Backward compatible - existing logins continue working
- No changes needed in client applications

## Usage Examples

### Creating Users via GraphQL

```graphql
# Traditional username
mutation {
  createUser(input: {
    username: "john_doe"
    email: "john@example.com"
    password: "SecurePass123!"
    role: USER
  }) {
    id
    username
    email
  }
}

# Email as username
mutation {
  createUser(input: {
    username: "john.doe@example.com"
    email: "john.doe@example.com"
    password: "SecurePass123!"
    role: USER
  }) {
    id
    username
    email
  }
}
```

### Login with Email or Username

Both approaches work interchangeably:

```graphql
# Login with username
mutation {
  login(username: "john_doe", password: "SecurePass123!") {
    accessToken
    refreshToken
    user {
      id
      username
      email
      role
    }
  }
}

# Login with email (works even if username is different)
mutation {
  login(username: "john@example.com", password: "SecurePass123!") {
    accessToken
    refreshToken
    user {
      id
      username
      email
      role
    }
  }
}
```

**Example Scenario:**
```
User in database:
  username: "john_doe"
  email: "john@example.com"

Both of these will work:
  ✅ login(username: "john_doe", ...)
  ✅ login(username: "john@example.com", ...)
```

## Security Considerations

1. **Case Insensitivity**: Email addresses are automatically normalized to lowercase before storage to ensure case-insensitive matching
2. **Unique Constraint**: Database enforces unique constraint on username field, preventing duplicate emails
3. **Validation**: Basic regex validation prevents malformed email addresses
4. **Separate Email Field**: Users table maintains both `username` and `email` fields - when using email as username, typically both fields contain the same value

## Best Practices

### For API Consumers

1. **Normalization**: The backend automatically normalizes emails to lowercase
   ```graphql
   # Input: John.Doe@Example.COM
   # Stored as: john.doe@example.com
   ```

2. **Validation**: Both formats are accepted in username field:
   - Traditional: `john_doe`
   - Email: `john.doe@example.com`

3. **Dual Login Support**: Users can login with EITHER `username` OR `email`
   - Login with username: `john_doe`
   - Login with email: `john@example.com`
   - The system automatically tries both fields for maximum flexibility

### For Developers

1. **Using Utility Functions**: Available in `pkg/utils/email.go` for general email handling:
   ```go
   import "github.com/javicabdev/asam-backend/pkg/utils"

   // Check if string is email
   if utils.IsEmail(input) {
       normalized := utils.NormalizeEmail(input)
   }

   // Obfuscate for display
   safe := utils.ObfuscateEmail("john@example.com")
   // Returns: "j***n@example.com"
   ```

2. **Service Validation**: User service has built-in validation - no need to pre-validate:
   ```go
   // Service handles validation internally
   user, err := userService.CreateUser(ctx, username, password, role)
   ```

## Limitations

1. **Basic Validation**: Uses simple regex pattern, doesn't validate if email domain actually exists
2. **No MX Record Check**: Doesn't verify if email domain has valid MX records
3. **No Verification**: Email addresses are not verified (no confirmation email sent)
4. **Sequential Lookup**: Login tries username first, then email - slight performance impact for email-based logins

## Future Enhancements

1. **Email Verification**: Implement email verification workflow with confirmation tokens
2. **Enhanced Validation**: Add MX record validation and disposable email detection
3. **Password Recovery**: Implement password reset via email functionality (already implemented: `resetPasswordWithToken`)
4. **OAuth Integration**: Use email for social login matching
5. **Smart Login Optimization**: Detect if input looks like email and try email lookup first to reduce database queries
