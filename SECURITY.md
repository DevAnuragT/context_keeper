# Security Guide

## 🔒 Security Features

### Authentication & Authorization
- **JWT Tokens**: HMAC-SHA256 signed tokens with 24-hour expiration
- **GitHub OAuth**: Secure OAuth 2.0 flow with scope verification
- **Token Validation**: Comprehensive JWT validation with signature verification
- **Clock Skew Tolerance**: 60-second tolerance for time synchronization issues

### Data Protection
- **Parameterized Queries**: All database queries use parameterized statements to prevent SQL injection
- **SSL/TLS**: Database connections require SSL by default
- **Sensitive Data**: GitHub tokens are not stored in JWT tokens for security
- **Input Validation**: All user inputs are validated and sanitized

### HTTP Security
- **Security Headers**: Comprehensive security headers including CSP, HSTS, X-Frame-Options
- **CORS**: Configurable CORS with origin restrictions
- **Content Type**: Proper content-type headers to prevent MIME sniffing attacks

### Infrastructure Security
- **Non-root User**: Docker containers run as non-root user
- **Resource Limits**: Memory and CPU limits configured
- **Secrets Management**: Sensitive data stored in Docker secrets
- **SSL Required**: Database connections require SSL encryption

## 🚨 Security Configuration

### Required Environment Variables

```bash
# Database (SSL required)
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require

# JWT Secret (generate with: openssl rand -hex 32)
JWT_SECRET=your-cryptographically-secure-secret-here

# GitHub OAuth
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
GITHUB_REDIRECT_URL=https://yourdomain.com/api/auth/github

# CORS Origins (comma-separated)
ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
```

### Docker Secrets Setup

Create secret files:
```bash
mkdir -p secrets
echo "your-secure-postgres-password" > secrets/postgres_password.txt
echo "your-jwt-secret-key" > secrets/jwt_secret.txt
echo "your-github-client-secret" > secrets/github_client_secret.txt
```

### Production Checklist

- [ ] Generate cryptographically secure JWT secret (32+ bytes)
- [ ] Use strong database passwords
- [ ] Enable SSL for all database connections
- [ ] Configure proper CORS origins (no wildcards)
- [ ] Set up proper TLS certificates
- [ ] Configure rate limiting (reverse proxy)
- [ ] Set up monitoring and alerting
- [ ] Regular security updates
- [ ] Backup encryption
- [ ] Network segmentation

## 🛡️ Security Best Practices

### Development
- Never commit secrets to version control
- Use environment variables for all configuration
- Test with SSL enabled locally
- Validate all inputs
- Use parameterized queries only

### Deployment
- Use HTTPS everywhere
- Configure proper firewall rules
- Regular security updates
- Monitor for suspicious activity
- Implement proper logging
- Use secrets management systems

### Monitoring
- Failed authentication attempts
- Unusual API usage patterns
- Database connection errors
- SSL certificate expiration
- Resource usage anomalies

## 🚨 Incident Response

### If Compromised
1. **Immediate**: Rotate all secrets (JWT, database, GitHub OAuth)
2. **Assess**: Check logs for unauthorized access
3. **Notify**: Inform affected users if data was accessed
4. **Patch**: Apply security updates
5. **Monitor**: Increased monitoring for suspicious activity

### Reporting Security Issues
Please report security vulnerabilities to: security@yourdomain.com

## 📋 Security Audit Log

### Fixed Vulnerabilities
- ✅ Weak default JWT secret → Cryptographically secure generation
- ✅ SSL disabled by default → SSL required for database connections
- ✅ Overly permissive CORS → Configurable origin restrictions
- ✅ Hardcoded credentials → Docker secrets management
- ✅ Missing security headers → Comprehensive security headers
- ✅ GitHub token in JWT → Removed sensitive data from tokens
- ✅ Docker root user → Non-root user with proper permissions
- ✅ No resource limits → Memory and CPU limits configured

### Remaining Considerations
- Rate limiting (implement at reverse proxy level)
- Session management (consider refresh tokens for longer sessions)
- API versioning (for future security updates)
- Audit logging (detailed security event logging)