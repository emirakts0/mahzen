import { useState } from "react"
import { Copy, FileIcon } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import type { SearchResult } from "@/types/api"
import { toast } from "sonner"
import { EntryPreviewModal } from "./entry-preview-modal"
import { PathBreadcrumb, VisibilityLabel, DateLabel } from "./card-meta"

const MATCH_THRESHOLDS = {
  HIGH: 80,
  MEDIUM: 60,
} as const

interface SemanticResultCardProps {
  result: SearchResult
  className?: string
}

function MatchBadge({ score }: { score: number }) {
  const percent = Math.round(score * 100)
  const style =
    percent >= MATCH_THRESHOLDS.HIGH
      ? { bg: "rgba(34, 197, 94, 0.15)", color: "#22c55e" }
      : percent >= MATCH_THRESHOLDS.MEDIUM
        ? { bg: "rgba(59, 130, 246, 0.15)", color: "#3b82f6" }
        : { bg: "var(--glass-hover)", color: "var(--glass-text-muted)" }

  return (
    <span
      className="shrink-0 rounded-full px-2 py-0.5 text-xs font-semibold tabular-nums"
      style={{ background: style.bg, color: style.color }}
    >
      {percent}% match
    </span>
  )
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function SemanticResultCard({ result, className }: SemanticResultCardProps) {
  const [copying, setCopying] = useState(false)
  const [previewOpen, setPreviewOpen] = useState(false)

  const handleCopy = async () => {
    setCopying(true)
    try {
      await navigator.clipboard.writeText(result.content ?? result.summary ?? "")
      toast.success("Copied")
    } catch {
      toast.error("Failed to copy")
    } finally {
      setCopying(false)
    }
  }

  return (
    <>
      <div
        className={cn(
          "group relative flex flex-col rounded-xl border p-4 shadow-sm transition-all backdrop-blur-md cursor-pointer",
          className,
        )}
        style={{
          background: "var(--glass-bg)",
          borderColor: "var(--glass-border)",
        }}
        onClick={() => setPreviewOpen(true)}
        onMouseEnter={e => { (e.currentTarget as HTMLElement).style.background = "var(--glass-hover)" }}
        onMouseLeave={e => { (e.currentTarget as HTMLElement).style.background = "var(--glass-bg)" }}
      >
        {/* ── Header: title + match badge + copy ── */}
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            {/* Title — smaller, secondary role */}
            <span
              className="block truncate text-xs font-medium leading-snug"
              style={{ color: "var(--glass-text-muted)" }}
            >
              {result.title || "Untitled"}
            </span>
            <PathBreadcrumb path={result.path} />
          </div>

          <div className="mt-0.5 flex shrink-0 items-center gap-1.5">
            {result.score > 0 && <MatchBadge score={result.score} />}
            <button
              onClick={e => { e.stopPropagation(); void handleCopy() }}
              className="flex h-7 w-7 items-center justify-center rounded-md opacity-0 transition-opacity group-hover:opacity-100"
              style={{ color: "var(--glass-icon)" }}
              title="Copy"
              disabled={copying}
            >
              <Copy className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>

        {/* ── File type info ── */}
        {result.file_type && (
          <div
            className="mt-2 flex items-center gap-2 text-xs"
            style={{ color: "var(--glass-text-muted)" }}
          >
            <FileIcon className="h-3.5 w-3.5 shrink-0" />
            <span className="font-mono uppercase">{result.file_type}</span>
            {result.file_size !== undefined && (
              <span>{formatFileSize(result.file_size)}</span>
            )}
          </div>
        )}

        {/* ── Inline content (primary body) ── */}
        {result.content && (
          <p
            className="mt-2 line-clamp-3 text-sm leading-relaxed"
            style={{ color: "var(--glass-text)" }}
          >
            {result.content}
          </p>
        )}

        {/* ── Summary (secondary) ── */}
        {result.summary && (
          <p
            className="mt-1.5 line-clamp-2 text-xs leading-relaxed"
            style={{ color: "var(--glass-text-muted)" }}
          >
            {result.summary}
          </p>
        )}

        {/* ── Footer: tags · visibility · date ── */}
        <div className="mt-3 flex items-center justify-between gap-2">
          <div className="flex min-w-0 flex-wrap gap-1">
            {result.tags?.map(tag => (
              <Badge key={tag} variant="secondary" className="h-5 px-2 text-xs font-normal">
                #{tag}
              </Badge>
            ))}
          </div>
          <div className="flex shrink-0 items-center gap-2">
            <VisibilityLabel visibility={result.visibility} />
            <span style={{ color: "var(--glass-divider)" }}>·</span>
            <DateLabel date={result.created_at} />
          </div>
        </div>
      </div>

      <EntryPreviewModal
        entryId={previewOpen ? result.entry_id : null}
        onClose={() => setPreviewOpen(false)}
      />
    </>
  )
}
