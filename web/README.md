# MCP Context Engine - Landing Page & Marketing Site

This directory contains the landing page and marketing website for the MCP Context Engine platform.

## Overview

The landing page is designed to:
- Explain the product value proposition
- Showcase integrations (GitHub, Slack, Discord)
- Highlight developer use cases
- Provide a secure signup funnel
- Implement responsive design and SEO optimization

## Files Structure

```
web/
├── index.html          # Main landing page
├── signup.html         # User signup page
├── login.html          # User login page
├── styles.css          # Main stylesheet
├── auth.css            # Authentication pages stylesheet
├── script.js           # Main JavaScript functionality
├── auth.js             # Authentication JavaScript
└── README.md           # This file
```

## Features

### Landing Page (index.html)

#### Sections:
1. **Hero Section**
   - Compelling headline with gradient text
   - Value proposition
   - CTA buttons (Start Free Trial, Watch Demo)
   - Interactive code window demonstration
   - Trust signals (security, setup time, no credit card)

2. **Problem Statement**
   - Highlights common developer pain points
   - Scattered conversations
   - Lost knowledge
   - Slow onboarding
   - Context switching

3. **Features Section**
   - Semantic search across project knowledge
   - File context history
   - Decision tracking
   - Code reasoning
   - Architecture discussions
   - Knowledge graph visualization
   - Each feature includes example MCP tool calls

4. **Integrations Section**
   - GitHub integration details
   - Slack integration details
   - Discord integration details
   - Feature lists for each platform
   - Visual logos and branding

5. **Use Cases Section**
   - For Developers
   - For Team Leads
   - For Engineering Managers
   - Specific benefits for each persona

6. **Security Section**
   - End-to-end encryption
   - Tenant isolation
   - OAuth 2.0 authentication
   - Role-based access control
   - Audit logging
   - SOC 2 compliance

7. **CTA Section**
   - Final call-to-action
   - Multiple conversion paths
   - Trust reinforcement

8. **Footer**
   - Navigation links
   - Legal links
   - Social media links
   - Company information

### Signup Page (signup.html)

#### Features:
- OAuth login options (GitHub, Google)
- Email/password signup form
- Password strength indicator
- Company information (optional)
- Terms of service agreement
- Newsletter opt-in
- Trust signals (SSL, SOC 2, GDPR)
- Benefits sidebar with:
  - Feature highlights
  - Customer testimonial
  - Trial information

### Login Page (login.html)

#### Features:
- OAuth login options (GitHub, Google)
- Email/password login form
- Remember me checkbox
- Forgot password link
- Link to signup page
- Trust signals

## Design Principles

### Responsive Design
- Mobile-first approach
- Breakpoints at 768px and 480px
- Flexible grid layouts
- Touch-friendly interactive elements
- Optimized for all screen sizes

### SEO Optimization
- Semantic HTML5 structure
- Meta tags for description, keywords, author
- Open Graph tags for social sharing
- Twitter Card tags
- Descriptive alt text (when images added)
- Clean URL structure
- Fast loading times
- Accessible navigation

### Accessibility
- ARIA labels for interactive elements
- Keyboard navigation support
- Focus states for all interactive elements
- Semantic HTML structure
- Color contrast compliance
- Screen reader friendly

### Performance
- Minimal external dependencies
- Optimized CSS with CSS Grid and Flexbox
- Efficient JavaScript with event delegation
- Lazy loading for images (when added)
- Prefetching for important pages
- Service Worker ready (commented out)

## JavaScript Functionality

### Main Features (script.js)
- Mobile menu toggle
- Smooth scrolling for anchor links
- Navbar scroll effects
- Intersection Observer for animations
- Code window animation
- Analytics event tracking
- Demo video modal
- Performance monitoring
- Prefetching on hover

### Authentication Features (auth.js)
- Form validation
- Password strength calculation
- OAuth flow initiation
- Signup/login API integration
- Error handling and display
- Success messages
- Loading states
- Email verification resend
- Password reset functionality

## Styling

### Color Scheme
- Primary: #6366f1 (Indigo)
- Secondary: #10b981 (Green)
- Text: #1f2937 (Dark Gray)
- Background: #ffffff (White)
- Gradients for visual interest

### Typography
- Font: Inter (Google Fonts)
- Weights: 300, 400, 500, 600, 700
- Responsive font sizes
- Clear hierarchy

### Components
- Buttons (primary, secondary, large)
- Cards (problem, feature, integration, use case, security)
- Forms (input fields, checkboxes, validation)
- Navigation (fixed navbar, mobile menu)
- Modals (demo video)
- Code windows (syntax highlighting)

## Integration with Backend

### API Endpoints Expected

```
POST /api/auth/signup
- Body: { name, email, password, company, newsletter }
- Response: { success, requiresVerification, redirectUrl }

POST /api/auth/login
- Body: { email, password, remember }
- Response: { success, token, redirectUrl }

GET /api/auth/oauth/:provider
- Query: { redirect }
- Redirects to OAuth provider

POST /api/auth/resend-verification
- Body: { email }
- Response: { success }

POST /api/auth/reset-password
- Body: { email }
- Response: { success }
```

### Authentication Flow

1. **Signup Flow**
   - User fills signup form
   - Client validates input
   - POST to /api/auth/signup
   - If requiresVerification: redirect to /verify-email
   - Else: redirect to /dashboard

2. **Login Flow**
   - User fills login form
   - Client validates input
   - POST to /api/auth/login
   - Store token (localStorage or sessionStorage)
   - Redirect to /dashboard

3. **OAuth Flow**
   - User clicks OAuth button
   - Redirect to /api/auth/oauth/:provider
   - Provider authentication
   - Callback to /auth/callback
   - Token stored, redirect to /dashboard

## Deployment

### Static Hosting
The landing page can be served as static files from:
- Nginx
- Apache
- CDN (CloudFront, Cloudflare)
- Static hosting services (Netlify, Vercel)

### Server Configuration
Ensure the backend API is accessible at `/api/*` endpoints through:
- Reverse proxy configuration
- CORS headers properly set
- SSL/TLS certificates installed

### Environment Variables
No environment variables needed for static files. Backend API URL can be configured in JavaScript if needed.

## Browser Support

- Chrome (latest 2 versions)
- Firefox (latest 2 versions)
- Safari (latest 2 versions)
- Edge (latest 2 versions)
- Mobile browsers (iOS Safari, Chrome Mobile)

## Future Enhancements

### Planned Features
- [ ] Add product screenshots/videos
- [ ] Implement pricing page
- [ ] Add customer testimonials section
- [ ] Create blog section
- [ ] Add documentation pages
- [ ] Implement live chat widget
- [ ] Add A/B testing framework
- [ ] Create case studies page
- [ ] Add comparison page (vs competitors)
- [ ] Implement progressive web app (PWA)

### Performance Optimizations
- [ ] Add image optimization
- [ ] Implement lazy loading
- [ ] Add service worker for offline support
- [ ] Optimize font loading
- [ ] Implement critical CSS
- [ ] Add resource hints (preload, prefetch)

### Analytics Integration
- [ ] Google Analytics 4
- [ ] Mixpanel for user behavior
- [ ] Hotjar for heatmaps
- [ ] Conversion tracking
- [ ] A/B testing tools

## Testing

### Manual Testing Checklist
- [ ] All links work correctly
- [ ] Forms validate properly
- [ ] Mobile responsive on all devices
- [ ] Cross-browser compatibility
- [ ] Accessibility with screen readers
- [ ] Performance (Lighthouse score > 90)
- [ ] SEO optimization (meta tags, structure)

### Automated Testing
Consider adding:
- E2E tests with Playwright/Cypress
- Visual regression tests
- Performance monitoring
- Accessibility audits

## Security Considerations

### Implemented
- HTTPS only (enforce in production)
- CSRF protection (backend)
- XSS prevention (input sanitization)
- Secure password requirements
- OAuth 2.0 flows
- JWT token management

### Best Practices
- No sensitive data in client-side code
- Secure cookie flags (HttpOnly, Secure, SameSite)
- Content Security Policy headers
- Rate limiting on API endpoints
- Input validation on both client and server

## Maintenance

### Regular Updates
- Keep dependencies updated
- Monitor performance metrics
- Review analytics data
- Update content based on user feedback
- A/B test different variations
- Optimize conversion funnel

### Content Updates
- Update feature descriptions as product evolves
- Add new integration showcases
- Update testimonials and case studies
- Refresh screenshots and demos
- Update pricing information

## Support

For questions or issues with the landing page:
1. Check this README
2. Review the code comments
3. Contact the development team
4. Submit issues to the project repository

## License

Copyright © 2024 MCP Context Engine. All rights reserved.


## Authentication System

### Overview
Complete authentication system with email/password and OAuth support. See [AUTH_README.md](./AUTH_README.md) for detailed documentation.

### Authentication Pages

#### Login Page (`/login`)
- Email/password authentication
- OAuth login (GitHub, Google, Slack)
- "Remember me" functionality
- Password reset link
- Responsive design matching landing page

#### Signup Page (`/signup`)
- Account creation with email/password
- OAuth signup options
- Password strength indicator
- Terms of Service acceptance
- Email verification flow

#### Password Reset (`/password-reset`)
- Request password reset link
- Email delivery confirmation
- Resend link option
- Security best practices

#### Password Reset Confirmation (`/password-reset-confirm`)
- Set new password with token
- Password strength validation
- Confirm password matching
- Success/error states

#### Email Verification (`/verify-email`)
- Token-based email verification
- Resend verification email
- Pending/success/error states
- Automatic verification on token

### Authentication Module (`auth.js`)

#### AuthManager Class
Handles all authentication operations:
- Email/password validation
- Session management (JWT tokens)
- API calls (signup, login, logout, password reset, email verification)
- OAuth flows with CSRF protection
- Token storage in localStorage

#### AuthUI Class
Provides UI helper functions:
- Error/success message display
- Loading state management
- Password strength indicator
- Field-level validation feedback

### Security Features
- **Password Requirements**: 8+ characters, uppercase, lowercase, number
- **CSRF Protection**: State parameter for OAuth flows
- **JWT Tokens**: Secure token-based authentication
- **Email Verification**: Required for new accounts
- **Password Reset**: Secure token-based reset flow
- **Input Validation**: Client-side validation before API calls

### Backend API Endpoints

The authentication system requires these backend endpoints:

```
POST /api/auth/signup
POST /api/auth/login
POST /api/auth/logout
POST /api/auth/password-reset/request
POST /api/auth/password-reset/confirm
POST /api/auth/verify-email
POST /api/auth/verify-email/resend
GET  /api/auth/oauth/{provider}
POST /api/auth/oauth/callback
```

See [AUTH_README.md](./AUTH_README.md) for detailed API specifications and request/response formats.

### Usage Example

```javascript
// Initialize auth manager
const authManager = new AuthManager();

// Login
const result = await authManager.login('user@example.com', 'password');
if (result.success) {
  window.location.href = '/dashboard';
}

// Check authentication
if (authManager.isAuthenticated()) {
  const user = authManager.getUserData();
  console.log('Logged in as:', user.email);
}

// OAuth login
authManager.initiateOAuthFlow('github');
```

### Styling
Authentication pages use `auth.css` which provides:
- Consistent design with landing page
- Responsive layouts
- Form validation styles
- Loading states
- Success/error states
- OAuth button styling
- Password strength indicator

### Testing Authentication
1. Test email/password signup and login
2. Verify OAuth flows (GitHub, Google, Slack)
3. Test password reset flow
4. Verify email verification flow
5. Test error handling
6. Check responsive design
7. Verify accessibility features
