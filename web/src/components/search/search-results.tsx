import { AnimatePresence, motion } from "framer-motion"
import { BookText, Sparkles, SearchX } from "lucide-react"
import { KeywordResultCard } from "./keyword-result-card"
import { SemanticResultCard } from "./semantic-result-card"
import { SearchLoading } from "./search-loading"
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
      <div
        className="flex h-7 w-7 items-center justify-center rounded-lg backdrop-blur-sm"
        style={{ background: "var(--glass-hover)" }}
      >
        <Icon className="h-4 w-4" style={{ color: "var(--glass-accent, var(--glass-text))" }} />
      </div>
      <span
        className="text-sm font-semibold uppercase tracking-widest"
        style={{ color: "var(--glass-text-muted)" }}
      >
        {title}
      </span>
      {count !== undefined && count > 0 && (
        <span className="ml-auto text-xs tabular-nums" style={{ color: "var(--glass-text-muted)" }}>
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
    <div
      className="flex flex-col items-center justify-center gap-2 rounded-xl border border-dashed py-10 text-center backdrop-blur-sm"
      style={{
        borderColor: "var(--glass-border)",
        background: "var(--glass-bg)",
      }}
    >
      <SearchX className="h-8 w-8" style={{ color: "var(--glass-text-muted)", opacity: 0.4 }} />
      <p className="text-sm" style={{ color: "var(--glass-text-muted)" }}>No {label} results</p>
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
        <SearchLoading />
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
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: 10 }}
        transition={{ duration: 0.25, ease: "easeOut" }}
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
            <div
              className="rounded-xl border p-4"
              style={{
                borderColor: "var(--glass-error, #ef4444)",
                background: "rgba(239, 68, 68, 0.1)",
              }}
            >
              <p className="text-sm" style={{ color: "var(--glass-error, #ef4444)" }}>
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
            <div
              className="flex flex-col items-center justify-center gap-2 rounded-xl border border-dashed py-10 text-center backdrop-blur-sm"
              style={{
                borderColor: "var(--glass-border)",
                background: "var(--glass-bg)",
              }}
            >
              <Sparkles className="h-8 w-8" style={{ color: "var(--glass-text-muted)", opacity: 0.3 }} />
              <p className="text-sm" style={{ color: "var(--glass-text-muted)" }}>
                Type at least{" "}
                <span className="font-semibold" style={{ color: "var(--glass-text)" }}>2 words</span>{" "}
                for semantic search
              </p>
            </div>
          ) : semanticError ? (
            <div
              className="rounded-xl border p-4"
              style={{
                borderColor: "var(--glass-error, #ef4444)",
                background: "rgba(239, 68, 68, 0.1)",
              }}
            >
              <p className="text-sm" style={{ color: "var(--glass-error, #ef4444)" }}>
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
