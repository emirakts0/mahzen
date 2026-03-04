import { useState, useRef, useCallback } from "react"
import { useQuery } from "@tanstack/react-query"
import { motion, AnimatePresence } from "framer-motion"
import { SearchInput } from "@/components/search/search-input"
import { SearchResults } from "@/components/search/search-results"
import { useDebounce } from "@/hooks/use-debounce"
import { keywordSearch, semanticSearch } from "@/api/search"
import { useAuth } from "@/hooks/use-auth"
import { Button } from "@/components/ui/button"
import { Link } from "react-router"
import { Lock } from "lucide-react"
import type { SearchResult } from "@/types/api"

const DEBOUNCE_MS = 350
const MIN_KEYWORD_LENGTH = 1
const MIN_SEMANTIC_WORDS = 2

export default function SearchPage() {
  const { isAuthenticated, isLoading: authLoading } = useAuth()
  const [query, setQuery] = useState("")
  const inputRef = useRef<HTMLInputElement>(null)

  const debouncedQuery = useDebounce(query, DEBOUNCE_MS)

  const trimmedQuery = debouncedQuery.trim()
  const wordCount = trimmedQuery.split(/\s+/).filter(Boolean).length
  const isQueryReady = trimmedQuery.length >= MIN_KEYWORD_LENGTH
  const isSemanticReady = wordCount >= MIN_SEMANTIC_WORDS

  const {
    data: keywordData,
    isFetching: isKeywordFetching,
    error: keywordError,
  } = useQuery({
    queryKey: ["search", "keyword", trimmedQuery],
    queryFn: () => keywordSearch({ query: trimmedQuery, limit: 10 }),
    enabled: isAuthenticated && isQueryReady,
    placeholderData: (prev) => prev,
    staleTime: 1000 * 15,
  })

  const {
    data: semanticData,
    isFetching: isSemanticFetching,
    error: semanticError,
  } = useQuery({
    queryKey: ["search", "semantic", trimmedQuery],
    queryFn: () => semanticSearch({ query: trimmedQuery, limit: 10 }),
    enabled: isAuthenticated && isSemanticReady,
    placeholderData: (prev) => prev,
    staleTime: 1000 * 15,
  })

  const handleChange = useCallback((val: string) => setQuery(val), [])
  const handleClear = useCallback(() => {
    setQuery("")
    inputRef.current?.focus()
  }, [])

  const keywordResults: SearchResult[] = keywordData?.results ?? []
  const semanticResults: SearchResult[] = semanticData?.results ?? []

  const showResults = isAuthenticated && trimmedQuery.length > 0
  const isLoading = isKeywordFetching || isSemanticFetching

  const hint =
    query.trim().length > 0 && wordCount < MIN_SEMANTIC_WORDS
      ? "Type at least 2 words to enable semantic search"
      : null

  if (authLoading) {
    return (
      <div className="flex h-[calc(100vh-5rem)] items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    )
  }

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        height: "calc(100vh - 5rem)",
        overflowY: "auto",
      }}
    >
      {/* Hero — takes all remaining height, centers content */}
      <div
        style={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          padding: "0 1rem 1rem",
        }}
      >
        {/* Title */}
        <div style={{ textAlign: "center", marginBottom: "2rem" }}>
          <h1
            style={{
              fontSize: "3rem",
              fontWeight: 700,
              letterSpacing: "-0.02em",
              color: "var(--glass-text)",
              margin: 0,
            }}
          >
            mahzen
          </h1>
          <p
            style={{
              marginTop: "0.5rem",
              fontSize: "0.875rem",
              color: "var(--glass-text-muted)",
            }}
          >
            Discover insights across your knowledge base
          </p>
        </div>

        {/* Search bar — fixed width, centered */}
        <div style={{ width: "100%", maxWidth: "640px" }}>
          {isAuthenticated ? (
            <SearchInput
              ref={inputRef}
              value={query}
              onChange={handleChange}
              onClear={handleClear}
              hint={hint}
              autoFocus
            />
          ) : (
            <>
              <SearchInput
                value=""
                onChange={() => undefined}
                disabled
                placeholder="Sign in to search your knowledge base..."
              />
              <div
                style={{
                  marginTop: "1rem",
                  display: "flex",
                  flexDirection: "column",
                  alignItems: "center",
                  gap: "0.75rem",
                }}
              >
                <div
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: "0.5rem",
                    fontSize: "0.875rem",
                    color: "var(--glass-text-muted)",
                  }}
                >
                  <Lock style={{ width: "1rem", height: "1rem" }} />
                  <span>Search requires authentication</span>
                </div>
                <div style={{ display: "flex", gap: "0.75rem" }}>
                  <Button
                    asChild
                    style={{
                      background: "var(--glass-bg)",
                      border: "1px solid var(--glass-border)",
                      color: "var(--glass-text)",
                    }}
                  >
                    <Link to="/login">Sign in</Link>
                  </Button>
                  <Button
                    variant="outline"
                    asChild
                    style={{
                      background: "var(--glass-bg)",
                      border: "1px solid var(--glass-border)",
                      color: "var(--glass-text)",
                    }}
                  >
                    <Link to="/signup">Create account</Link>
                  </Button>
                </div>
              </div>
            </>
          )}
        </div>
      </div>

      {/* Results */}
      <AnimatePresence>
        {showResults && (
          <motion.div
            key="results"
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.25, ease: "easeOut" }}
            style={{
              width: "100%",
              maxWidth: "72rem",
              margin: "0 auto",
              padding: "0 1rem 3rem",
            }}
          >
            <SearchResults
              keywordResults={keywordResults}
              semanticResults={semanticResults}
              isKeywordLoading={isLoading}
              keywordError={keywordError instanceof Error ? keywordError : null}
              semanticError={semanticError instanceof Error ? semanticError : null}
              query={trimmedQuery}
            />
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
