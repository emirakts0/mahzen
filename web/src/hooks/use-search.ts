import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api"

export function useKeywordSearch(params: {
  query: string
  limit?: number
  offset?: number
}) {
  return useQuery({
    queryKey: ["search", "keyword", params],
    queryFn: () => api.keywordSearch(params),
    enabled: params.query.length > 0,
  })
}

export function useSemanticSearch(params: {
  query: string
  limit?: number
  offset?: number
}) {
  return useQuery({
    queryKey: ["search", "semantic", params],
    queryFn: () => api.semanticSearch(params),
    enabled: params.query.length > 0,
  })
}
