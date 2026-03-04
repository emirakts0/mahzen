import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { Copy, ExternalLink, ChevronRight, Calendar, WrapText, ArrowLeftRight } from "lucide-react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { getEntry } from "@/api/entries"
import { toast } from "sonner"

interface EntryPreviewModalProps {
  entryId: string | null
  onClose: () => void
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

export function EntryPreviewModal({ entryId, onClose }: EntryPreviewModalProps) {
  const [copied, setCopied] = useState(false)
  const [wrap, setWrap] = useState(true)

  const { data: entry, isLoading, error } = useQuery({
    queryKey: ["entry", entryId],
    queryFn: () => getEntry(entryId!),
    enabled: !!entryId,
  })

  const handleCopy = async () => {
    if (!entry) return
    try {
      await navigator.clipboard.writeText(entry.content)
      setCopied(true)
      toast.success("Content copied")
      setTimeout(() => setCopied(false), 2000)
    } catch {
      toast.error("Failed to copy")
    }
  }

  if (!entryId) return null

  return (
    <Dialog open={!!entryId} onOpenChange={(open) => !open && onClose()}>
      <DialogContent
        className="max-h-[80vh] overflow-y-auto"
        style={{
          background: "var(--glass-bg)",
          border: "1px solid var(--glass-border)",
          color: "var(--glass-text)",
          maxWidth: "800px",
        }}
      >
        <DialogHeader>
          <DialogTitle
            style={{ color: "var(--glass-text)" }}
            className="text-lg leading-none font-semibold"
          >
            {isLoading ? "Loading..." : entry?.title || "Untitled"}
          </DialogTitle>
        </DialogHeader>

        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <div
              className="h-6 w-6 animate-spin rounded-full border-2"
              style={{ borderColor: "var(--glass-border)", borderTopColor: "transparent" }}
            />
          </div>
        ) : error ? (
          <p style={{ color: "var(--glass-error, #ef4444)" }}>Failed to load entry</p>
        ) : entry ? (
          <div className="flex flex-col gap-4">
            <PathBreadcrumb path={entry.path} />

            <div className="flex flex-wrap gap-3 text-xs" style={{ color: "var(--glass-text-muted)" }}>
              <span className="flex items-center gap-1">
                <Calendar className="h-3 w-3" />
                {new Date(entry.created_at).toLocaleDateString()}
              </span>
              <span
                className="flex items-center gap-1 rounded-full px-2 py-0.5"
                style={{
                  background: entry.visibility === "public" ? "rgba(34, 197, 94, 0.15)" : "rgba(234, 179, 8, 0.15)",
                  color: entry.visibility === "public" ? "#22c55e" : "#eab308",
                }}
              >
                {entry.visibility}
              </span>
            </div>

            <div
              className="rounded-lg p-4 text-sm leading-relaxed"
              style={{
                background: "var(--glass-hover)",
                color: "var(--glass-text-muted)",
                whiteSpace: wrap ? "pre-wrap" : "pre",
                overflowX: wrap ? "visible" : "auto",
              }}
            >
              {entry.content}
            </div>

            <div className="flex items-center justify-end gap-2 pt-2">
              <button
                onClick={() => setWrap(!wrap)}
                className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs transition-colors"
                style={{
                  background: "var(--glass-hover)",
                  color: "var(--glass-text)",
                }}
              >
                {wrap ? <WrapText className="h-3.5 w-3.5" /> : <ArrowLeftRight className="h-3.5 w-3.5" />}
                {wrap ? "Wrap" : "No Wrap"}
              </button>
            </div>

            {entry.tags && entry.tags.length > 0 && (
              <div className="flex flex-wrap gap-1.5">
                {entry.tags.map((tag) => (
                  <Badge
                    key={tag}
                    variant="secondary"
                    className="h-5 px-2 text-xs font-normal"
                    style={{
                      background: "var(--glass-hover)",
                      color: "var(--glass-text-muted)",
                    }}
                  >
                    #{tag}
                  </Badge>
                ))}
              </div>
            )}

            <div className="flex gap-2 pt-2">
              <button
                onClick={handleCopy}
                className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs transition-colors"
                style={{
                  background: "var(--glass-hover)",
                  color: "var(--glass-text)",
                }}
              >
                <Copy className="h-3.5 w-3.5" />
                {copied ? "Copied!" : "Copy"}
              </button>
              <a
                href={`/entries/${entry.id}`}
                className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs transition-colors"
                style={{
                  background: "var(--glass-hover)",
                  color: "var(--glass-text)",
                }}
              >
                <ExternalLink className="h-3.5 w-3.5" />
                Open in full
              </a>
            </div>
          </div>
        ) : null}
      </DialogContent>
    </Dialog>
  )
}
