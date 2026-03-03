// ---------------------------------------------------------------------------
// Domain types matching the backend JSON schema
// ---------------------------------------------------------------------------

export type Visibility = "public" | "private"

// -- Entry ------------------------------------------------------------------

export interface Entry {
  id: string
  user_id: string
  title: string
  content: string
  summary: string
  path: string
  visibility: Visibility
  tags: string[]
  created_at: string
  updated_at: string
}

export interface CreateEntryRequest {
  title: string
  content: string
  path?: string
  visibility?: Visibility
  tag_ids?: string[]
}

export interface UpdateEntryRequest {
  title?: string
  content?: string
  path?: string
  visibility?: Visibility
  tag_ids?: string[]
}

export interface ListEntriesResponse {
  entries: Entry[]
  total: number
}

export interface EntryResponse {
  entry: Entry
}

// -- Tag --------------------------------------------------------------------

export interface Tag {
  id: string
  name: string
  slug: string
  created_at: string
}

export interface CreateTagRequest {
  name: string
}

export interface ListTagsResponse {
  tags: Tag[]
  total: number
}

export interface TagResponse {
  tag: Tag
}

export interface AttachTagRequest {
  tag_id: string
}

// -- Search -----------------------------------------------------------------

export interface SearchResult {
  entry_id: string
  title: string
  snippet: string
  score: number
  highlights: string[]
}

export interface SearchResponse {
  results: SearchResult[]
  total: number
}

// -- User -------------------------------------------------------------------

export interface User {
  id: string
  email: string
  display_name: string
}

export interface GetCurrentUserResponse {
  user: User
}

// -- Error ------------------------------------------------------------------

export interface ApiError {
  code: number
  message: string
}

// -- Pagination -------------------------------------------------------------

export interface PaginationParams {
  limit?: number
  offset?: number
}
