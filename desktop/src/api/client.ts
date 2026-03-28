import { invoke } from "@tauri-apps/api/core";
import type { ApiError } from "@/types/api";

let cachedBaseUrl: string | null | undefined = undefined;

export async function getBaseUrl(): Promise<string | null> {
  if (cachedBaseUrl !== undefined) return cachedBaseUrl;

  try {
    const config = await invoke<{ backend_url: string }>("get_config");
    cachedBaseUrl = config.backend_url || null;
    return cachedBaseUrl;
  } catch {
    return null;
  }
}

export async function setBaseUrl(url: string): Promise<void> {
  await invoke("set_config", { config: { backend_url: url } });
  cachedBaseUrl = url;
}

export function clearBaseUrlCache() {
  cachedBaseUrl = undefined;
}

const ACCESS_TOKEN_KEY = "mahzen_access_token";

export const tokenStorage = {
  getAccess: (): string | null => localStorage.getItem(ACCESS_TOKEN_KEY),
  setAccessToken: (token: string, _serverUrl?: string): void => {
    localStorage.setItem(ACCESS_TOKEN_KEY, token);
  },
  clearTokens: (): void => {
    localStorage.removeItem(ACCESS_TOKEN_KEY);
  },
};

export class ApiRequestError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
    this.name = "ApiRequestError";
  }
}

interface ApiResponse {
  status: number;
  headers: Record<string, string>;
  body: string;
}

async function doRequest(
  url: string,
  method: string,
  headers: Record<string, string>,
  body?: string
): Promise<ApiResponse> {
  return invoke<ApiResponse>("api_request", {
    request: { url, method, headers, body },
  });
}

interface FetchOptions extends RequestInit {
  skipAuth?: boolean;
}

export async function apiFetch<T>(path: string, options: FetchOptions = {}): Promise<T> {
  const { skipAuth = false, ...init } = options;

  const baseUrl = await getBaseUrl();
  if (!baseUrl) {
    throw new ApiRequestError(400, "Backend URL not configured");
  }

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init.headers as Record<string, string>),
  };

  const accessToken = tokenStorage.getAccess();
  if (!skipAuth && accessToken) {
    headers["Authorization"] = `Bearer ${accessToken}`;
  }

  const response = await doRequest(
    `${baseUrl}${path}`,
    init.method || "GET",
    headers,
    init.body?.toString()
  );

  // On 401, opaque tokens can't be refreshed — just throw.
  if (response.status === 401 && !skipAuth) {
    tokenStorage.clearTokens();
    throw new ApiRequestError(401, "Session expired");
  }

  if (response.status >= 400) {
    const errBody = JSON.parse(response.body || "{}") as ApiError;
    throw new ApiRequestError(response.status, errBody.error);
  }

  const text = response.body;
  if (!text || text === "{}") return {} as T;

  return JSON.parse(text) as T;
}

export const apiGet = <T>(path: string, options?: FetchOptions) =>
  apiFetch<T>(path, { method: "GET", ...options });

export const apiPost = <T>(path: string, body?: unknown, options?: FetchOptions) =>
  apiFetch<T>(path, { method: "POST", body: JSON.stringify(body), ...options });

export const apiPut = <T>(path: string, body?: unknown, options?: FetchOptions) =>
  apiFetch<T>(path, { method: "PUT", body: JSON.stringify(body), ...options });

export const apiDelete = <T>(path: string, options?: FetchOptions) =>
  apiFetch<T>(path, { method: "DELETE", ...options });
