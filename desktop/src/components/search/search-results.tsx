import { AnimatePresence, motion } from "framer-motion"
import { BookText, Sparkles, SearchX } from "lucide-react"
import { KeywordResultCard } from "./keyword-result-card"
import { SemanticResultCard } from "./semantic-result-card"
import { SearchLoading } from "./search-loading"
import type { SearchResult } from "@/types/api"
import type { SearchColumn } from "@/hooks/use-keyboard-nav"

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

function ColumnHeader({
  icon: Icon,
  title,
  count,
  isActive,
}: {
  icon: React.ElementType
  title: string
  count?: number
  isActive?: boolean
}) {
  return (
    <div className="flex items-center gap-1.5 pb-1">
      <div
        className="flex h-5 w-5 items-center justify-center rounded-lg backdrop-blur-sm transition-colors"
        style={{ 
          background: isActive ? "var(--color-primary)" : "var(--glass-hover)",
        }}
      >
        <Icon 
          className="h-3 w-3" 
          style={{ 
            color: isActive ? "var(--color-primary-foreground)" : "var(--glass-accent, var(--glass-text))" 
          }} 
        />
      </div>
      <span
        className="text-[10px] font-semibold uppercase tracking-widest transition-colors"
        style={{ color: isActive ? "var(--glass-text)" : "var(--glass-text-muted)" }}
      >
        {title}
      </span>
      {count !== undefined && count > 0 && (
        <span className="ml-auto text-[10px] tabular-nums" style={{ color: "var(--glass-text-muted)" }}>
          {count}
        </span>
      )}
    </div>
  )
}

function EmptyColumn({ label }: { label: string }) {
  return (
    <div
      className="flex flex-col items-center justify-center gap-1.5 rounded-lg border border-dashed py-6 text-center backdrop-blur-sm"
      style={{
        borderColor: "var(--glass-border)",
        background: "var(--glass-bg)",
      }}
    >
      <SearchX className="h-5 w-5" style={{ color: "var(--glass-text-muted)", opacity: 0.4 }} />
      <p className="text-[10px]" style={{ color: "var(--glass-text-muted)" }}>No {label} results</p>
    </div>
  )
}

interface SearchResultsProps {
  keywordResults: SearchResult[]
  semanticResults: SearchResult[]
  isKeywordLoading: boolean
  isSemanticLoading: boolean
  keywordError: Error | null
  semanticError: Error | null
  query: string
  selectedIndex?: number
  activeColumn?: SearchColumn
  selectedRef?: React.RefObject<HTMLDivElement | null>
}

export function SearchResults({
  keywordResults,
  semanticResults,
  isKeywordLoading,
  isSemanticLoading,
  keywordError,
  semanticError,
  query,
  selectedIndex = -1,
  activeColumn = "keyword",
  selectedRef,
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
    <AnimatePresence mode="wait">
      <motion.div
        key={query}
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: 10 }}
        transition={{ duration: 0.25, ease: "easeOut" }}
        className="grid grid-cols-1 gap-4 md:grid-cols-2 h-full"
        style={{ marginTop: "2rem" }}
      >
        <div
          className="flex flex-col gap-2 min-h-0"
        >
          <ColumnHeader
            icon={BookText}
            title="Keyword Search"
            count={keywordResults.length}
            isActive={activeColumn === "keyword"}
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
            <div className="flex flex-col gap-2 overflow-y-auto pr-1 flex-1 min-h-0">
              {keywordResults.map((result, i) => (
                <CardWrapper key={result.entry_id} index={i}>
                  <KeywordResultCard 
                    result={result} 
                    isSelected={activeColumn === "keyword" && selectedIndex === i}
                    ref={activeColumn === "keyword" && selectedIndex === i ? selectedRef : undefined}
                  />
                </CardWrapper>
              ))}
            </div>
          )}
        </div>

        <div
          className="flex flex-col gap-2 min-h-0"
        >
          <ColumnHeader
            icon={Sparkles}
            title="Semantic Search"
            count={semanticResults.length}
            isActive={activeColumn === "semantic"}
          />

          {wordCount < 2 ? (
            <div
              className="flex flex-col items-center justify-center gap-1.5 rounded-lg border border-dashed py-6 text-center backdrop-blur-sm"
              style={{
                borderColor: "var(--glass-border)",
                background: "var(--glass-bg)",
              }}
            >
              <Sparkles className="h-5 w-5" style={{ color: "var(--glass-text-muted)", opacity: 0.3 }} />
              <p className="text-[10px]" style={{ color: "var(--glass-text-muted)" }}>
                Type at least{" "}
                <span className="font-semibold" style={{ color: "var(--glass-text)" }}>2 words</span>
              </p>
            </div>
          ) : isSemanticLoading ? (
            <div className="flex justify-center py-6">
              <div className="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
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
            <div className="flex flex-col gap-2 overflow-y-auto pr-1 flex-1 min-h-0">
              {semanticResults.map((result, i) => (
                <CardWrapper key={result.entry_id} index={i}>
                  <SemanticResultCard 
                    result={result} 
                    isSelected={activeColumn === "semantic" && selectedIndex === i}
                    ref={activeColumn === "semantic" && selectedIndex === i ? selectedRef : undefined}
                  />
                </CardWrapper>
              ))}
            </div>
          )}
        </div>
      </motion.div>
    </AnimatePresence>
  )
}
