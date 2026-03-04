// ─────────────────────────────────────────────
// Auth
// ─────────────────────────────────────────────

export interface AuthTokens {
  access_token: string
  refresh_token: string
}

export interface RegisterRequest {
  email: string
  display_name?: string
  password: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LogoutRequest {
  refresh_token: string
}

export interface RefreshRequest {
  refresh_token: string
}

export interface User {
  id: string
  email: string
  display_name: string
}

export interface UserResponse {
  user: User
}

// ─────────────────────────────────────────────
// Search
// ─────────────────────────────────────────────

export interface SearchResult {
  entry_id: string
  title: string
  snippet: string
  score: number
  highlights: string[]
  path: string
  visibility: "public" | "private"
  tags: string[]
  created_at: string
}

export interface SearchResponse {
  results: SearchResult[]
  total: number
}

export interface SearchParams {
  query: string
  limit?: number
  offset?: number
}

// ─────────────────────────────────────────────
// Entries
// ─────────────────────────────────────────────

export interface Entry {
  id: string
  user_id: string
  title: string
  content: string
  summary?: string
  path: string
  visibility: "public" | "private"
  tags?: string[]
  created_at: string
  updated_at: string
}

export interface EntryResponse {
  entry: Entry
}

export interface EntriesResponse {
  entries: Entry[]
  total: number
}

export interface CreateEntryRequest {
  title?: string
  content?: string
  path?: string
  visibility?: "public" | "private"
  tag_ids?: string[]
}

export interface UpdateEntryRequest {
  title?: string
  content?: string
  path?: string
  visibility?: "public" | "private"
  tag_ids?: string[]
}

export interface ListEntriesParams {
  path?: string
  limit?: number
  offset?: number
}

// ─────────────────────────────────────────────
// Tags
// ─────────────────────────────────────────────

export interface Tag {
  id: string
  name: string
  slug: string
  created_at: string
}

export interface TagResponse {
  tag: Tag
}

export interface TagsResponse {
  tags: Tag[]
  total: number
}

export interface CreateTagRequest {
  name: string
}

export interface AttachTagRequest {
  tag_id: string
}

export interface ListTagsParams {
  limit?: number
  offset?: number
}

// ─────────────────────────────────────────────
// API Error
// ─────────────────────────────────────────────

export interface ApiError {
  error: string
}
