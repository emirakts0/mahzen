import { apiGet } from "./client";
import type { SearchResponse, SearchFilters } from "@/types/api";

interface SearchParams extends SearchFilters {
  query: string;
  limit?: number;
  offset?: number;
}

export async function keywordSearch(params: SearchParams): Promise<SearchResponse> {
  const searchParams = new URLSearchParams();
  searchParams.set("query", params.query);
  if (params.limit) searchParams.set("limit", String(params.limit));
  if (params.offset) searchParams.set("offset", String(params.offset));
  if (params.tags?.length) searchParams.set("tags", params.tags.join(","));
  if (params.path) searchParams.set("path", params.path);
  if (params.from_date) searchParams.set("from_date", params.from_date);
  if (params.to_date) searchParams.set("to_date", params.to_date);
  if (params.only_mine) searchParams.set("only_mine", "true");
  if (params.visibility) searchParams.set("visibility", params.visibility);

  return apiGet<SearchResponse>(`/v1/search/keyword?${searchParams.toString()}`);
}

export async function semanticSearch(params: SearchParams): Promise<SearchResponse> {
  const searchParams = new URLSearchParams();
  searchParams.set("query", params.query);
  if (params.limit) searchParams.set("limit", String(params.limit));
  if (params.offset) searchParams.set("offset", String(params.offset));
  if (params.tags?.length) searchParams.set("tags", params.tags.join(","));
  if (params.path) searchParams.set("path", params.path);
  if (params.from_date) searchParams.set("from_date", params.from_date);
  if (params.to_date) searchParams.set("to_date", params.to_date);
  if (params.only_mine) searchParams.set("only_mine", "true");
  if (params.visibility) searchParams.set("visibility", params.visibility);

  return apiGet<SearchResponse>(`/v1/search/semantic?${searchParams.toString()}`);
}
