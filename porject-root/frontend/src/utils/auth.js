import * as msal from '@azure/msal-browser';
import api from '../services/api';

// MSAL Configuration
const msalConfig = {
    auth: {
        clientId: process.env.REACT_APP_MICROSOFT_CLIENT_ID,
        authority: `https://login.microsoftonline.com/${process.env.REACT_APP_MICROSOFT_TENANT_ID}`,
        redirectUri: window.location.origin,
        navigateToLoginRequestUrl: true,
    },
    cache: {
        cacheLocation: 'sessionStorage',
        storeAuthStateInCookie: false,
    },
    system: {
        loggerOptions: {
            logLevel: msal.LogLevel.Error,
            piiLoggingEnabled: false
        }
    }
};

// Authentication scopes
const loginScopes = {
    scopes: ['user.read', 'profile', 'email', 'openid']
};

// MSAL instance
const msalInstance = new msal.PublicClientApplication(msalConfig);

const auth = {
    // Initialize authentication
    async initialize() {
        try {
            // Handle redirect promise
            await msalInstance.handleRedirectPromise();
            
            // Check if user is already signed in
            const accounts = msalInstance.getAllAccounts();
            if (accounts.length > 0) {
                msalInstance.setActiveAccount(accounts[0]);
                return true;
            }
            return false;
        } catch (error) {
            console.error('Auth initialization failed:', error);
            return false;
        }
    },

    // Handle Microsoft login
    async loginWithMicrosoft() {
        try {
            const loginResponse = await msalInstance.loginPopup(loginScopes);
            if (loginResponse) {
                // Exchange Microsoft token for our API token
                const apiResponse = await api.auth.loginWithMicrosoft();
                return {
                    success: true,
                    data: apiResponse
                };
            }
        } catch (error) {
            console.error('Microsoft login failed:', error);
            return {
                success: false,
                error: error.message
            };
        }
    },

    // Silent token acquisition
    async acquireToken() {
        try {
            const account = msalInstance.getActiveAccount();
            if (!account) {
                throw new Error('No active account');
            }

            const response = await msalInstance.acquireTokenSilent({
                ...loginScopes,
                account
            });
            return response.accessToken;
        } catch (error) {
            // If silent token acquisition fails, try interactive
            if (error instanceof msal.InteractionRequiredAuthError) {
                try {
                    const response = await msalInstance.acquireTokenPopup(loginScopes);
                    return response.accessToken;
                } catch (interactiveError) {
                    console.error('Interactive token acquisition failed:', interactiveError);
                    throw interactiveError;
                }
            }
            throw error;
        }
    },

    // Handle logout
    async logout() {
        try {
            // Clear local storage
            localStorage.removeItem('token');
            sessionStorage.clear();

            // Clear MSAL cache and accounts
            const account = msalInstance.getActiveAccount();
            if (account) {
                await msalInstance.logoutPopup({
                    account,
                    postLogoutRedirectUri: window.location.origin
                });
            }

            // Call API logout endpoint if needed
            await api.auth.logout();

            return true;
        } catch (error) {
            console.error('Logout failed:', error);
            return false;
        }
    },

    // Token management
    async refreshToken() {
        try {
            const token = await this.acquireToken();
            const response = await api.auth.refreshToken(token);
            if (response.data?.token) {
                localStorage.setItem('token', response.data.token);
                return response.data.token;
            }
            throw new Error('No token received');
        } catch (error) {
            console.error('Token refresh failed:', error);
            throw error;
        }
    },

    // Session management
    isAuthenticated() {
        return !!localStorage.getItem('token');
    },

    getToken() {
        return localStorage.getItem('token');
    },

    // User account management
    getCurrentAccount() {
        return msalInstance.getActiveAccount();
    },

    // Helper functions
    parseJwt(token) {
        try {
            const base64Url = token.split('.')[1];
            const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
            const jsonPayload = decodeURIComponent(atob(base64).split('').map(c => {
                return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
            }).join(''));
            return JSON.parse(jsonPayload);
        } catch (error) {
            console.error('JWT parsing failed:', error);
            return null;
        }
    },

    isTokenExpired(token) {
        if (!token) return true;
        const decodedToken = this.parseJwt(token);
        if (!decodedToken) return true;
        const currentTime = Date.now() / 1000;
        return decodedToken.exp < currentTime;
    },

    // Error handling
    handleAuthError(error) {
        // Common authentication errors
        const errorMessages = {
            'interaction_in_progress': 'Authentication is already in progress.',
            'login_required': 'Please sign in to continue.',
            'interaction_required': 'Please complete the authentication.',
            'no_account_error': 'No account is currently signed in.',
            'user_cancelled': 'Authentication was cancelled.',
            'token_expired': 'Your session has expired. Please sign in again.'
        };

        // Get user-friendly error message
        const errorMessage = errorMessages[error.errorCode] || error.message || 'An authentication error occurred.';

        return {
            code: error.errorCode,
            message: errorMessage,
            original: error
        };
    }
};

export default auth;