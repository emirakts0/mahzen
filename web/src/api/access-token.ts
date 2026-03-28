import { apiGet, apiPost } from "@/api/client"
import type {
  AccessTokenListResponse,
  CreateTokenRequest,
  CreateTokenResponse,
} from "@/types/api"

export function listAccessTokens(): Promise<AccessTokenListResponse> {
  return apiGet<AccessTokenListResponse>("/tokens")
}

export function createAccessToken(data: CreateTokenRequest): Promise<CreateTokenResponse> {
  return apiPost<CreateTokenResponse>("/tokens", data)
}

export function revokeAccessToken(id: string): Promise<unknown> {
  return apiPost(`/tokens/${id}/revoke`)
}
