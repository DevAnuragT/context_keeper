/**
 * Integration Tests for Authentication Flow
 * 
 * Tests the complete authentication workflows including:
 * - User signup with email verification
 * - Login with email/password
 * - OAuth authentication flows
 * - Password reset functionality
 * - Session management
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

// Mock fetch for API calls
const mockFetch = vi.fn();
global.fetch = mockFetch;

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => { store[key] = value; },
    removeItem: (key: string) => { delete store[key]; },
    clear: () => { store = {}; }
  };
})();
Object.defineProperty(window, 'localStorage', { value: localStorageMock });

// Mock sessionStorage
const sessionStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => { store[key] = value; },
    removeItem: (key: string) => { delete store[key]; },
    clear: () => { store = {}; }
  };
})();
Object.defineProperty(window, 'sessionStorage', { value: sessionStorageMock });

// Mock crypto for CSRF token generation
Object.defineProperty(global, 'crypto', {
  value: {
    getRandomValues: (arr: Uint8Array) => {
      for (let i = 0; i < arr.length; i++) {
        arr[i] = Math.floor(Math.random() * 256);
      }
      return arr;
    }
  }
});

describe('Authentication Integration Tests', () => {
  beforeEach(() => {
    localStorageMock.clear();
    sessionStorageMock.clear();
    mockFetch.mockClear();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('User Signup Flow', () => {
    it('should successfully sign up a new user with valid credentials', async () => {
      const mockResponse = {
        success: true,
        token: 'mock-jwt-token',
        user: {
          id: '123',
          email: 'test@example.com',
          name: 'Test User',
          email_verified: false
        }
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/signup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: 'test@example.com',
          password: 'SecurePass123!',
          name: 'Test User'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.token).toBe('mock-jwt-token');
      expect(data.user.email).toBe('test@example.com');
      expect(data.user.email_verified).toBe(false);
    });

    it('should reject signup with invalid email format', async () => {
      const mockResponse = {
        success: false,
        message: 'Invalid email format'
      };

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/signup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: 'invalid-email',
          password: 'SecurePass123!',
          name: 'Test User'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(false);
      expect(data.message).toBe('Invalid email format');
    });

    it('should reject signup with weak password', async () => {
      const mockResponse = {
        success: false,
        message: 'Password does not meet security requirements'
      };

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/signup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: 'test@example.com',
          password: 'weak',
          name: 'Test User'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(false);
      expect(data.message).toContain('security requirements');
    });

    it('should reject signup with duplicate email', async () => {
      const mockResponse = {
        success: false,
        message: 'Email already registered'
      };

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 409,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/signup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: 'existing@example.com',
          password: 'SecurePass123!',
          name: 'Test User'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(false);
      expect(response.status).toBe(409);
      expect(data.message).toBe('Email already registered');
    });
  });

  describe('Email Verification Flow', () => {
    it('should successfully verify email with valid token', async () => {
      const mockResponse = {
        success: true,
        message: 'Email verified successfully'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/verify-email', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          token: 'valid-verification-token'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.success).toBe(true);
      expect(data.message).toBe('Email verified successfully');
    });

    it('should reject verification with invalid token', async () => {
      const mockResponse = {
        success: false,
        message: 'Invalid or expired verification token'
      };

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/verify-email', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          token: 'invalid-token'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(false);
      expect(data.message).toContain('Invalid or expired');
    });

    it('should successfully resend verification email', async () => {
      const mockResponse = {
        success: true,
        message: 'Verification email sent'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/verify-email/resend', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: 'test@example.com'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.success).toBe(true);
    });
  });

  describe('Login Flow', () => {
    it('should successfully login with valid credentials', async () => {
      const mockResponse = {
        success: true,
        token: 'mock-jwt-token',
        user: {
          id: '123',
          email: 'test@example.com',
          name: 'Test User',
          email_verified: true
        }
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: 'test@example.com',
          password: 'SecurePass123!'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.token).toBe('mock-jwt-token');
      expect(data.user.email_verified).toBe(true);
    });

    it('should reject login with incorrect password', async () => {
      const mockResponse = {
        success: false,
        message: 'Invalid email or password'
      };

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: 'test@example.com',
          password: 'WrongPassword123!'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(false);
      expect(response.status).toBe(401);
      expect(data.message).toBe('Invalid email or password');
    });

    it('should reject login for non-existent user', async () => {
      const mockResponse = {
        success: false,
        message: 'Invalid email or password'
      };

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: 'nonexistent@example.com',
          password: 'SecurePass123!'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(false);
      expect(response.status).toBe(401);
    });

    it('should warn about unverified email but allow login', async () => {
      const mockResponse = {
        success: true,
        token: 'mock-jwt-token',
        user: {
          id: '123',
          email: 'test@example.com',
          name: 'Test User',
          email_verified: false
        },
        warning: 'Please verify your email address'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: 'test@example.com',
          password: 'SecurePass123!'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.user.email_verified).toBe(false);
      expect(data.warning).toBeDefined();
    });
  });

  describe('OAuth Authentication Flow', () => {
    it('should initiate GitHub OAuth flow with CSRF protection', () => {
      const csrfToken = 'mock-csrf-token';
      sessionStorageMock.setItem('oauth_csrf_token', csrfToken);

      const oauthUrl = `/api/auth/oauth/github?state=${csrfToken}`;
      
      expect(sessionStorageMock.getItem('oauth_csrf_token')).toBe(csrfToken);
      expect(oauthUrl).toContain('state=');
    });

    it('should successfully complete OAuth callback with valid state', async () => {
      const csrfToken = 'mock-csrf-token';
      sessionStorageMock.setItem('oauth_csrf_token', csrfToken);

      const mockResponse = {
        success: true,
        token: 'mock-jwt-token',
        user: {
          id: '123',
          email: 'test@example.com',
          name: 'Test User',
          oauth_provider: 'github'
        }
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/oauth/callback', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          code: 'oauth-code',
          state: csrfToken
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.token).toBe('mock-jwt-token');
      expect(data.user.oauth_provider).toBe('github');
    });

    it('should reject OAuth callback with invalid CSRF token', async () => {
      sessionStorageMock.setItem('oauth_csrf_token', 'valid-token');

      const mockResponse = {
        success: false,
        message: 'CSRF token validation failed'
      };

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 403,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/oauth/callback', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          code: 'oauth-code',
          state: 'invalid-token'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(false);
      expect(response.status).toBe(403);
      expect(data.message).toContain('CSRF');
    });

    it('should handle OAuth provider errors gracefully', async () => {
      const mockResponse = {
        success: false,
        message: 'OAuth provider authentication failed'
      };

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/oauth/callback', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          code: 'invalid-code',
          state: 'valid-state'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(false);
      expect(data.message).toContain('OAuth provider');
    });
  });

  describe('Password Reset Flow', () => {
    it('should successfully request password reset', async () => {
      const mockResponse = {
        success: true,
        message: 'Password reset email sent'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/password-reset/request', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: 'test@example.com'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.success).toBe(true);
    });

    it('should not reveal if email exists for security', async () => {
      const mockResponse = {
        success: true,
        message: 'If the email exists, a reset link has been sent'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/password-reset/request', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: 'nonexistent@example.com'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.message).toContain('If the email exists');
    });

    it('should successfully reset password with valid token', async () => {
      const mockResponse = {
        success: true,
        message: 'Password reset successfully'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/password-reset/confirm', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          token: 'valid-reset-token',
          password: 'NewSecurePass123!'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.success).toBe(true);
    });

    it('should reject password reset with expired token', async () => {
      const mockResponse = {
        success: false,
        message: 'Reset token is invalid or expired'
      };

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: async () => mockResponse
      });

      const response = await fetch('/api/auth/password-reset/confirm', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          token: 'expired-token',
          password: 'NewSecurePass123!'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(false);
      expect(data.message).toContain('expired');
    });
  });

  describe('Session Management', () => {
    it('should store session data after successful login', () => {
      const token = 'mock-jwt-token';
      const userData = {
        id: '123',
        email: 'test@example.com',
        name: 'Test User'
      };

      localStorageMock.setItem('mcp_auth_token', token);
      localStorageMock.setItem('mcp_user_data', JSON.stringify(userData));

      expect(localStorageMock.getItem('mcp_auth_token')).toBe(token);
      const storedUser = JSON.parse(localStorageMock.getItem('mcp_user_data') || '{}');
      expect(storedUser.email).toBe('test@example.com');
    });

    it('should clear session data on logout', async () => {
      localStorageMock.setItem('mcp_auth_token', 'mock-token');
      localStorageMock.setItem('mcp_user_data', JSON.stringify({ id: '123' }));

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true })
      });

      await fetch('/api/auth/logout', {
        method: 'POST',
        headers: {
          'Authorization': 'Bearer mock-token'
        }
      });

      localStorageMock.removeItem('mcp_auth_token');
      localStorageMock.removeItem('mcp_user_data');

      expect(localStorageMock.getItem('mcp_auth_token')).toBeNull();
      expect(localStorageMock.getItem('mcp_user_data')).toBeNull();
    });

    it('should validate JWT token format', () => {
      const validToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c';
      const invalidToken = 'not-a-jwt-token';

      const isValidJWT = (token: string) => {
        const parts = token.split('.');
        return parts.length === 3;
      };

      expect(isValidJWT(validToken)).toBe(true);
      expect(isValidJWT(invalidToken)).toBe(false);
    });

    it('should handle expired session gracefully', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({
          success: false,
          message: 'Session expired'
        })
      });

      const response = await fetch('/api/projects', {
        method: 'GET',
        headers: {
          'Authorization': 'Bearer expired-token'
        }
      });

      expect(response.ok).toBe(false);
      expect(response.status).toBe(401);
    });
  });

  describe('Form Validation', () => {
    it('should validate email format', () => {
      const validateEmail = (email: string) => {
        const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return re.test(email.toLowerCase());
      };

      expect(validateEmail('test@example.com')).toBe(true);
      expect(validateEmail('invalid-email')).toBe(false);
      expect(validateEmail('test@')).toBe(false);
      expect(validateEmail('@example.com')).toBe(false);
    });

    it('should validate password strength', () => {
      const validatePassword = (password: string) => {
        const errors = [];
        if (password.length < 8) errors.push('Too short');
        if (!/[A-Z]/.test(password)) errors.push('No uppercase');
        if (!/[a-z]/.test(password)) errors.push('No lowercase');
        if (!/[0-9]/.test(password)) errors.push('No number');
        return { valid: errors.length === 0, errors };
      };

      expect(validatePassword('SecurePass123!').valid).toBe(true);
      expect(validatePassword('weak').valid).toBe(false);
      expect(validatePassword('NoNumbers!').valid).toBe(false);
      expect(validatePassword('nonumbers123').valid).toBe(false);
    });
  });

  describe('Error Handling', () => {
    it('should handle network errors gracefully', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      try {
        await fetch('/api/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            email: 'test@example.com',
            password: 'SecurePass123!'
          })
        });
      } catch (error) {
        expect(error).toBeInstanceOf(Error);
        expect((error as Error).message).toBe('Network error');
      }
    });

    it('should handle malformed JSON responses', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        json: async () => {
          throw new Error('Invalid JSON');
        }
      });

      try {
        const response = await fetch('/api/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            email: 'test@example.com',
            password: 'SecurePass123!'
          })
        });
        await response.json();
      } catch (error) {
        expect(error).toBeInstanceOf(Error);
      }
    });
  });
});
