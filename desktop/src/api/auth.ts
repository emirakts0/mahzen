import { invoke } from "@tauri-apps/api/core";
import { apiGet, tokenStorage } from "./client";
import type { UserResponse } from "@/types/api";

interface ApiResponse {
  status: number;
  headers: Record<string, string>;
  body: string;
}

export async function connectWithToken(serverUrl: string, accessToken: string): Promise<void> {
  tokenStorage.setAccessToken(accessToken, serverUrl);

  const baseUrl = serverUrl.replace(/\/$/, "");
  const response = await invoke<ApiResponse>("api_request", {
    request: {
      url: `${baseUrl}/v1/users/me`,
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${accessToken}`,
      },
    },
  });

  if (response.status !== 200) {
    tokenStorage.clearTokens();
    throw new Error("Invalid token or server unreachable");
  }
}

export async function logout(): Promise<void> {
  tokenStorage.clearTokens();
}

export async function getCurrentUser(): Promise<UserResponse> {
  return apiGet<UserResponse>("/v1/users/me");
}
