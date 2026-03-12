import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import {
  Copy,
  ChevronRight,
  WrapText,
  ArrowLeftRight,
  X,
  Clock,
  HardDrive,
} from "lucide-react"
import {
  Dialog,
  DialogContent,
} from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { getEntry } from "@/api/entries"
import { toast } from "sonner"
import { getFileTypeLabel, formatFileSize } from "../entries/file-icons"

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
    if (!entry?.content) return
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
        showCloseButton={false}
        className="flex max-h-[85vh] flex-col overflow-hidden p-0 sm:p-0 gap-0"
        style={{
          background: "var(--glass-bg)",
          border: "1px solid var(--glass-border)",
          color: "var(--glass-text)",
          maxWidth: "900px",
          backdropFilter: "blur(20px)",
        }}
      >
        <div
          className="flex shrink-0 items-center justify-between px-6 pt-5 pb-4 border-b"
          style={{ borderColor: "var(--glass-border)" }}
        >
          <div className="flex-1 min-w-0 mr-4">
            <h2
              style={{ color: "var(--glass-text)" }}
              className="text-lg font-semibold truncate"
            >
              {isLoading ? "Loading..." : entry?.title || "Untitled"}
            </h2>
            {entry && <PathBreadcrumb path={entry.path} />}
          </div>
          <button
            onClick={onClose}
            className="rounded-md p-1.5 transition-colors hover:opacity-70 shrink-0"
            style={{ color: "var(--glass-text-muted)" }}
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="overflow-y-auto px-6 py-5">
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <div
                className="h-8 w-8 animate-spin rounded-full border-2"
                style={{ borderColor: "var(--glass-border)", borderTopColor: "var(--color-primary)" }}
              />
            </div>
          ) : error ? (
            <p style={{ color: "var(--glass-error, #ef4444)" }}>Failed to load entry</p>
          ) : entry ? (
            <div className="flex flex-col gap-5">
              <div className="flex flex-wrap gap-4 text-xs" style={{ color: "var(--glass-text-muted)" }}>
                <span className="flex items-center gap-1.5 px-2 py-1 rounded-md" style={{ background: "var(--glass-hover)" }}>
                  <Clock className="h-3.5 w-3.5" />
                  {new Date(entry.updated_at).toLocaleDateString("tr-TR", {
                    day: "numeric",
                    month: "short",
                    year: "numeric",
                  })}
                </span>

                <span
                  className="flex items-center gap-1.5 rounded-full px-2.5 py-1 font-medium"
                  style={{
                    background: entry.visibility === "public" ? "rgba(34, 197, 94, 0.15)" : "rgba(234, 179, 8, 0.15)",
                    color: entry.visibility === "public" ? "#22c55e" : "#eab308",
                  }}
                >
                  {entry.visibility}
                </span>

                {entry.file_type && (
                  <span className="flex items-center gap-1.5 px-2 py-1 rounded-md font-mono" style={{ background: "var(--glass-hover)" }}>
                    {getFileTypeLabel(entry.file_type)}
                  </span>
                )}

                {entry.file_size && entry.file_size > 0 && (
                  <span className="flex items-center gap-1.5 px-2 py-1 rounded-md" style={{ background: "var(--glass-hover)" }}>
                    <HardDrive className="h-3.5 w-3.5" />
                    {formatFileSize(entry.file_size)}
                  </span>
                )}
              </div>

              {entry.summary && (
                <div className="text-sm italic px-3 py-2 rounded-lg border-l-2" style={{ borderColor: "var(--color-primary)", color: "var(--glass-text-muted)", background: "var(--glass-hover)" }}>
                  {entry.summary}
                </div>
              )}

              <div className="relative">
                <div
                  className="selectable rounded-lg p-4 text-sm leading-relaxed font-mono"
                  style={{
                    background: "var(--glass-hover)",
                    color: "var(--glass-text)",
                    whiteSpace: wrap ? "pre-wrap" : "pre",
                    overflowX: wrap ? "visible" : "auto",
                    maxHeight: "400px",
                    overflowY: "auto",
                    paddingTop: entry.content ? "3rem" : undefined,
                  }}
                >
                  {entry.content ? (
                    entry.content
                  ) : (
                    <span style={{ color: "var(--glass-text-muted)", fontStyle: "italic" }}>No content available</span>
                  )}
                </div>
                
                {entry.content && (
                  <div className="absolute top-2 right-2 flex items-center gap-1.5">
                    <button
                      onClick={() => setWrap(!wrap)}
                      className="flex items-center gap-1 rounded-lg px-2 py-1 text-xs transition-colors hover:opacity-80"
                      style={{
                        background: "var(--glass-bg)",
                        color: "var(--glass-text-muted)",
                        border: "1px solid var(--glass-border)",
                      }}
                    >
                      {wrap ? <WrapText className="h-3 w-3" /> : <ArrowLeftRight className="h-3 w-3" />}
                      {wrap ? "Wrap" : "No Wrap"}
                    </button>
                    <button
                      onClick={handleCopy}
                      className="flex items-center gap-1 rounded-lg px-2 py-1 text-xs transition-colors hover:opacity-80"
                      style={{
                        background: "var(--glass-bg)",
                        color: copied ? "#22c55e" : "var(--glass-text-muted)",
                        border: "1px solid var(--glass-border)",
                      }}
                    >
                      <Copy className="h-3 w-3" />
                      {copied ? "Copied!" : "Copy"}
                    </button>
                  </div>
                )}
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
            </div>
          ) : null}
        </div>
      </DialogContent>
    </Dialog>
  )
}
