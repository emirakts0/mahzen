import { useState, forwardRef } from "react"
import { Copy, FileIcon, Check } from "lucide-react"
import { writeText } from "@tauri-apps/plugin-clipboard-manager"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import type { SearchResult } from "@/types/api"
import { EntryPreviewModal } from "./entry-preview-modal"
import { PathBreadcrumb, VisibilityLabel, DateLabel } from "./card-meta"

interface KeywordResultCardProps {
  result: SearchResult
  className?: string
  isSelected?: boolean
}

function MarkedText({ html }: { html: string }) {
  return (
    <span
      className="[&_mark]:rounded-sm [&_mark]:px-0.5 [&_mark]:font-semibold"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  )
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export const KeywordResultCard = forwardRef<HTMLDivElement, KeywordResultCardProps>(
  function KeywordResultCard({ result, className, isSelected }, ref) {
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

  const titleHighlight = result.highlights?.find((h) => h.field === "title")
  const contentHighlight = result.highlights?.find((h) => h.field === "content")
  const summaryHighlight = result.highlights?.find((h) => h.field === "summary")

  const contentBody = contentHighlight?.snippet ?? result.content ?? null
  const summaryBody = summaryHighlight?.snippet ?? result.summary ?? null

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
              {titleHighlight ? (
                <MarkedText html={titleHighlight.snippet} />
              ) : (
                result.title || "Untitled"
              )}
            </span>
            <PathBreadcrumb path={result.path} />
          </div>

          <button
            onClick={handleCopy}
            className="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-md opacity-0 transition-opacity group-hover:opacity-100"
            style={{ color: copied ? "oklch(0.7 0.15 150)" : "var(--glass-icon)" }}
            title="Copy"
          >
            {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
          </button>
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

        {contentBody && (
          <p
            className="mt-1.5 line-clamp-2 text-[11px] leading-relaxed"
            style={{ color: "var(--glass-text)" }}
          >
            <MarkedText html={contentBody} />
          </p>
        )}

        {summaryBody && (
          <p
            className="mt-1 line-clamp-1 text-[10px] leading-relaxed"
            style={{ color: "var(--glass-text-muted)" }}
          >
            <MarkedText html={summaryBody} />
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
