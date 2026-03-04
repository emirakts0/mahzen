import { apiGet, apiPost, apiDelete } from "@/api/client"
import type {
  Tag,
  TagResponse,
  TagsResponse,
  CreateTagRequest,
  ListTagsParams,
} from "@/types/api"

export function listTags(params: ListTagsParams = {}): Promise<TagsResponse> {
  const sp = new URLSearchParams()
  if (params.limit !== undefined) sp.set("limit", String(params.limit))
  if (params.offset !== undefined) sp.set("offset", String(params.offset))
  const qs = sp.toString()
  return apiGet<TagsResponse>(`/tags${qs ? `?${qs}` : ""}`)
}

export async function getTag(tagId: string): Promise<Tag> {
  const res = await apiGet<TagResponse>(`/tags/${tagId}`)
  return res.tag
}

export async function createTag(data: CreateTagRequest): Promise<Tag> {
  const res = await apiPost<TagResponse>("/tags", data)
  return res.tag
}

export function deleteTag(tagId: string): Promise<unknown> {
  return apiDelete(`/tags/${tagId}`)
}
