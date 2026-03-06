import { useState } from "react"
import { Copy, Download, FileIcon } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import type { SearchResult } from "@/types/api"
import { toast } from "sonner"
import { EntryPreviewModal } from "./entry-preview-modal"
import { PathBreadcrumb, VisibilityLabel, DateLabel } from "./card-meta"

interface KeywordResultCardProps {
  result: SearchResult
  className?: string
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

export function KeywordResultCard({ result, className }: KeywordResultCardProps) {
  const [copying, setCopying] = useState(false)
  const [previewOpen, setPreviewOpen] = useState(false)

  const isBinary = Boolean(result.file_type && result.s3_key)

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

  const handleDownload = (e: React.MouseEvent) => {
    e.stopPropagation()
    window.open(`/v1/entries/${result.entry_id}/download`, "_blank")
  }

  // Pick highlights by field
  const titleHighlight = result.highlights?.find(h => h.field === "title")
  const contentHighlight = result.highlights?.find(h => h.field === "content")
  const summaryHighlight = result.highlights?.find(h => h.field === "summary")

  // Content body: prefer content highlight → inline content
  const contentBody = contentHighlight?.snippet ?? result.content ?? null

  // Summary line: prefer summary highlight → plain summary
  const summaryBody = summaryHighlight?.snippet ?? result.summary ?? null

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
        {/* ── Header: title + copy/download ── */}
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            {/* Title — smaller, secondary role */}
            <span
              className="block truncate text-xs font-medium leading-snug"
              style={{ color: "var(--glass-text-muted)" }}
            >
              {titleHighlight
                ? <MarkedText html={titleHighlight.snippet} />
                : result.title || "Untitled"}
            </span>
            <PathBreadcrumb path={result.path} />
          </div>

          {isBinary ? (
            <button
              onClick={handleDownload}
              className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-md opacity-0 transition-opacity group-hover:opacity-100"
              style={{ color: "var(--glass-icon)" }}
              title="Download"
            >
              <Download className="h-3.5 w-3.5" />
            </button>
          ) : (
            <button
              onClick={e => { e.stopPropagation(); void handleCopy() }}
              className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-md opacity-0 transition-opacity group-hover:opacity-100"
              style={{ color: "var(--glass-icon)" }}
              title="Copy"
              disabled={copying}
            >
              <Copy className="h-3.5 w-3.5" />
            </button>
          )}
        </div>

        {/* ── Binary file info ── */}
        {isBinary && (
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

        {/* ── Content highlight (primary body, text entries only) ── */}
        {!isBinary && contentBody && (
          <p
            className="mt-2 line-clamp-3 text-sm leading-relaxed"
            style={{ color: "var(--glass-text)" }}
          >
            <MarkedText html={contentBody} />
          </p>
        )}

        {/* ── Summary (secondary, only if different context from content) ── */}
        {summaryBody && (
          <p
            className="mt-1.5 line-clamp-2 text-xs leading-relaxed"
            style={{ color: "var(--glass-text-muted)" }}
          >
            <MarkedText html={summaryBody} />
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
