// Authentication Module
// Handles form validation, API calls, and session management

class AuthManager {
    constructor() {
        this.apiBaseUrl = '/api/auth';
        this.tokenKey = 'mcp_auth_token';
        this.userKey = 'mcp_user_data';
    }

    // Email validation
    validateEmail(email) {
        const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return re.test(email.toLowerCase());
    }

    // Password validation
    validatePassword(password) {
        const errors = [];
        
        if (password.length < 8) {
            errors.push('Password must be at least 8 characters long');
        }
        if (!/[A-Z]/.test(password)) {
            errors.push('Password must contain at least one uppercase letter');
        }
        if (!/[a-z]/.test(password)) {
            errors.push('Password must contain at least one lowercase letter');
        }
        if (!/[0-9]/.test(password)) {
            errors.push('Password must contain at least one number');
        }
        
        return {
            valid: errors.length === 0,
            errors: errors
        };
    }

    // Password strength indicator
    getPasswordStrength(password) {
        let strength = 0;
        
        if (password.length >= 8) strength++;
        if (password.length >= 12) strength++;
        if (/[a-z]/.test(password) && /[A-Z]/.test(password)) strength++;
        if (/[0-9]/.test(password)) strength++;
        if (/[^A-Za-z0-9]/.test(password)) strength++;
        
        if (strength <= 2) return { level: 'weak', color: '#ef4444' };
        if (strength <= 3) return { level: 'medium', color: '#f59e0b' };
        return { level: 'strong', color: '#10b981' };
    }

    // Session management
    setSession(token, userData) {
        localStorage.setItem(this.tokenKey, token);
        localStorage.setItem(this.userKey, JSON.stringify(userData));
    }

    getToken() {
        return localStorage.getItem(this.tokenKey);
    }

    getUserData() {
        const data = localStorage.getItem(this.userKey);
        return data ? JSON.parse(data) : null;
    }

    clearSession() {
        localStorage.removeItem(this.tokenKey);
        localStorage.removeItem(this.userKey);
    }

    isAuthenticated() {
        return !!this.getToken();
    }

    // API calls
    async signup(email, password, name = '') {
        try {
            const response = await fetch(`${this.apiBaseUrl}/signup`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password, name }),
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || 'Signup failed');
            }

            if (data.token) {
                this.setSession(data.token, data.user);
            }

            return { success: true, data };
        } catch (error) {
            return { success: false, error: error.message };
        }
    }

    async login(email, password) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password }),
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || 'Login failed');
            }

            if (data.token) {
                this.setSession(data.token, data.user);
            }

            return { success: true, data };
        } catch (error) {
            return { success: false, error: error.message };
        }
    }

    async logout() {
        try {
            const token = this.getToken();
            if (token) {
                await fetch(`${this.apiBaseUrl}/logout`, {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${token}`,
                    },
                });
            }
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            this.clearSession();
        }
    }

    async requestPasswordReset(email) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/password-reset/request`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email }),
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || 'Password reset request failed');
            }

            return { success: true, data };
        } catch (error) {
            return { success: false, error: error.message };
        }
    }

    async resetPassword(token, newPassword) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/password-reset/confirm`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ token, password: newPassword }),
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || 'Password reset failed');
            }

            return { success: true, data };
        } catch (error) {
            return { success: false, error: error.message };
        }
    }

    async verifyEmail(token) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/verify-email`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ token }),
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || 'Email verification failed');
            }

            return { success: true, data };
        } catch (error) {
            return { success: false, error: error.message };
        }
    }

    async resendVerificationEmail(email) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/verify-email/resend`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email }),
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || 'Failed to resend verification email');
            }

            return { success: true, data };
        } catch (error) {
            return { success: false, error: error.message };
        }
    }

    // OAuth flows
    initiateOAuthFlow(provider) {
        // Generate CSRF token
        const csrfToken = this.generateCSRFToken();
        sessionStorage.setItem('oauth_csrf_token', csrfToken);
        
        // Redirect to OAuth provider
        window.location.href = `${this.apiBaseUrl}/oauth/${provider}?state=${csrfToken}`;
    }

    generateCSRFToken() {
        const array = new Uint8Array(32);
        crypto.getRandomValues(array);
        return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
    }

    validateCSRFToken(token) {
        const storedToken = sessionStorage.getItem('oauth_csrf_token');
        sessionStorage.removeItem('oauth_csrf_token');
        return token === storedToken;
    }

    // Handle OAuth callback
    async handleOAuthCallback() {
        const urlParams = new URLSearchParams(window.location.search);
        const code = urlParams.get('code');
        const state = urlParams.get('state');
        const error = urlParams.get('error');

        if (error) {
            return { success: false, error: error };
        }

        if (!code || !state) {
            return { success: false, error: 'Invalid OAuth callback' };
        }

        if (!this.validateCSRFToken(state)) {
            return { success: false, error: 'CSRF token validation failed' };
        }

        try {
            const response = await fetch(`${this.apiBaseUrl}/oauth/callback`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ code, state }),
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || 'OAuth authentication failed');
            }

            if (data.token) {
                this.setSession(data.token, data.user);
            }

            return { success: true, data };
        } catch (error) {
            return { success: false, error: error.message };
        }
    }
}

// UI Helper Functions
class AuthUI {
    static showError(formElement, message) {
        let errorEl = formElement.querySelector('.form-error');
        if (!errorEl) {
            errorEl = document.createElement('div');
            errorEl.className = 'form-error';
            formElement.insertBefore(errorEl, formElement.firstChild);
        }
        errorEl.textContent = message;
        errorEl.style.display = 'block';
        
        // Auto-hide after 5 seconds
        setTimeout(() => {
            errorEl.style.display = 'none';
        }, 5000);
    }

    static showSuccess(formElement, message) {
        let successEl = formElement.querySelector('.form-success');
        if (!successEl) {
            successEl = document.createElement('div');
            successEl.className = 'form-success';
            formElement.insertBefore(successEl, formElement.firstChild);
        }
        successEl.textContent = message;
        successEl.style.display = 'block';
        
        // Auto-hide after 5 seconds
        setTimeout(() => {
            successEl.style.display = 'none';
        }, 5000);
    }

    static clearMessages(formElement) {
        const errorEl = formElement.querySelector('.form-error');
        const successEl = formElement.querySelector('.form-success');
        if (errorEl) errorEl.style.display = 'none';
        if (successEl) successEl.style.display = 'none';
    }

    static setLoading(button, loading) {
        if (loading) {
            button.disabled = true;
            button.dataset.originalText = button.textContent;
            button.innerHTML = '<span class="spinner"></span> Loading...';
        } else {
            button.disabled = false;
            button.textContent = button.dataset.originalText || button.textContent;
        }
    }

    static updatePasswordStrength(password, strengthElement) {
        const auth = new AuthManager();
        const strength = auth.getPasswordStrength(password);
        
        strengthElement.textContent = strength.level;
        strengthElement.style.color = strength.color;
        
        const strengthBar = strengthElement.parentElement.querySelector('.strength-bar');
        if (strengthBar) {
            strengthBar.style.width = password.length === 0 ? '0%' : 
                strength.level === 'weak' ? '33%' :
                strength.level === 'medium' ? '66%' : '100%';
            strengthBar.style.backgroundColor = strength.color;
        }
    }

    static showFieldError(inputElement, message) {
        const errorEl = inputElement.parentElement.querySelector('.field-error') || 
            document.createElement('div');
        errorEl.className = 'field-error';
        errorEl.textContent = message;
        
        if (!inputElement.parentElement.querySelector('.field-error')) {
            inputElement.parentElement.appendChild(errorEl);
        }
        
        inputElement.classList.add('error');
    }

    static clearFieldError(inputElement) {
        const errorEl = inputElement.parentElement.querySelector('.field-error');
        if (errorEl) {
            errorEl.remove();
        }
        inputElement.classList.remove('error');
    }
}

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { AuthManager, AuthUI };
}
