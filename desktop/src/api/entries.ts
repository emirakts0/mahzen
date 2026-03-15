import { apiGet, apiPost } from "@/api/client"
import type { Entry, EntryResponse, EntriesResponse, CreateEntryRequest, ListEntriesParams } from "@/types/api"

export function listEntries(params: ListEntriesParams = {}): Promise<EntriesResponse> {
  const sp = new URLSearchParams()
  if (params.path) sp.set("path", params.path)
  if (params.limit !== undefined) sp.set("limit", String(params.limit))
  if (params.offset !== undefined) sp.set("offset", String(params.offset))
  if (params.own) sp.set("own", "true")
  if (params.visibility) sp.set("visibility", params.visibility)
  if (params.tags?.length) sp.set("tags", params.tags.join(","))
  if (params.from_date) sp.set("from_date", params.from_date)
  if (params.to_date) sp.set("to_date", params.to_date)
   const qs = sp.toString()
   return apiGet<EntriesResponse>(`/v1/entries${qs ? `?${qs}` : ""}`)
}

export async function getEntry(entryId: string): Promise<Entry> {
   const res = await apiGet<EntryResponse>(`/v1/entries/${entryId}`)
   return res.entry
 }

export async function createEntry(data: CreateEntryRequest): Promise<Entry> {
   const res = await apiPost<EntryResponse>("/v1/entries", data)
   return res.entry
 }
