import React, { createContext, useCallback, useContext, useEffect, useState } from 'react';
import Auth0 from 'react-native-auth0';
import * as SecureStore from 'expo-secure-store';

const AUTH0_DOMAIN = process.env.EXPO_PUBLIC_AUTH0_DOMAIN ?? '';
const AUTH0_CLIENT_ID = process.env.EXPO_PUBLIC_AUTH0_CLIENT_ID ?? '';
const AUTH0_AUDIENCE = process.env.EXPO_PUBLIC_AUTH0_AUDIENCE ?? '';

const auth0 = new Auth0({ domain: AUTH0_DOMAIN, clientId: AUTH0_CLIENT_ID });

const ACCESS_TOKEN_KEY = 'auth0_access_token';

interface AuthContextValue {
  accessToken: string | null;
  userID: string | null;
  isLoading: boolean;
  login: () => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue>({
  accessToken: null,
  userID: null,
  isLoading: true,
  login: async () => {},
  logout: async () => {},
});

function parseJwtSub(token: string): string | null {
  try {
    const payload = token.split('.')[1];
    const decoded = JSON.parse(atob(payload));
    return decoded.sub ?? null;
  } catch {
    return null;
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [userID, setUserID] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    SecureStore.getItemAsync(ACCESS_TOKEN_KEY).then((stored) => {
      if (stored) {
        setAccessToken(stored);
        setUserID(parseJwtSub(stored));
      }
      setIsLoading(false);
    });
  }, []);

  const login = useCallback(async () => {
    const credentials = await auth0.webAuth.authorize({
      scope: 'openid profile email',
      audience: AUTH0_AUDIENCE,
    });
    await SecureStore.setItemAsync(ACCESS_TOKEN_KEY, credentials.accessToken);
    setAccessToken(credentials.accessToken);
    setUserID(parseJwtSub(credentials.accessToken));
  }, []);

  const logout = useCallback(async () => {
    await auth0.webAuth.clearSession();
    await SecureStore.deleteItemAsync(ACCESS_TOKEN_KEY);
    setAccessToken(null);
    setUserID(null);
  }, []);

  return (
    <AuthContext.Provider value={{ accessToken, userID, isLoading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
