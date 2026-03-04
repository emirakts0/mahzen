import type { SearchResponse, SearchParams } from "@/types/api"
import { apiGet } from "./client"

function buildQuery(params: Record<string, string | number | undefined>): string {
  const searchParams = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== "") {
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
  })
  return apiGet<SearchResponse>(`/search/keyword?${qs}`)
}

export async function semanticSearch(params: SearchParams): Promise<SearchResponse> {
  const qs = buildQuery({
    query: params.query,
    limit: params.limit ?? 10,
    offset: params.offset ?? 0,
  })
  return apiGet<SearchResponse>(`/search/semantic?${qs}`)
}
