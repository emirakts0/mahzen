import { useState } from "react"
import { Copy, ChevronRight } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import type { SearchResult } from "@/types/api"
import { toast } from "sonner"
import { EntryPreviewModal } from "./entry-preview-modal"

interface KeywordResultCardProps {
  result: SearchResult
  className?: string
}

function PathBreadcrumb({ path }: { path: string }) {
  if (!path || path === "/") return null
  const parts = path.replace(/^\//, "").split("/").filter(Boolean)
  return (
    <div className="flex flex-wrap items-center gap-0.5 text-xs" style={{ color: "var(--glass-text-muted)" }}>
      {parts.map((part, i) => (
        <span key={i} className="flex items-center gap-0.5">
          {i > 0 && <ChevronRight className="h-3 w-3 opacity-50" />}
          <span className="font-medium uppercase tracking-wide">{part}</span>
        </span>
      ))}
    </div>
  )
}

function HighlightedText({ html }: { html: string }) {
  return (
    <span
      className="[&_mark]:rounded-sm [&_mark]:px-0.5 [&_mark]:font-semibold"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  )
}

export function KeywordResultCard({ result, className }: KeywordResultCardProps) {
  const [copying, setCopying] = useState(false)
  const [previewOpen, setPreviewOpen] = useState(false)

  const handleCopy = async () => {
    setCopying(true)
    try {
      await navigator.clipboard.writeText(result.snippet)
      toast.success("Snippet copied")
    } catch {
      toast.error("Failed to copy")
    } finally {
      setCopying(false)
    }
  }

  const primaryHighlight = result.highlights?.[0] ?? result.snippet

  return (
    <>
      <div
        className={cn(
          "group relative rounded-xl border p-4 shadow-sm transition-all backdrop-blur-md cursor-pointer",
          className,
        )}
        style={{
          background: "var(--glass-bg)",
          borderColor: "var(--glass-border)",
        }}
        onClick={() => setPreviewOpen(true)}
        onMouseEnter={e => {
          const target = e.currentTarget as HTMLElement
          target.style.background = "var(--glass-hover)"
          target.style.borderColor = "var(--glass-border)"
        }}
        onMouseLeave={e => {
          const target = e.currentTarget as HTMLElement
          target.style.background = "var(--glass-bg)"
          target.style.borderColor = "var(--glass-border)"
        }}
      >
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <span
              className="truncate text-sm font-semibold"
              style={{ color: "var(--glass-text)" }}
            >
              {result.title || "Untitled"}
            </span>
            <PathBreadcrumb path={result.path} />
          </div>

          <div className="flex shrink-0 items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
            <button
              onClick={(e) => {
                e.stopPropagation()
                void handleCopy()
              }}
              className="flex h-7 w-7 items-center justify-center rounded-md transition-colors"
              style={{ color: "var(--glass-icon)" }}
              title="Copy snippet"
              disabled={copying}
            >
              <Copy className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>

        {primaryHighlight && (
          <p className="mt-2 line-clamp-3 text-sm leading-relaxed" style={{ color: "var(--glass-text-muted)" }}>
            <HighlightedText html={primaryHighlight} />
          </p>
        )}

        {result.tags && result.tags.length > 0 && (
          <div className="mt-3 flex flex-wrap gap-1.5">
            {result.tags.map((tag) => (
              <Badge
                key={tag}
                variant="secondary"
                className="h-5 px-2 text-xs font-normal"
              >
                #{tag}
              </Badge>
            ))}
          </div>
        )}
      </div>

      <EntryPreviewModal
        entryId={previewOpen ? result.entry_id : null}
        onClose={() => setPreviewOpen(false)}
      />
    </>
  )
}
