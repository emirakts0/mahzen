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
  Lock,
  Globe,
  User,
} from "lucide-react"
import { writeText } from "@tauri-apps/plugin-clipboard-manager"
import {
  Dialog,
  DialogContent,
} from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { getEntry } from "@/api/entries"
import { useAuth } from "@/providers/auth-provider"
import { toast } from "sonner"

function getFileTypeLabel(fileType?: string): string {
  if (!fileType) return "Text"

  const type = fileType.toLowerCase()

  const labels: Record<string, string> = {
    go: "Go",
    ts: "TypeScript",
    tsx: "TypeScript React",
    js: "JavaScript",
    jsx: "JavaScript React",
    py: "Python",
    rs: "Rust",
    java: "Java",
    c: "C",
    cpp: "C++",
    css: "CSS",
    html: "HTML",
    json: "JSON",
    yaml: "YAML",
    yml: "YAML",
    toml: "TOML",
    sh: "Shell",
    bash: "Bash",
    mp4: "MP4 Video",
    avi: "AVI Video",
    mov: "QuickTime",
    mkv: "Matroska",
    webm: "WebM",
    png: "PNG Image",
    jpg: "JPEG Image",
    jpeg: "JPEG Image",
    gif: "GIF Image",
    webp: "WebP Image",
    svg: "SVG Image",
    zip: "ZIP Archive",
    tar: "TAR Archive",
    gz: "GZIP Archive",
    rar: "RAR Archive",
    "7z": "7-Zip Archive",
    xlsx: "Excel Spreadsheet",
    xls: "Excel Spreadsheet",
    csv: "CSV Data",
    pdf: "PDF Document",
    doc: "Word Document",
    docx: "Word Document",
    txt: "Plain Text",
    md: "Markdown",
  }

  return labels[type] || type.toUpperCase()
}

function formatFileSize(bytes?: number): string {
  if (!bytes) return ""

  const units = ["B", "KB", "MB", "GB"]
  let size = bytes
  let unitIndex = 0

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }

  return `${size.toFixed(unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

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
  const { user } = useAuth()

  const { data: entry, isLoading, error } = useQuery({
    queryKey: ["entry", entryId],
    queryFn: () => getEntry(entryId!),
    enabled: !!entryId,
  })

  const isOwner = user?.id === entry?.user_id

  const handleCopy = async () => {
    if (!entry?.content) return
    try {
      await writeText(entry.content)
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
        className="flex max-h-[75vh] w-[90vw] max-w-[700px] flex-col overflow-hidden p-0 sm:p-0 gap-0"
        style={{
          background: "var(--glass-bg)",
          border: "1px solid var(--glass-border)",
          color: "var(--glass-text)",
          backdropFilter: "blur(20px)",
        }}
      >
        <div
          className="flex shrink-0 items-center justify-between px-4 pt-4 pb-3 border-b"
          style={{ borderColor: "var(--glass-border)" }}
        >
          <div className="flex-1 min-w-0 mr-3">
            <h2
              style={{ color: "var(--glass-text)" }}
              className="text-base font-semibold truncate"
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

        <div className="overflow-y-auto px-4 py-4">
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <div
                className="h-6 w-6 animate-spin rounded-full border-2"
                style={{ borderColor: "var(--glass-border)", borderTopColor: "var(--color-primary)" }}
              />
            </div>
          ) : error ? (
            <p className="text-sm" style={{ color: "var(--glass-error, #ef4444)" }}>Failed to load entry</p>
          ) : entry ? (
            <div className="flex flex-col gap-3">
              <div className="flex flex-wrap gap-2 text-[10px]" style={{ color: "var(--glass-text-muted)" }}>
                <span className="flex items-center gap-1 px-1.5 py-0.5 rounded-md" style={{ background: "var(--glass-hover)" }}>
                  <Clock className="h-3 w-3" />
                  {new Date(entry.updated_at).toLocaleDateString("tr-TR", {
                    day: "numeric",
                    month: "short",
                    year: "numeric",
                  })}
                </span>

                {isOwner && (
                  <span
                    className="flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-medium"
                    style={{
                      background: "rgba(59, 130, 246, 0.15)",
                      color: "#3b82f6",
                    }}
                  >
                    <User className="h-3 w-3" />
                    Mine
                  </span>
                )}

                <span
                  className="flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-medium"
                  style={{
                    background: entry.visibility === "public" ? "rgba(34, 197, 94, 0.15)" : "rgba(234, 179, 8, 0.15)",
                    color: entry.visibility === "public" ? "#22c55e" : "#eab308",
                  }}
                >
                  {entry.visibility === "public" ? (
                    <>
                      <Globe className="h-3 w-3" />
                      Public
                    </>
                  ) : (
                    <>
                      <Lock className="h-3 w-3" />
                      Private
                    </>
                  )}
                </span>

                {entry.file_type && (
                  <span className="flex items-center gap-1 px-1.5 py-0.5 rounded-md font-mono text-[10px]" style={{ background: "var(--glass-hover)" }}>
                    {getFileTypeLabel(entry.file_type)}
                  </span>
                )}

                {entry.file_size && entry.file_size > 0 && (
                  <span className="flex items-center gap-1 px-1.5 py-0.5 rounded-md" style={{ background: "var(--glass-hover)" }}>
                    <HardDrive className="h-3 w-3" />
                    {formatFileSize(entry.file_size)}
                  </span>
                )}
              </div>

              {entry.summary && (
                <div className="text-xs italic px-2.5 py-1.5 rounded-lg border-l-2" style={{ borderColor: "var(--color-primary)", color: "var(--glass-text-muted)", background: "var(--glass-hover)" }}>
                  {entry.summary}
                </div>
              )}

              <div className="relative">
                <div
                  className="selectable rounded-lg p-3 text-xs leading-relaxed font-mono"
                  style={{
                    background: "var(--glass-hover)",
                    color: "var(--glass-text)",
                    whiteSpace: wrap ? "pre-wrap" : "pre",
                    overflowX: wrap ? "visible" : "auto",
                    maxHeight: "250px",
                    overflowY: "auto",
                    paddingTop: entry.content ? "2.5rem" : undefined,
                  }}
                >
                  {entry.content ? (
                    entry.content
                  ) : (
                    <span style={{ color: "var(--glass-text-muted)", fontStyle: "italic" }}>No content available</span>
                  )}
                </div>

                {entry.content && (
                  <div className="absolute top-1.5 right-1.5 flex items-center gap-1">
                    <button
                      onClick={() => setWrap(!wrap)}
                      className="flex items-center gap-1 rounded-lg px-1.5 py-1 text-[10px] transition-colors hover:opacity-80"
                      style={{
                        background: "var(--glass-bg)",
                        color: "var(--glass-text-muted)",
                        border: "1px solid var(--glass-border)",
                      }}
                    >
                      {wrap ? <WrapText className="h-2.5 w-2.5" /> : <ArrowLeftRight className="h-2.5 w-2.5" />}
                      {wrap ? "Wrap" : "No Wrap"}
                    </button>
                    <button
                      onClick={handleCopy}
                      className="flex items-center gap-1 rounded-lg px-1.5 py-1 text-[10px] transition-colors hover:opacity-80"
                      style={{
                        background: "var(--glass-bg)",
                        color: copied ? "#22c55e" : "var(--glass-text-muted)",
                        border: "1px solid var(--glass-border)",
                      }}
                    >
                      <Copy className="h-2.5 w-2.5" />
                      {copied ? "Copied!" : "Copy"}
                    </button>
                  </div>
                )}
              </div>

              {entry.tags && entry.tags.length > 0 && (
                <div className="flex flex-wrap gap-1">
                  {entry.tags.map((tag) => (
                    <Badge
                      key={tag}
                      variant="secondary"
                      className="h-4 px-1.5 text-[10px] font-normal"
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
