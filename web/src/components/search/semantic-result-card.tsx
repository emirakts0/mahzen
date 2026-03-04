import { useState } from "react"
import { Link } from "react-router"
import { Copy, ExternalLink, ChevronRight } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import type { SearchResult } from "@/types/api"
import { toast } from "sonner"

interface SemanticResultCardProps {
  result: SearchResult
  className?: string
}

function PathBreadcrumb({ path }: { path: string }) {
  if (!path || path === "/") return null
  const parts = path.replace(/^\//, "").split("/").filter(Boolean)
  return (
    <div className="flex flex-wrap items-center gap-0.5 text-xs text-muted-foreground">
      {parts.map((part, i) => (
        <span key={i} className="flex items-center gap-0.5">
          {i > 0 && <ChevronRight className="h-3 w-3 opacity-50" />}
          <span className="font-medium uppercase tracking-wide">{part}</span>
        </span>
      ))}
    </div>
  )
}

function MatchBadge({ score }: { score: number }) {
  const percent = Math.round(score * 100)

  const colorClass =
    percent >= 80
      ? "bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400"
      : percent >= 60
        ? "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400"
        : "bg-muted text-muted-foreground"

  return (
    <span
      className={cn(
        "shrink-0 rounded-full px-2 py-0.5 text-xs font-semibold tabular-nums",
        colorClass,
      )}
    >
      {percent}% Match
    </span>
  )
}

export function SemanticResultCard({ result, className }: SemanticResultCardProps) {
  const [copying, setCopying] = useState(false)

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

  return (
    <div
      className={cn(
        "group relative rounded-xl border border-border/60 bg-card/75 backdrop-blur-sm p-4 shadow-sm transition-all",
        "hover:border-primary/40 hover:shadow-md hover:bg-card/90",
        className,
      )}
    >
      {/* Header */}
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <Link
            to={`/entries/${result.entry_id}`}
            className="truncate text-sm font-semibold text-foreground hover:text-primary transition-colors"
          >
            {result.title || "Untitled"}
          </Link>
          <PathBreadcrumb path={result.path} />
        </div>

        {/* Match badge + action buttons */}
        <div className="flex shrink-0 items-center gap-2">
          {result.score > 0 && <MatchBadge score={result.score} />}

          <div className="flex items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={() => void handleCopy()}
              title="Copy snippet"
              disabled={copying}
            >
              <Copy className="h-3.5 w-3.5" />
            </Button>
            <Button variant="ghost" size="icon" className="h-7 w-7" asChild title="Open entry">
              <Link to={`/entries/${result.entry_id}`}>
                <ExternalLink className="h-3.5 w-3.5" />
              </Link>
            </Button>
          </div>
        </div>
      </div>

      {/* Snippet */}
      {result.snippet && (
        <p className="mt-2 line-clamp-3 text-sm leading-relaxed text-muted-foreground">
          {result.snippet}
        </p>
      )}

      {/* Tags */}
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
  )
}
