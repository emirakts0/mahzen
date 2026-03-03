// In production (served from Go binary), API is same-origin so we use ""
// which makes fetch use relative URLs. Override via VITE_API_BASE_URL.
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ""

export const DEFAULT_PAGE_SIZE = 20

export const VISIBILITY = {
  PUBLIC: "public",
  PRIVATE: "private",
} as const

export const VISIBILITY_LABELS: Record<string, string> = {
  public: "Public",
  private: "Private",
}
