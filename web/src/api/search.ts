import type { SearchResponse, SearchParams } from "@/types/api"
import { apiGet } from "./client"

function buildQuery(params: Record<string, string | number | boolean | undefined>): string {
  const searchParams = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== "" && value !== false) {
      searchParams.set(key, String(value))
    }
  }
  return searchParams.toString()
}

export async function keywordSearch(params: SearchParams): Promise<SearchResponse> {
  const qs = buildQuery({
    query: params.query,
    limit: params.limit ?? 10,
    offset: params.offset ?? 0,
    tags: params.tags?.join(","),
    path: params.path,
    from_date: params.from_date,
    to_date: params.to_date,
    only_mine: params.only_mine,
    visibility: params.visibility,
  })
  return apiGet<SearchResponse>(`/search/keyword?${qs}`)
}

export async function semanticSearch(params: SearchParams): Promise<SearchResponse> {
  const qs = buildQuery({
    query: params.query,
    limit: params.limit ?? 10,
    offset: params.offset ?? 0,
    tags: params.tags?.join(","),
    path: params.path,
    from_date: params.from_date,
    to_date: params.to_date,
    only_mine: params.only_mine,
    visibility: params.visibility,
  })
  return apiGet<SearchResponse>(`/search/semantic?${qs}`)
}
