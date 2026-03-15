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

export function clearBaseUrlCache() {
  cachedBaseUrl = undefined;
}

const ACCESS_TOKEN_KEY = "mahzen_access_token";
const REFRESH_TOKEN_KEY = "mahzen_refresh_token";

export const tokenStorage = {
  getAccess: (): string | null => localStorage.getItem(ACCESS_TOKEN_KEY),
  getRefresh: (): string | null => localStorage.getItem(REFRESH_TOKEN_KEY),
  setTokens: (access: string, refresh: string): void => {
    localStorage.setItem(ACCESS_TOKEN_KEY, access);
    localStorage.setItem(REFRESH_TOKEN_KEY, refresh);
  },
  clearTokens: (): void => {
    localStorage.removeItem(ACCESS_TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
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

let isRefreshing = false;
let refreshSubscribers: Array<(token: string) => void> = [];

function subscribeToRefresh(cb: (token: string) => void) {
  refreshSubscribers.push(cb);
}

function notifyRefreshSubscribers(newToken: string) {
  refreshSubscribers.forEach((cb) => cb(newToken));
  refreshSubscribers = [];
}

async function refreshAccessToken(): Promise<string> {
  const refreshToken = tokenStorage.getRefresh();
  if (!refreshToken) {
    throw new ApiRequestError(401, "No refresh token available");
  }

  const baseUrl = await getBaseUrl();
  if (!baseUrl) {
    throw new ApiRequestError(400, "Backend URL not configured");
  }
  
  const response = await doRequest(
    `${baseUrl}/v1/auth/refresh`,
    "POST",
    { "Content-Type": "application/json" },
    JSON.stringify({ refresh_token: refreshToken })
  );

  if (response.status !== 200) {
    tokenStorage.clearTokens();
    throw new ApiRequestError(401, "Session expired, please log in again");
  }

  const data = JSON.parse(response.body) as { access_token: string; refresh_token: string };
  tokenStorage.setTokens(data.access_token, data.refresh_token);
  return data.access_token;
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

  if (response.status === 401 && !skipAuth && tokenStorage.getRefresh()) {
    if (!isRefreshing) {
      isRefreshing = true;
      try {
        const newToken = await refreshAccessToken();
        isRefreshing = false;
        notifyRefreshSubscribers(newToken);
      } catch (err) {
        isRefreshing = false;
        throw err;
      }
    }

    const newToken = await new Promise<string>((resolve) => {
      subscribeToRefresh(resolve);
    });

    const retryResponse = await doRequest(
      `${baseUrl}${path}`,
      init.method || "GET",
      {
        ...headers,
        Authorization: `Bearer ${newToken}`,
      },
      init.body?.toString()
    );

    if (retryResponse.status >= 400) {
      const errBody = JSON.parse(retryResponse.body || "{}") as ApiError;
      throw new ApiRequestError(retryResponse.status, errBody.error);
    }

    const text = retryResponse.body;
    if (!text || text === "{}") return {} as T;
    return JSON.parse(text) as T;
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
