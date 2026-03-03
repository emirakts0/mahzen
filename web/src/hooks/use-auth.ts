import { useCallback, useEffect, useState } from "react"
import { api } from "@/lib/api"

interface AuthUser {
  email: string
}

// Token storage keys
const ACCESS_TOKEN_KEY = "mahzen_access_token"
const REFRESH_TOKEN_KEY = "mahzen_refresh_token"

// Get/set tokens in localStorage
function getAccessToken(): string | null {
  return localStorage.getItem(ACCESS_TOKEN_KEY)
}

function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

function setTokens(accessToken: string, refreshToken: string) {
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
}

function clearTokens() {
  localStorage.removeItem(ACCESS_TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
}

export function useAuth() {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Check if we have a valid session by calling /v1/users/me
  const checkSession = useCallback(async () => {
    const token = getAccessToken()
    if (!token) {
      setUser(null)
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      const data = await api.getCurrentUser()
      setUser({ email: data.user.email })
      setError(null)
    } catch {
      // Token might be expired, try to refresh
      const refreshed = await tryRefresh()
      if (!refreshed) {
        setUser(null)
        clearTokens()
      }
    } finally {
      setLoading(false)
    }
  }, [])

  const tryRefresh = async (): Promise<boolean> => {
    const refreshToken = getRefreshToken()
    if (!refreshToken) return false

    try {
      const data = await api.refreshToken(refreshToken)
      setTokens(data.access_token, data.refresh_token)
      // Retry getting user info with new token
      const userData = await api.getCurrentUser()
      setUser({ email: userData.user.email })
      return true
    } catch {
      clearTokens()
      return false
    }
  }

  useEffect(() => {
    checkSession()
  }, [checkSession])

  const login = useCallback(
    async (email: string, password: string) => {
      setError(null)
      try {
        const data = await api.login(email, password)
        setTokens(data.access_token, data.refresh_token)
        // Fetch user info
        const userData = await api.getCurrentUser()
        setUser({ email: userData.user.email })
      } catch (err) {
        const message =
          err && typeof err === "object" && "message" in err
            ? String((err as { message: string }).message)
            : "Login failed"
        setError(message)
        throw err
      }
    },
    []
  )

  const register = useCallback(
    async (email: string, displayName: string, password: string) => {
      setError(null)
      try {
        const data = await api.register(email, displayName, password)
        setTokens(data.access_token, data.refresh_token)
        setUser({ email })
      } catch (err) {
        const message =
          err && typeof err === "object" && "message" in err
            ? String((err as { message: string }).message)
            : "Registration failed"
        setError(message)
        throw err
      }
    },
    []
  )

  const logout = useCallback(async () => {
    const refreshToken = getRefreshToken()
    if (refreshToken) {
      try {
        await api.logout(refreshToken)
      } catch {
        // Best-effort logout
      }
    }
    clearTokens()
    setUser(null)
  }, [])

  return {
    user,
    loading,
    error,
    isAuthenticated: !!user,
    login,
    register,
    logout,
    checkSession,
  }
}

// Export token getter for use by api.ts
export { getAccessToken, getRefreshToken, setTokens, clearTokens }
