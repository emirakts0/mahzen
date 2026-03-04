import type { ApiError } from "@/types/api"

const BASE_URL = "/v1"

// ─────────────────────────────────────────────
// Token storage helpers
// ─────────────────────────────────────────────

const ACCESS_TOKEN_KEY = "mahzen_access_token"
const REFRESH_TOKEN_KEY = "mahzen_refresh_token"

export const tokenStorage = {
  getAccess: (): string | null => localStorage.getItem(ACCESS_TOKEN_KEY),
  getRefresh: (): string | null => localStorage.getItem(REFRESH_TOKEN_KEY),
  setTokens: (access: string, refresh: string): void => {
    localStorage.setItem(ACCESS_TOKEN_KEY, access)
    localStorage.setItem(REFRESH_TOKEN_KEY, refresh)
  },
  clearTokens: (): void => {
    localStorage.removeItem(ACCESS_TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
  },
}

// ─────────────────────────────────────────────
// Custom error class
// ─────────────────────────────────────────────

export class ApiRequestError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
    this.name = "ApiRequestError"
  }
}

// ─────────────────────────────────────────────
// Refresh token (used internally to avoid circular deps)
// ─────────────────────────────────────────────

let isRefreshing = false
let refreshSubscribers: Array<(token: string) => void> = []

function subscribeToRefresh(cb: (token: string) => void) {
  refreshSubscribers.push(cb)
}

function notifyRefreshSubscribers(newToken: string) {
  refreshSubscribers.forEach((cb) => cb(newToken))
  refreshSubscribers = []
}

async function refreshAccessToken(): Promise<string> {
  const refreshToken = tokenStorage.getRefresh()
  if (!refreshToken) {
    throw new ApiRequestError(401, "No refresh token available")
  }

  const res = await fetch(`${BASE_URL}/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refreshToken }),
  })

  if (!res.ok) {
    tokenStorage.clearTokens()
    throw new ApiRequestError(401, "Session expired, please log in again")
  }

  const data = (await res.json()) as { access_token: string; refresh_token: string }
  tokenStorage.setTokens(data.access_token, data.refresh_token)
  return data.access_token
}

// ─────────────────────────────────────────────
// Core fetch wrapper
// ─────────────────────────────────────────────

interface FetchOptions extends RequestInit {
  skipAuth?: boolean
}

export async function apiFetch<T>(path: string, options: FetchOptions = {}): Promise<T> {
  const { skipAuth = false, ...init } = options

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init.headers as Record<string, string>),
  }

  const accessToken = tokenStorage.getAccess()
  if (!skipAuth && accessToken) {
    headers["Authorization"] = `Bearer ${accessToken}`
  }

  const response = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers,
  })

  // Token expired → try refresh once
  if (response.status === 401 && !skipAuth && tokenStorage.getRefresh()) {
    if (!isRefreshing) {
      isRefreshing = true
      try {
        const newToken = await refreshAccessToken()
        isRefreshing = false
        notifyRefreshSubscribers(newToken)
      } catch (err) {
        isRefreshing = false
        throw err
      }
    }

    // Wait for refresh and retry
    const newToken = await new Promise<string>((resolve) => {
      subscribeToRefresh(resolve)
    })

    const retryResponse = await fetch(`${BASE_URL}${path}`, {
      ...init,
      headers: {
        ...headers,
        Authorization: `Bearer ${newToken}`,
      },
    })

    if (!retryResponse.ok) {
      const errBody = (await retryResponse.json().catch(() => ({ error: "Unknown error" }))) as ApiError
      throw new ApiRequestError(retryResponse.status, errBody.error)
    }

    return retryResponse.json() as Promise<T>
  }

  if (!response.ok) {
    const errBody = (await response.json().catch(() => ({ error: "Unknown error" }))) as ApiError
    throw new ApiRequestError(response.status, errBody.error)
  }

  // 200 with empty body (e.g. logout, delete)
  const text = await response.text()
  if (!text || text === "{}") return {} as T

  return JSON.parse(text) as T
}

// ─────────────────────────────────────────────
// Convenience helpers
// ─────────────────────────────────────────────

export const apiGet = <T>(path: string, options?: FetchOptions) =>
  apiFetch<T>(path, { method: "GET", ...options })

export const apiPost = <T>(path: string, body?: unknown, options?: FetchOptions) =>
  apiFetch<T>(path, { method: "POST", body: JSON.stringify(body), ...options })

export const apiPut = <T>(path: string, body?: unknown, options?: FetchOptions) =>
  apiFetch<T>(path, { method: "PUT", body: JSON.stringify(body), ...options })

export const apiDelete = <T>(path: string, options?: FetchOptions) =>
  apiFetch<T>(path, { method: "DELETE", ...options })
