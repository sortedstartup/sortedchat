import { atom, computed } from 'nanostores';

// Constants
export const JWT_TOKEN_KEY = 'sortedchat.jwt' as const;

// Types
interface JWT {
  sub: string;      // user id
  email: string;
  roles: string[];
  iss: string;      // issuer
  exp: number;      // expiration timestamp
  iat: number;      // issued at timestamp
}

interface AuthState {
  isLoggedIn: boolean;
  token: string | null;
  user: {
    id: string;
    email: string;
    roles: string[];
  } | null;
}

// Helper function to decode JWT payload (without verification)
function decodeJWTPayload(token: string): JWT | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) {
      return null;
    }
    
    const payload = parts[1];
    const decoded = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')));
    return decoded as JWT;
  } catch (error) {
    console.error('Failed to decode JWT payload:', error);
    return null;
  }
}

// Helper function to check if token is expired
function isTokenExpired(payload: JWT): boolean {
  const now = Math.floor(Date.now() / 1000);
  return payload.exp < now;
}

// Helper function to validate token
function validateToken(token: string | null): { isValid: boolean; jwt: JWT | null } {
  if (!token) {
    return { isValid: false, jwt: null };
  }

  const payload = decodeJWTPayload(token);
  if (!payload) {
    return { isValid: false, jwt: null };
  }

  if (isTokenExpired(payload)) {
    return { isValid: false, jwt: null };
  }

  return { isValid: true, jwt: payload };
}

// Helper function to get token from localStorage
function getStoredToken(): string | null {
  try {
    return localStorage.getItem(JWT_TOKEN_KEY);
  } catch (error) {
    console.error('Failed to get token from localStorage:', error);
    return null;
  }
}

// Initialize token atom with stored token
const initialToken = getStoredToken();
export const $token = atom<string | null>(initialToken);

// Computed store for authentication state
export const $auth = computed($token, (token) => {
  const { isValid, jwt: payload } = validateToken(token);
  
  if (!isValid || !payload) {
    return {
      isLoggedIn: false,
      token: null,
      user: null,
    } as AuthState;
  }

  return {
    isLoggedIn: true,
    token,
    user: {
      id: payload.sub,
      email: payload.email,
      roles: payload.roles,
    },
  } as AuthState;
});

// Actions
export const authActions = {
  // Set token and store in localStorage
  setToken(token: string): void {
    try {
      localStorage.setItem(JWT_TOKEN_KEY, token);
      $token.set(token);
    } catch (error) {
      console.error('Failed to store token:', error);
    }
  },

  // Clear token and remove from localStorage
  clearToken(): void {
    try {
      localStorage.removeItem(JWT_TOKEN_KEY);
      $token.set(null);
    } catch (error) {
      console.error('Failed to clear token:', error);
    }
  },

  // Check if current token is valid
  isTokenValid(): boolean {
    const token = $token.get();
    const { isValid } = validateToken(token);
    return isValid;
  },

  // Get time until token expires (in seconds)
  getTimeUntilExpiry(): number | null {
    const token = $token.get();
    const { isValid, jwt: payload } = validateToken(token);
    
    if (!isValid || !payload) {
      return null;
    }

    const now = Math.floor(Date.now() / 1000);
    return payload.exp - now;
  },

  // Refresh token validation (useful for periodic checks)
  refreshValidation(): void {
    const currentToken = $token.get();
    const { isValid } = validateToken(currentToken);
    
    if (!isValid && currentToken) {
      // Token is invalid, clear it
      authActions.clearToken();
    }
  },
};

// Auto-refresh validation periodically (every 30 seconds)
if (typeof window !== 'undefined') {
  setInterval(() => {
    authActions.refreshValidation();
  }, 30000);

  // Listen for storage changes from other tabs
  window.addEventListener('storage', (event) => {
    if (event.key === JWT_TOKEN_KEY) {
      $token.set(event.newValue);
    }
  });
}
