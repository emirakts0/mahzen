import { Lock, Globe } from "lucide-react"

// ── Path ──────────────────────────────────────────────────────────────────────

/**
 * Renders a path like `/work/notes` as:
 *   / work / notes
 * Lowercase, monospace, slash-separated, very muted.
 */
export function PathBreadcrumb({ path }: { path: string }) {
  if (!path || path === "/") return null
  const parts = path.replace(/^\//, "").split("/").filter(Boolean)
  if (parts.length === 0) return null

  return (
    <div
      className="mt-0.5 flex min-w-0 items-center gap-0 overflow-hidden text-xs"
      style={{ color: "var(--glass-text-muted)", fontFamily: "inherit" }}
    >
      <span className="mr-0.5 opacity-40">/</span>
      {parts.map((part, i) => (
        <span key={i} className="flex min-w-0 shrink-0 items-center">
          {i > 0 && <span className="mx-1 opacity-30">/</span>}
          <span className="truncate opacity-70">{part}</span>
        </span>
      ))}
    </div>
  )
}

// ── Visibility ────────────────────────────────────────────────────────────────

export function VisibilityLabel({
  visibility,
}: {
  visibility: "public" | "private"
}) {
  const isPrivate = visibility === "private"
  return (
    <span
      className="flex shrink-0 items-center gap-1 text-xs"
      style={{ color: "var(--glass-text-muted)" }}
    >
      {isPrivate ? (
        <Lock className="h-3 w-3 opacity-70" />
      ) : (
        <Globe className="h-3 w-3 opacity-70" />
      )}
      <span className="opacity-70">{isPrivate ? "private" : "public"}</span>
    </span>
  )
}

// ── Date ──────────────────────────────────────────────────────────────────────

function formatDate(iso: string): string {
  if (!iso) return ""
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ""
  const now = new Date()
  const sameYear = date.getFullYear() === now.getFullYear()
  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    ...(!sameYear && { year: "numeric" }),
  })
}

export function DateLabel({ date }: { date: string }) {
  const label = formatDate(date)
  if (!label) return null
  return (
    <span
      className="shrink-0 text-xs tabular-nums"
      style={{ color: "var(--glass-text-muted)", opacity: 0.65 }}
    >
      {label}
    </span>
  )
}
