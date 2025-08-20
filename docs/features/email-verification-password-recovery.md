# Email Verification and Password Recovery

## Overview

ASAM backend now includes email verification and password recovery features for users with email usernames.

## Features

### 1. Email Verification
- Automatic verification email sent when creating user with email username
- Users can request verification email resend
- 24-hour token expiration
- Email verified status in user profile

### 2. Password Recovery
- Request password reset via email
- 1-hour token expiration
- Rate limiting: max 3 requests per hour
- Email notification after successful password change

## Database Schema

### User Model Updates
```sql
email_verified BOOLEAN NOT NULL DEFAULT FALSE
email_verified_at TIMESTAMP NULL
```

### Verification Tokens Table
```sql
CREATE TABLE verification_tokens (
    id BIGINT PRIMARY KEY,
    token VARCHAR(64) UNIQUE NOT NULL,
    user_id BIGINT NOT NULL,
    type VARCHAR(20) NOT NULL,
    email VARCHAR(100) NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);
```

## GraphQL API

### Mutations

#### Send Verification Email
```graphql
mutation {
  sendVerificationEmail {
    success
    message
    error
  }
}
```

#### Verify Email
```graphql
mutation {
  verifyEmail(token: "verification_token_here") {
    success
    message
    error
  }
}
```

#### Resend Verification Email
```graphql
mutation {
  resendVerificationEmail(email: "user@example.com") {
    success
    message
    error
  }
}
```

#### Request Password Reset
```graphql
mutation {
  requestPasswordReset(email: "user@example.com") {
    success
    message
    error
  }
}
```

#### Reset Password With Token
```graphql
mutation {
  resetPasswordWithToken(
    token: "reset_token_here"
    newPassword: "NewSecurePass123!"
  ) {
    success
    message
    error
  }
}
```

## Email Service

### Mock Email Service (Development)
The system includes a mock email service that prints emails to console during development.

### Email Templates

#### Verification Email
- Subject: "Verifica tu cuenta en ASAM"
- Contains verification link
- 24-hour expiration notice

#### Password Reset Email
- Subject: "Recuperación de contraseña - ASAM"
- Contains reset link
- 1-hour expiration notice

#### Password Changed Email
- Subject: "Tu contraseña ha sido cambiada - ASAM"
- Security notification

## Security Considerations

### Email Enumeration Prevention
- Always returns success for password reset requests
- Doesn't reveal if email exists in system
- Logs warnings for non-existent emails

### Token Security
- 64-character cryptographically secure tokens
- One-time use enforcement
- Automatic cleanup of expired tokens
- Rate limiting on password reset requests

### Best Practices
1. **Always use HTTPS** for verification/reset links
2. **Short expiration times** (24h for verification, 1h for reset)
3. **Rate limiting** to prevent abuse
4. **One-time tokens** that are marked as used
5. **Email notifications** for security events

## Configuration

### Environment Variables
```env
# Base URL for email links
BASE_URL=https://asam.example.com

# Email Service Configuration (for production)
EMAIL_SERVICE=smtp
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=noreply@asam.example.com
SMTP_PASS=your-app-password
SMTP_FROM=ASAM <noreply@asam.example.com>
```

## Implementation Guide

### 1. Run Migrations
```bash
# Apply the email verification migration
mysql -u root -p asam_db < migrations/004_add_email_verification.sql
```

### 2. Update Configuration
Add base URL to your configuration for building email links.

### 3. Set Up Email Service
- Development: Uses MockEmailService (console output)
- Production: Implement real email service (SMTP, SendGrid, etc.)

### 4. Test the Flow

#### Email Verification Flow
1. Create user with email username
2. Check console for verification email (dev)
3. Click verification link
4. Verify email status updated

#### Password Reset Flow
1. Request password reset
2. Check console for reset email (dev)
3. Click reset link
4. Set new password
5. Verify login with new password

## Future Enhancements

1. **Email Templates**: HTML email templates with branding
2. **Multi-language Support**: Emails in user's preferred language
3. **2FA Integration**: Two-factor authentication after email verification
4. **Email Change**: Allow users to change email with verification
5. **Audit Trail**: Log all email/password changes for security

## Troubleshooting

### Common Issues

1. **Token Not Found**
   - Token may be expired
   - Token already used
   - Invalid token format

2. **Email Not Sending**
   - Check email service configuration
   - Verify SMTP credentials
   - Check spam folder

3. **Rate Limit Exceeded**
   - Wait before requesting another reset
   - Default: 3 requests per hour

### Debug Mode
Enable debug logging to see email content in logs:
```go
logger.SetLevel(zap.DebugLevel)
```
