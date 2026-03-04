import type {
  AuthTokens,
  LoginRequest,
  RegisterRequest,
  LogoutRequest,
  RefreshRequest,
  UserResponse,
} from "@/types/api"
import { apiGet, apiPost, tokenStorage } from "./client"

export async function register(data: RegisterRequest): Promise<AuthTokens> {
  const tokens = await apiPost<AuthTokens>("/auth/register", data, { skipAuth: true })
  tokenStorage.setTokens(tokens.access_token, tokens.refresh_token)
  return tokens
}

export async function login(data: LoginRequest): Promise<AuthTokens> {
  const tokens = await apiPost<AuthTokens>("/auth/login", data, { skipAuth: true })
  tokenStorage.setTokens(tokens.access_token, tokens.refresh_token)
  return tokens
}

export async function logout(): Promise<void> {
  const refreshToken = tokenStorage.getRefresh()
  if (refreshToken) {
    const body: LogoutRequest = { refresh_token: refreshToken }
    await apiPost<unknown>("/auth/logout", body).catch(() => {
      // Best-effort: clear tokens even if request fails
    })
  }
  tokenStorage.clearTokens()
}

export async function refreshToken(): Promise<AuthTokens> {
  const refreshTk = tokenStorage.getRefresh()
  if (!refreshTk) throw new Error("No refresh token")
  const body: RefreshRequest = { refresh_token: refreshTk }
  const tokens = await apiPost<AuthTokens>("/auth/refresh", body, { skipAuth: true })
  tokenStorage.setTokens(tokens.access_token, tokens.refresh_token)
  return tokens
}

export async function getCurrentUser(): Promise<UserResponse> {
  return apiGet<UserResponse>("/users/me")
}
