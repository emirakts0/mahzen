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

export interface Highlight {
  field: "title" | "content" | "summary"
  snippet: string
}

export interface SearchResult {
  entry_id: string
  title: string
  summary?: string
  content?: string
  score: number
  highlights?: Highlight[]
  path: string
  visibility: "public" | "private"
  tags: string[]
  created_at: string
  file_type?: string
  file_size?: number
  s3_key?: string
}

export interface SearchResponse {
  results: SearchResult[]
  total: number
}

export interface SearchParams {
  query: string
  limit?: number
  offset?: number
  tags?: string[]
  path?: string
  from_date?: string
  to_date?: string
  only_mine?: boolean
  visibility?: "public" | "private"
}

export interface SearchFilters {
  tags?: string[]
  path?: string
  from_date?: string
  to_date?: string
  only_mine?: boolean
  visibility?: "public" | "private"
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
  file_type?: string
  file_size?: number
  s3_key?: string
}

export interface EntryResponse {
  entry: Entry
}

export interface FolderInfo {
  path: string
  count: number
}

export interface EntriesResponse {
  entries: Entry[]
  folders: FolderInfo[]
  total: number
}

export interface CreateEntryRequest {
  title?: string
  content?: string
  path?: string
  visibility?: "public" | "private"
  tag_ids?: string[]
  file_type?: string
}

export interface UpdateEntryRequest {
  title?: string
  content?: string
  path?: string
  visibility?: "public" | "private"
  tag_ids?: string[]
  file_type?: string
}

export interface ListEntriesParams {
  path?: string
  limit?: number
  offset?: number
  own?: boolean
  visibility?: "public" | "private"
  tags?: string[]
  from_date?: string
  to_date?: string
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
