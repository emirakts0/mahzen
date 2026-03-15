export interface SearchResult {
  entry_id: string
  is_mine: boolean
  title: string
  summary?: string
  content?: string
  score?: number
  highlights?: { field: string; snippet: string }[]
  path: string
  visibility: "public" | "private"
  tags: string[]
  created_at: string
  file_type?: string
  file_size?: number
}

export interface SearchFilters {
  tags?: string[]
  path?: string
  from_date?: string
  to_date?: string
  only_mine?: boolean
  visibility?: "public" | "private"
}

export interface SearchResponse {
  results: SearchResult[]
  total: number
}

export interface ApiError {
  error: string
}

export interface AuthTokens {
  access_token: string
  refresh_token: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  display_name?: string
}

export interface User {
  id: string
  email: string
  display_name: string
}

export interface UserResponse {
  user: User
}

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
}

export interface EntryResponse {
  entry: Entry
}

export interface Tag {
  id: string
  name: string
  slug: string
  created_at: string
}

export interface TagsResponse {
  tags: Tag[]
  total: number
}

export interface ListTagsParams {
  limit?: number
  offset?: number
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

export interface CreateTagRequest {
  name: string
}

export interface TagResponse {
  tag: Tag
}
