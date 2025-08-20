# Email as Username Feature

## Overview

The ASAM backend now supports using email addresses as usernames. This provides more flexibility for user authentication and allows users to log in with their email addresses.

## Features

### 1. Username Validation
- Supports both traditional usernames and email addresses
- Traditional usernames: 3-100 characters, alphanumeric with `_`, `-`, `.`
- Email usernames: Must be valid email format, max 100 characters total

### 2. Email Validation Rules
- Must contain exactly one `@` symbol
- Local part (before @): 1-64 characters
- Domain part (after @): 3-255 characters
- No consecutive dots (..)
- Cannot start or end with dots
- Must have a valid TLD (e.g., .com, .org, .co.uk)

### 3. Utility Functions

Located in `pkg/utils/email.go`:

- `IsEmail(str string) bool` - Checks if a string is a valid email
- `NormalizeEmail(email string) string` - Converts to lowercase and trims
- `ExtractUsernameFromEmail(email string) string` - Gets the local part
- `ObfuscateEmail(email string) string` - Partially hides email for privacy

## Database Changes

### Migration Required
Run migration `003_increase_username_size.sql` to update the username field size:

```sql
-- For MySQL/MariaDB
ALTER TABLE users MODIFY COLUMN username VARCHAR(100) NOT NULL;

-- For PostgreSQL
ALTER TABLE users ALTER COLUMN username TYPE VARCHAR(100);
```

## Usage Examples

### Creating a User with Email

```go
// Traditional username
user, err := userService.CreateUser(ctx, "john_doe", "SecurePass123!", models.RoleUser)

// Email as username
user, err := userService.CreateUser(ctx, "john.doe@example.com", "SecurePass123!", models.RoleUser)
```

### Login with Email

```graphql
mutation {
  login(username: "john.doe@example.com", password: "SecurePass123!") {
    accessToken
    refreshToken
    user {
      id
      username
      role
    }
  }
}
```

## Testing

Run the email functionality tests:

```bash
# Windows PowerShell
.\scripts\test-email-functionality.ps1

# Linux/Mac
go test -v ./test/pkg/utils/... ./test/internal/domain/services/... -run '(TestUserService.*Email|TestEmail)'
```

## Security Considerations

1. **Email Privacy**: Use `ObfuscateEmail()` when displaying emails in logs or UI
2. **Case Sensitivity**: Emails are normalized to lowercase for consistency
3. **Validation**: Strict validation prevents email injection attacks
4. **Database**: Unique index on username prevents duplicate emails

## Best Practices

1. **Normalization**: Always normalize emails before storing:
   ```go
   normalizedEmail := utils.NormalizeEmail(userInput)
   ```

2. **Display**: When showing emails in UI:
   ```go
   displayEmail := utils.ObfuscateEmail(user.Username)
   // Shows: j***e@example.com
   ```

3. **Type Detection**: Check if username is email:
   ```go
   if utils.IsEmail(username) {
       // Handle as email
   } else {
       // Handle as regular username
   }
   ```

## Future Enhancements

1. **Email Verification**: Add email verification workflow
2. **Multiple Emails**: Allow users to have multiple email addresses
3. **OAuth Integration**: Use email for social login matching
4. **Password Recovery**: Implement password reset via email
