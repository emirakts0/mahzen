import { API_BASE_URL } from "./constants"
import type {
  ApiError,
  AttachTagRequest,
  CreateEntryRequest,
  CreateTagRequest,
  EntryResponse,
  GetCurrentUserResponse,
  ListEntriesResponse,
  ListTagsResponse,
  SearchResponse,
  TagResponse,
  UpdateEntryRequest,
} from "./types"

// ---------------------------------------------------------------------------
// Auth types
// ---------------------------------------------------------------------------

interface AuthResponse {
  access_token: string
  refresh_token: string
}

// ---------------------------------------------------------------------------
// Token access (lazy import to avoid circular dependency)
// ---------------------------------------------------------------------------

function getAccessToken(): string | null {
  return localStorage.getItem("mahzen_access_token")
}

// ---------------------------------------------------------------------------
// Base fetch wrapper
// ---------------------------------------------------------------------------

class ApiClient {
  private baseUrl: string

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl
  }

  private async request<T>(
    path: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${path}`
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string>),
    }

    // Attach Bearer token if available
    const token = getAccessToken()
    if (token) {
      headers["Authorization"] = `Bearer ${token}`
    }

    const response = await fetch(url, {
      ...options,
      headers,
    })

    if (!response.ok) {
      let error: ApiError
      try {
        error = await response.json()
      } catch {
        error = { code: response.status, message: response.statusText }
      }
      throw error
    }

    // Handle 204 No Content or empty body
    const text = await response.text()
    if (!text) return {} as T
    return JSON.parse(text)
  }

  // -------------------------------------------------------------------------
  // Auth
  // -------------------------------------------------------------------------

  async login(email: string, password: string): Promise<AuthResponse> {
    return this.request<AuthResponse>("/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    })
  }

  async register(
    email: string,
    displayName: string,
    password: string
  ): Promise<AuthResponse> {
    return this.request<AuthResponse>("/v1/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, display_name: displayName, password }),
    })
  }

  async refreshToken(refreshToken: string): Promise<AuthResponse> {
    return this.request<AuthResponse>("/v1/auth/refresh", {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
  }

  async logout(refreshToken: string): Promise<void> {
    await this.request("/v1/auth/logout", {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
  }

  // -------------------------------------------------------------------------
  // Entries
  // -------------------------------------------------------------------------

  async createEntry(data: CreateEntryRequest): Promise<EntryResponse> {
    return this.request<EntryResponse>("/v1/entries", {
      method: "POST",
      body: JSON.stringify(data),
    })
  }

  async listEntries(params?: {
    path?: string
    limit?: number
    offset?: number
  }): Promise<ListEntriesResponse> {
    const searchParams = new URLSearchParams()
    if (params?.path) searchParams.set("path", params.path)
    if (params?.limit) searchParams.set("limit", String(params.limit))
    if (params?.offset) searchParams.set("offset", String(params.offset))
    const qs = searchParams.toString()
    return this.request<ListEntriesResponse>(
      `/v1/entries${qs ? `?${qs}` : ""}`
    )
  }

  async getEntry(id: string): Promise<EntryResponse> {
    return this.request<EntryResponse>(`/v1/entries/${id}`)
  }

  async updateEntry(
    id: string,
    data: UpdateEntryRequest
  ): Promise<EntryResponse> {
    return this.request<EntryResponse>(`/v1/entries/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    })
  }

  async deleteEntry(id: string): Promise<void> {
    await this.request(`/v1/entries/${id}`, { method: "DELETE" })
  }

  // -------------------------------------------------------------------------
  // Tags
  // -------------------------------------------------------------------------

  async createTag(data: CreateTagRequest): Promise<TagResponse> {
    return this.request<TagResponse>("/v1/tags", {
      method: "POST",
      body: JSON.stringify(data),
    })
  }

  async listTags(params?: {
    limit?: number
    offset?: number
  }): Promise<ListTagsResponse> {
    const searchParams = new URLSearchParams()
    if (params?.limit) searchParams.set("limit", String(params.limit))
    if (params?.offset) searchParams.set("offset", String(params.offset))
    const qs = searchParams.toString()
    return this.request<ListTagsResponse>(`/v1/tags${qs ? `?${qs}` : ""}`)
  }

  async getTag(id: string): Promise<TagResponse> {
    return this.request<TagResponse>(`/v1/tags/${id}`)
  }

  async deleteTag(id: string): Promise<void> {
    await this.request(`/v1/tags/${id}`, { method: "DELETE" })
  }

  async attachTag(
    entryId: string,
    data: AttachTagRequest
  ): Promise<void> {
    await this.request(`/v1/entries/${entryId}/tags`, {
      method: "POST",
      body: JSON.stringify(data),
    })
  }

  async detachTag(entryId: string, tagId: string): Promise<void> {
    await this.request(`/v1/entries/${entryId}/tags/${tagId}`, {
      method: "DELETE",
    })
  }

  // -------------------------------------------------------------------------
  // Search
  // -------------------------------------------------------------------------

  async keywordSearch(params: {
    query: string
    limit?: number
    offset?: number
  }): Promise<SearchResponse> {
    const searchParams = new URLSearchParams()
    searchParams.set("query", params.query)
    if (params.limit) searchParams.set("limit", String(params.limit))
    if (params.offset) searchParams.set("offset", String(params.offset))
    return this.request<SearchResponse>(
      `/v1/search/keyword?${searchParams.toString()}`
    )
  }

  async semanticSearch(params: {
    query: string
    limit?: number
    offset?: number
  }): Promise<SearchResponse> {
    const searchParams = new URLSearchParams()
    searchParams.set("query", params.query)
    if (params.limit) searchParams.set("limit", String(params.limit))
    if (params.offset) searchParams.set("offset", String(params.offset))
    return this.request<SearchResponse>(
      `/v1/search/semantic?${searchParams.toString()}`
    )
  }

  // -------------------------------------------------------------------------
  // Users
  // -------------------------------------------------------------------------

  async getCurrentUser(): Promise<GetCurrentUserResponse> {
    return this.request<GetCurrentUserResponse>("/v1/users/me")
  }
}

export const api = new ApiClient(API_BASE_URL)
