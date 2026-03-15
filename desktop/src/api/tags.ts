import { apiGet, apiPost } from "@/api/client"
import type { TagsResponse, ListTagsParams, CreateTagRequest, TagResponse } from "@/types/api"

export function listTags(params: ListTagsParams = {}): Promise<TagsResponse> {
  const sp = new URLSearchParams()
  if (params.limit !== undefined) sp.set("limit", String(params.limit))
  if (params.offset !== undefined) sp.set("offset", String(params.offset))
   const qs = sp.toString()
   return apiGet<TagsResponse>(`/v1/tags${qs ? `?${qs}` : ""}`)
 }
 
 export async function createTag(data: CreateTagRequest): Promise<TagResponse["tag"]> {
   const res = await apiPost<TagResponse>("/v1/tags", data)
   return res.tag
 }
