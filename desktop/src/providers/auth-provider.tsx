import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from "react";
import { tokenStorage, getBaseUrl, setBaseUrl } from "@/api/client";
import { connectWithToken, logout as apiLogout, getCurrentUser } from "@/api/auth";
import type { User } from "@/types/api";

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  connect: (serverUrl: string, accessToken: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const isAuthenticated = user !== null;

  useEffect(() => {
    const restore = async () => {
      const token = tokenStorage.getAccess();
      const baseUrl = await getBaseUrl();
      if (!token || !baseUrl) {
        setIsLoading(false);
        return;
      }
      try {
        const res = await getCurrentUser();
        setUser(res.user);
      } catch {
        tokenStorage.clearTokens();
        setUser(null);
      } finally {
        setIsLoading(false);
      }
    };

    restore();
  }, []);

  const connect = useCallback(async (serverUrl: string, accessToken: string) => {
    await setBaseUrl(serverUrl);
    await connectWithToken(serverUrl, accessToken);
    const res = await getCurrentUser();
    setUser(res.user);
  }, []);

  const logout = useCallback(async () => {
    await apiLogout();
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ user, isAuthenticated, isLoading, connect, logout }}>
      {children}
    </AuthContext.Provider>
  );
}
