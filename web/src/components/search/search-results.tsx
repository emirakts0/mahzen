import { AnimatePresence, motion } from "framer-motion"
import { BookText, Sparkles, SearchX } from "lucide-react"
import { KeywordResultCard } from "./keyword-result-card"
import { SemanticResultCard } from "./semantic-result-card"
import { SearchSkeleton } from "./search-skeleton"
import type { SearchResult } from "@/types/api"

// ─────────────────────────────────────────────
// Card fade-in — simple, no stagger, no layout shift
// ─────────────────────────────────────────────

function CardWrapper({ children, index }: { children: React.ReactNode; index: number }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 6 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2, ease: "easeOut", delay: index * 0.04 }}
    >
      {children}
    </motion.div>
  )
}

// ─────────────────────────────────────────────
// Column header
// ─────────────────────────────────────────────

function ColumnHeader({
  icon: Icon,
  title,
  count,
}: {
  icon: React.ElementType
  title: string
  count?: number
}) {
  return (
    <div className="flex items-center gap-2 pb-1">
      <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-primary/15 backdrop-blur-sm">
        <Icon className="h-4 w-4 text-primary" />
      </div>
      <span className="text-sm font-semibold uppercase tracking-widest text-muted-foreground">
        {title}
      </span>
      {count !== undefined && count > 0 && (
        <span className="ml-auto text-xs text-muted-foreground tabular-nums">
          {count} result{count !== 1 ? "s" : ""}
        </span>
      )}
    </div>
  )
}

// ─────────────────────────────────────────────
// Empty state
// ─────────────────────────────────────────────

function EmptyColumn({ label }: { label: string }) {
  return (
    <div className="flex flex-col items-center justify-center gap-2 rounded-xl border border-dashed border-border/50 bg-card/30 backdrop-blur-sm py-10 text-center">
      <SearchX className="h-8 w-8 text-muted-foreground/40" />
      <p className="text-sm text-muted-foreground">No {label} results</p>
    </div>
  )
}

// ─────────────────────────────────────────────
// Props
// ─────────────────────────────────────────────

interface SearchResultsProps {
  keywordResults: SearchResult[]
  semanticResults: SearchResult[]
  isKeywordLoading: boolean
  keywordError: Error | null
  semanticError: Error | null
  query: string
}

// ─────────────────────────────────────────────
// Main component
// ─────────────────────────────────────────────

export function SearchResults({
  keywordResults,
  semanticResults,
  isKeywordLoading,
  keywordError,
  semanticError,
  query,
}: SearchResultsProps) {
  const wordCount = query.trim().split(/\s+/).filter(Boolean).length

  if (isKeywordLoading) {
    return (
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        transition={{ duration: 0.15 }}
      >
        <SearchSkeleton count={3} />
      </motion.div>
    )
  }

  return (
    /*
     * Key on query so the entire grid fades out/in when the
     * query changes — no per-card AnimatePresence needed.
     */
    <AnimatePresence mode="wait">
      <motion.div
        key={query}
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        transition={{ duration: 0.18 }}
        className="grid grid-cols-1 gap-6 md:grid-cols-2"
      >
        {/* ── Keyword column ── */}
        <div className="flex flex-col gap-3">
          <ColumnHeader
            icon={BookText}
            title="Keyword Search"
            count={keywordResults.length}
          />

          {keywordError ? (
            <div className="rounded-xl border border-destructive/20 bg-destructive/5 p-4">
              <p className="text-sm text-destructive">
                Failed to load keyword results.
              </p>
            </div>
          ) : keywordResults.length === 0 ? (
            <EmptyColumn label="keyword" />
          ) : (
            <div className="flex flex-col gap-3">
              {keywordResults.map((result, i) => (
                <CardWrapper key={result.entry_id} index={i}>
                  <KeywordResultCard result={result} />
                </CardWrapper>
              ))}
            </div>
          )}
        </div>

        {/* ── Semantic column ── */}
        <div className="flex flex-col gap-3">
          <ColumnHeader
            icon={Sparkles}
            title="Semantic Search"
            count={semanticResults.length}
          />

          {wordCount < 2 ? (
            <div className="flex flex-col items-center justify-center gap-2 rounded-xl border border-dashed border-border/50 bg-card/30 backdrop-blur-sm py-10 text-center">
              <Sparkles className="h-8 w-8 text-muted-foreground/30" />
              <p className="text-sm text-muted-foreground">
                Type at least{" "}
                <span className="font-semibold text-foreground">2 words</span>{" "}
                for semantic search
              </p>
            </div>
          ) : semanticError ? (
            <div className="rounded-xl border border-destructive/20 bg-destructive/5 p-4">
              <p className="text-sm text-destructive">
                Failed to load semantic results.
              </p>
            </div>
          ) : semanticResults.length === 0 ? (
            <EmptyColumn label="semantic" />
          ) : (
            <div className="flex flex-col gap-3">
              {semanticResults.map((result, i) => (
                <CardWrapper key={result.entry_id} index={i}>
                  <SemanticResultCard result={result} />
                </CardWrapper>
              ))}
            </div>
          )}
        </div>
      </motion.div>
    </AnimatePresence>
  )
}
