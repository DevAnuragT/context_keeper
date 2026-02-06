# Authentication System Documentation

## Overview

The MCP Context Engine authentication system provides secure user authentication with multiple login methods, password management, and email verification. The system is built with security best practices and follows modern web standards.

## Features

### 🔐 Authentication Methods
- **Email/Password**: Traditional authentication with strong password requirements
- **OAuth Providers**: 
  - GitHub
  - Google
  - Slack

### 🛡️ Security Features
- **Password Requirements**:
  - Minimum 8 characters
  - At least one uppercase letter
  - At least one lowercase letter
  - At least one number
  - Password strength indicator

- **Session Management**:
  - JWT token-based authentication
  - Secure token storage in localStorage
  - CSRF protection for OAuth flows
  - Automatic token validation

- **Email Verification**:
  - Required for new accounts
  - Resend verification email option
  - Token-based verification

- **Password Reset**:
  - Secure token-based reset flow
  - Email delivery of reset links
  - Token expiration handling

## File Structure

```
web/
├── auth.js                      # Authentication module (shared)
├── auth.css                     # Authentication styles (shared)
├── login.html                   # Login page
├── signup.html                  # Signup page
├── password-reset.html          # Password reset request page
├── password-reset-confirm.html  # Password reset confirmation page
├── verify-email.html            # Email verification page
└── AUTH_README.md              # This file
```

## Pages

### 1. Login Page (`/login`)
**Purpose**: Authenticate existing users

**Features**:
- Email/password form
- OAuth login buttons (GitHub, Google, Slack)
- "Remember me" checkbox
- "Forgot password?" link
- Link to signup page

**API Endpoint**: `POST /api/auth/login`

**Request**:
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123"
}
```

**Response**:
```json
{
  "token": "jwt_token_here",
  "user": {
    "id": "user_id",
    "email": "user@example.com",
    "name": "User Name",
    "verified": true
  }
}
```

### 2. Signup Page (`/signup`)
**Purpose**: Create new user accounts

**Features**:
- Name, email, and password fields
- Password strength indicator
- Terms of Service acceptance
- OAuth signup buttons
- Link to login page

**API Endpoint**: `POST /api/auth/signup`

**Request**:
```json
{
  "name": "John Doe",
  "email": "user@example.com",
  "password": "SecurePassword123"
}
```

**Response**:
```json
{
  "token": "jwt_token_here",
  "user": {
    "id": "user_id",
    "email": "user@example.com",
    "name": "John Doe",
    "verified": false
  },
  "message": "Please verify your email address"
}
```

### 3. Password Reset Request (`/password-reset`)
**Purpose**: Request password reset link

**Features**:
- Email input field
- Success confirmation
- Resend link option
- Back to login link

**API Endpoint**: `POST /api/auth/password-reset/request`

**Request**:
```json
{
  "email": "user@example.com"
}
```

**Response**:
```json
{
  "message": "Password reset email sent"
}
```

### 4. Password Reset Confirmation (`/password-reset-confirm`)
**Purpose**: Set new password using reset token

**Features**:
- New password input with strength indicator
- Confirm password field
- Token validation
- Success/error states

**API Endpoint**: `POST /api/auth/password-reset/confirm`

**Request**:
```json
{
  "token": "reset_token_from_email",
  "password": "NewSecurePassword123"
}
```

**Response**:
```json
{
  "message": "Password reset successful"
}
```

### 5. Email Verification (`/verify-email`)
**Purpose**: Verify user email address

**Features**:
- Automatic verification when token present
- Pending state with resend option
- Success/error states
- Link to dashboard or login

**API Endpoint**: `POST /api/auth/verify-email`

**Request**:
```json
{
  "token": "verification_token_from_email"
}
```

**Response**:
```json
{
  "message": "Email verified successfully"
}
```

**Resend Endpoint**: `POST /api/auth/verify-email/resend`

**Request**:
```json
{
  "email": "user@example.com"
}
```

## OAuth Flow

### 1. Initiate OAuth
User clicks OAuth button → System generates CSRF token → Redirects to provider

```javascript
authManager.initiateOAuthFlow('github');
// Redirects to: /api/auth/oauth/github?state=csrf_token
```

### 2. OAuth Callback
Provider redirects back → System validates CSRF token → Exchanges code for token

**Callback URL**: `/login?code=auth_code&state=csrf_token`

**API Endpoint**: `POST /api/auth/oauth/callback`

**Request**:
```json
{
  "code": "authorization_code",
  "state": "csrf_token"
}
```

**Response**:
```json
{
  "token": "jwt_token_here",
  "user": {
    "id": "user_id",
    "email": "user@example.com",
    "name": "User Name",
    "verified": true
  }
}
```

## Authentication Module (`auth.js`)

### AuthManager Class

**Methods**:

#### Validation
- `validateEmail(email)` - Validates email format
- `validatePassword(password)` - Validates password strength
- `getPasswordStrength(password)` - Returns password strength level

#### Session Management
- `setSession(token, userData)` - Stores authentication session
- `getToken()` - Retrieves stored JWT token
- `getUserData()` - Retrieves stored user data
- `clearSession()` - Clears authentication session
- `isAuthenticated()` - Checks if user is authenticated

#### API Calls
- `signup(email, password, name)` - Create new account
- `login(email, password)` - Authenticate user
- `logout()` - End user session
- `requestPasswordReset(email)` - Request password reset
- `resetPassword(token, newPassword)` - Confirm password reset
- `verifyEmail(token)` - Verify email address
- `resendVerificationEmail(email)` - Resend verification email

#### OAuth
- `initiateOAuthFlow(provider)` - Start OAuth flow
- `handleOAuthCallback()` - Process OAuth callback
- `generateCSRFToken()` - Generate CSRF protection token
- `validateCSRFToken(token)` - Validate CSRF token

### AuthUI Class

**Methods**:
- `showError(formElement, message)` - Display error message
- `showSuccess(formElement, message)` - Display success message
- `clearMessages(formElement)` - Clear all messages
- `setLoading(button, loading)` - Toggle button loading state
- `updatePasswordStrength(password, element)` - Update strength indicator
- `showFieldError(inputElement, message)` - Show field-specific error
- `clearFieldError(inputElement)` - Clear field error

## Styling (`auth.css`)

### Key Classes

**Container**:
- `.auth-container` - Full-page container with gradient background
- `.auth-card` - White card containing auth form

**Form Elements**:
- `.form-group` - Form field container
- `.form-label` - Field label
- `.form-input` - Text/email/password input
- `.form-checkbox` - Checkbox with label
- `.form-error` - Error message banner
- `.form-success` - Success message banner
- `.field-error` - Field-specific error

**Buttons**:
- `.auth-submit` - Primary submit button
- `.oauth-button` - OAuth provider button
- `.oauth-button.github` - GitHub-styled button
- `.oauth-button.google` - Google-styled button
- `.oauth-button.slack` - Slack-styled button

**States**:
- `.auth-loading` - Loading state with spinner
- `.auth-success-state` - Success state with icon
- `.auth-error-state` - Error state with icon

**Password Strength**:
- `.password-strength` - Container for strength indicator
- `.strength-bar-container` - Progress bar container
- `.strength-bar` - Animated progress bar
- `.strength-text` - Strength level text

## Backend Integration

### Required API Endpoints

The frontend expects the following backend endpoints:

1. **POST /api/auth/signup** - Create new user account
2. **POST /api/auth/login** - Authenticate user
3. **POST /api/auth/logout** - End user session
4. **POST /api/auth/password-reset/request** - Request password reset
5. **POST /api/auth/password-reset/confirm** - Confirm password reset
6. **POST /api/auth/verify-email** - Verify email address
7. **POST /api/auth/verify-email/resend** - Resend verification email
8. **GET /api/auth/oauth/{provider}** - Initiate OAuth flow
9. **POST /api/auth/oauth/callback** - Handle OAuth callback

### JWT Token Format

The backend should return JWT tokens with the following structure:

```json
{
  "sub": "user_id",
  "email": "user@example.com",
  "name": "User Name",
  "verified": true,
  "iat": 1234567890,
  "exp": 1234567890
}
```

### Error Response Format

All API endpoints should return errors in this format:

```json
{
  "error": true,
  "message": "Human-readable error message",
  "code": "ERROR_CODE"
}
```

## Security Considerations

### Client-Side Security
1. **CSRF Protection**: OAuth flows use state parameter for CSRF protection
2. **Input Validation**: All inputs validated before submission
3. **Password Strength**: Enforced minimum password requirements
4. **XSS Prevention**: All user inputs are properly escaped
5. **Token Storage**: JWT tokens stored in localStorage (consider httpOnly cookies for production)

### Backend Requirements
1. **Rate Limiting**: Implement rate limiting on all auth endpoints
2. **Token Expiration**: JWT tokens should expire (recommended: 1 hour)
3. **Refresh Tokens**: Implement refresh token mechanism for long sessions
4. **Email Verification**: Require email verification before full access
5. **Password Hashing**: Use bcrypt or similar for password hashing
6. **HTTPS Only**: All authentication must use HTTPS in production
7. **Token Revocation**: Implement token blacklist for logout
8. **Brute Force Protection**: Lock accounts after failed login attempts

## Usage Examples

### Basic Login Flow

```javascript
const authManager = new AuthManager();

// Login
const result = await authManager.login('user@example.com', 'password123');
if (result.success) {
  // Redirect to dashboard
  window.location.href = '/dashboard';
} else {
  // Show error
  console.error(result.error);
}
```

### OAuth Login Flow

```javascript
const authManager = new AuthManager();

// Initiate OAuth (redirects to provider)
authManager.initiateOAuthFlow('github');

// After redirect back, handle callback
const result = await authManager.handleOAuthCallback();
if (result.success) {
  window.location.href = '/dashboard';
}
```

### Check Authentication Status

```javascript
const authManager = new AuthManager();

if (authManager.isAuthenticated()) {
  const userData = authManager.getUserData();
  console.log('Logged in as:', userData.email);
} else {
  window.location.href = '/login';
}
```

## Testing

### Manual Testing Checklist

**Login Page**:
- [ ] Email validation works
- [ ] Password validation works
- [ ] OAuth buttons redirect correctly
- [ ] Remember me checkbox functions
- [ ] Forgot password link works
- [ ] Error messages display correctly
- [ ] Success redirects to dashboard

**Signup Page**:
- [ ] All fields validate correctly
- [ ] Password strength indicator updates
- [ ] Terms checkbox required
- [ ] OAuth buttons work
- [ ] Success shows verification message
- [ ] Duplicate email handled

**Password Reset**:
- [ ] Email validation works
- [ ] Success state displays
- [ ] Resend link functions
- [ ] Invalid token shows error
- [ ] New password validates
- [ ] Success redirects to login

**Email Verification**:
- [ ] Token verification works
- [ ] Resend email functions
- [ ] Success redirects to dashboard
- [ ] Invalid token shows error

### Automated Testing

Consider implementing:
- Unit tests for validation functions
- Integration tests for API calls
- E2E tests for complete flows
- Security tests for XSS/CSRF

## Customization

### Styling
Modify `auth.css` to match your brand:
- Update CSS variables in `:root`
- Change gradient backgrounds
- Adjust spacing and sizing
- Customize button styles

### Validation Rules
Modify validation in `auth.js`:
- Change password requirements
- Add custom email validation
- Implement additional field validation

### OAuth Providers
Add new OAuth providers:
1. Add button HTML in auth pages
2. Add button styles in `auth.css`
3. Add click handler in page script
4. Implement backend OAuth flow

## Troubleshooting

### Common Issues

**OAuth not working**:
- Check CSRF token generation
- Verify callback URL configuration
- Check provider credentials

**Token not persisting**:
- Check localStorage availability
- Verify token format
- Check for CORS issues

**Email not sending**:
- Verify backend email configuration
- Check spam folder
- Verify email service credentials

**Styling issues**:
- Check CSS file loading
- Verify font loading
- Check responsive breakpoints

## Future Enhancements

Potential improvements:
- [ ] Two-factor authentication (2FA)
- [ ] Biometric authentication
- [ ] Social login (Twitter, LinkedIn)
- [ ] Magic link authentication
- [ ] Session management dashboard
- [ ] Security audit log
- [ ] Device management
- [ ] Account recovery options

## Support

For issues or questions:
- Check backend API logs
- Review browser console errors
- Verify network requests
- Test with different browsers
- Check mobile responsiveness

## License

This authentication system is part of the MCP Context Engine project.
