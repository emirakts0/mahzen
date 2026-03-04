import { useEffect, useState, useCallback } from "react"
import type { ReactNode } from "react"
import type { User } from "@/types/api"
import { getCurrentUser, login as apiLogin, logout as apiLogout, register as apiRegister } from "@/api/auth"
import type { LoginRequest, RegisterRequest } from "@/types/api"
import { tokenStorage } from "@/api/client"
import { AuthContext } from "@/context/auth-context"
import type { AuthContextValue } from "@/context/auth-context"

export type { AuthContextValue }

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  // On mount: try to restore session from stored token
  useEffect(() => {
    const restore = async () => {
      if (!tokenStorage.getAccess()) {
        setIsLoading(false)
        return
      }
      try {
        const res = await getCurrentUser()
        setUser(res.user)
      } catch {
        // Token invalid / expired and refresh failed inside client
        tokenStorage.clearTokens()
        setUser(null)
      } finally {
        setIsLoading(false)
      }
    }

    void restore()
  }, [])

  const login = useCallback(async (data: LoginRequest) => {
    await apiLogin(data)
    const res = await getCurrentUser()
    setUser(res.user)
  }, [])

  const register = useCallback(async (data: RegisterRequest) => {
    await apiRegister(data)
    const res = await getCurrentUser()
    setUser(res.user)
  }, [])

  const logout = useCallback(async () => {
    await apiLogout()
    setUser(null)
  }, [])

  const value: AuthContextValue = {
    user,
    isAuthenticated: user !== null,
    isLoading,
    login,
    register,
    logout,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
