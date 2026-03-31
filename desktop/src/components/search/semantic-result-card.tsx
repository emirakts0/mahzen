import { useState, forwardRef } from "react"
import { Copy, FileIcon, Check } from "lucide-react"
import { writeText } from "@tauri-apps/plugin-clipboard-manager"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import type { SearchResult } from "@/types/api"
import { EntryPreviewModal } from "./entry-preview-modal"
import { PathBreadcrumb, VisibilityLabel, DateLabel } from "./card-meta"

const MATCH_THRESHOLDS = {
  HIGH: 80,
  MEDIUM: 60,
} as const

interface SemanticResultCardProps {
  result: SearchResult
  className?: string
  isSelected?: boolean
}

function MatchBadge({ score }: { score: number }) {
  const percent = Math.round(score * 100)
  const style =
    percent >= MATCH_THRESHOLDS.HIGH
      ? { bg: "var(--glass-success-bg)", color: "var(--glass-success)" }
      : percent >= MATCH_THRESHOLDS.MEDIUM
        ? { bg: "rgba(59, 130, 246, 0.15)", color: "#3b82f6" }
        : { bg: "var(--glass-hover)", color: "var(--glass-text-muted)" }

  return (
    <span
      className="shrink-0 rounded-full px-1.5 py-0.5 text-[9px] font-semibold tabular-nums"
      style={{ background: style.bg, color: style.color }}
    >
      {percent}%
    </span>
  )
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export const SemanticResultCard = forwardRef<HTMLDivElement, SemanticResultCardProps>(
  function SemanticResultCard({ result, className, isSelected }, ref) {
  const [copied, setCopied] = useState(false)
  const [previewOpen, setPreviewOpen] = useState(false)

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation()
    setCopied(true)
    try {
      await writeText(result.content ?? result.summary ?? "")
    } catch {
      console.error("Failed to copy")
    } finally {
      setTimeout(() => setCopied(false), 1500)
    }
  }

  return (
    <>
      <div
        ref={ref}
        className={cn(
          "group relative flex flex-col rounded-lg border p-2.5 shadow-sm transition-all backdrop-blur-md cursor-pointer",
          isSelected && "ring-2 ring-primary/50",
          className
        )}
        style={{
          background: isSelected ? "var(--glass-hover)" : "var(--glass-bg)",
          borderColor: isSelected ? "var(--color-primary)" : "var(--glass-border)",
        }}
        onClick={() => setPreviewOpen(true)}
        onMouseEnter={(e) => {
          if (!isSelected) {
            (e.currentTarget as HTMLElement).style.background = "var(--glass-hover)"
          }
        }}
        onMouseLeave={(e) => {
          if (!isSelected) {
            (e.currentTarget as HTMLElement).style.background = "var(--glass-bg)"
          }
        }}
      >
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0 flex-1">
            <span
              className="block truncate text-[11px] font-medium leading-snug"
              style={{ color: "var(--glass-text-muted)" }}
            >
              {result.title || "Untitled"}
            </span>
            <PathBreadcrumb path={result.path} />
          </div>

          <div className="mt-0.5 flex shrink-0 items-center gap-1">
            {(result.score ?? 0) > 0 && <MatchBadge score={result.score!} />}
            <button
              onClick={handleCopy}
              className="flex h-5 w-5 items-center justify-center rounded-md opacity-0 transition-opacity group-hover:opacity-100"
              style={{ color: copied ? "oklch(0.7 0.15 150)" : "var(--glass-icon)" }}
              title="Copy"
            >
              {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
            </button>
          </div>
        </div>

        {result.file_type && (
          <div
            className="mt-1.5 flex items-center gap-1.5 text-[10px]"
            style={{ color: "var(--glass-text-muted)" }}
          >
            <FileIcon className="h-3 w-3 shrink-0" />
            <span className="font-mono uppercase">{result.file_type}</span>
            {result.file_size !== undefined && (
              <span>{formatFileSize(result.file_size)}</span>
            )}
          </div>
        )}

        {result.content && (
          <p
            className="mt-1.5 line-clamp-2 text-[11px] leading-relaxed"
            style={{ color: "var(--glass-text)" }}
          >
            {result.content}
          </p>
        )}

        {result.summary && (
          <p
            className="mt-1 line-clamp-1 text-[10px] leading-relaxed"
            style={{ color: "var(--glass-text-muted)" }}
          >
            {result.summary}
          </p>
        )}

        <div className="mt-2 flex items-center justify-between gap-2">
          <div className="flex min-w-0 flex-wrap gap-0.5">
            {result.tags?.slice(0, 3).map((tag) => (
              <Badge key={tag} variant="secondary" className="h-4 px-1.5 text-[9px] font-normal">
                #{tag}
              </Badge>
            ))}
            {(result.tags?.length ?? 0) > 3 && (
              <span className="text-[9px]" style={{ color: "var(--glass-text-muted)" }}>+{result.tags!.length - 3}</span>
            )}
          </div>
          <div className="flex shrink-0 items-center gap-1.5 text-[10px]">
            <VisibilityLabel visibility={result.visibility} isOwner={result.is_mine} />
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
)
