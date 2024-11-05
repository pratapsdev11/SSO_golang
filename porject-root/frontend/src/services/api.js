import axios from 'axios';
import * as msal from '@azure/msal-browser';

// API Configuration
const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api';

// MSAL Configuration
const msalConfig = {
    auth: {
        clientId: process.env.REACT_APP_MICROSOFT_CLIENT_ID,
        authority: `https://login.microsoftonline.com/${process.env.REACT_APP_MICROSOFT_TENANT_ID}`,
        redirectUri: window.location.origin,
    },
    cache: {
        cacheLocation: 'sessionStorage',
        storeAuthStateInCookie: false,
    },
};

// MSAL Instance
const msalInstance = new msal.PublicClientApplication(msalConfig);

// Create axios instance with default config
const axiosInstance = axios.create({
    baseURL: API_URL,
    headers: {
        'Content-Type': 'application/json',
    },
});

// Add authorization header to requests
axiosInstance.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

// Add response interceptor for handling token expiration
axiosInstance.interceptors.response.use(
    (response) => response,
    async (error) => {
        const originalRequest = error.config;
        
        // If error is 401 and we haven't tried to refresh the token yet
        if (error.response.status === 401 && !originalRequest._retry) {
            originalRequest._retry = true;
            
            try {
                // Attempt to get a new token
                const msalResponse = await msalInstance.acquireTokenSilent({
                    scopes: ['user.read'],
                    account: msalInstance.getAllAccounts()[0]
                });
                
                // Exchange Microsoft token for our API token
                const response = await axiosInstance.post('/auth/microsoft', {
                    token: msalResponse.accessToken
                });
                
                const { token } = response.data;
                localStorage.setItem('token', token);
                
                // Retry the original request with new token
                originalRequest.headers.Authorization = `Bearer ${token}`;
                return axiosInstance(originalRequest);
            } catch (refreshError) {
                // If refresh fails, logout user
                localStorage.removeItem('token');
                window.location.href = '/signin';
                return Promise.reject(refreshError);
            }
        }
        return Promise.reject(error);
    }
);

const api = {
    // Auth endpoints
    auth: {
        async loginWithMicrosoft() {
            try {
                const msalResponse = await msalInstance.loginPopup({
                    scopes: ['user.read'],
                    prompt: 'select_account'
                });

                const response = await axiosInstance.post('/auth/microsoft', {
                    token: msalResponse.accessToken
                });

                const { token } = response.data;
                localStorage.setItem('token', token);
                return response.data;
            } catch (error) {
                console.error('Microsoft login failed:', error);
                throw error;
            }
        },

        async validateToken() {
            return axiosInstance.post('/auth/validate');
        },

        logout() {
            localStorage.removeItem('token');
            msalInstance.logout();
        }
    },

    // User endpoints
    user: {
        async getCurrentUser() {
            const response = await axiosInstance.get('/user/me');
            return response.data;
        },

        async updateProfile(userData) {
            const response = await axiosInstance.put('/user/profile', userData);
            return response.data;
        }
    },

    // Session endpoints
    session: {
        async create() {
            const response = await axiosInstance.post('/session/create');
            return response.data;
        },

        async delete(sessionId) {
            await axiosInstance.delete(`/session/${sessionId}`);
        }
    },

    // Utility methods
    getMsalInstance() {
        return msalInstance;
    },

    setAuthToken(token) {
        localStorage.setItem('token', token);
    },

    getAuthToken() {
        return localStorage.getItem('token');
    },

    clearAuthToken() {
        localStorage.removeItem('token');
    }
};

export default api;