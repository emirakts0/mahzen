import { apiGet, apiPost, tokenStorage } from "./client";
import type { AuthTokens, LoginRequest, RegisterRequest, UserResponse } from "@/types/api";

export async function login(data: LoginRequest): Promise<AuthTokens> {
  const tokens = await apiPost<AuthTokens>("/v1/auth/login", data, { skipAuth: true });
  tokenStorage.setTokens(tokens.access_token, tokens.refresh_token);
  return tokens;
}

export async function register(data: RegisterRequest): Promise<AuthTokens> {
  const tokens = await apiPost<AuthTokens>("/v1/auth/register", data, { skipAuth: true });
  tokenStorage.setTokens(tokens.access_token, tokens.refresh_token);
  return tokens;
}

export async function logout(): Promise<void> {
   const refreshToken = tokenStorage.getRefresh()
   if (refreshToken) {
     await apiPost("/v1/auth/logout", { refresh_token: refreshToken }).catch(() => {})
   }
   tokenStorage.clearTokens()
 }

export async function getCurrentUser(): Promise<UserResponse> {
  return apiGet<UserResponse>("/v1/users/me");
}
