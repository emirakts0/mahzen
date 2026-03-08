import { useState, useRef, useCallback, useEffect } from "react"
import { useQuery } from "@tanstack/react-query"
import { motion, AnimatePresence } from "framer-motion"
import { SearchInput } from "@/components/search/search-input"
import { SearchResults } from "@/components/search/search-results"
import { SearchFilters } from "@/components/search/search-filters"
import { useDebounce } from "@/hooks/use-debounce"
import { keywordSearch, semanticSearch } from "@/api/search"
import { useAuth } from "@/hooks/use-auth"
import type { SearchResult, SearchFilters as SearchFiltersType } from "@/types/api"

const DEBOUNCE_MS = 350
const MIN_KEYWORD_LENGTH = 1
const MIN_SEMANTIC_WORDS = 2

export default function SearchPage() {
  const { isAuthenticated, isLoading: authLoading } = useAuth()
  const [query, setQuery] = useState("")
  const [filters, setFilters] = useState<SearchFiltersType>({})
  const [isFilterOpen, setIsFilterOpen] = useState(false)
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
    queryKey: ["search", "keyword", trimmedQuery, filters],
    queryFn: () => keywordSearch({ query: trimmedQuery, limit: 10, ...filters }),
    enabled: isAuthenticated && isQueryReady,
    placeholderData: (prev) => prev,
    staleTime: 1000 * 15,
  })

  const {
    data: semanticData,
    isFetching: isSemanticFetching,
    error: semanticError,
  } = useQuery({
    queryKey: ["search", "semantic", trimmedQuery, filters],
    queryFn: () => semanticSearch({ query: trimmedQuery, limit: 10, ...filters }),
    enabled: isAuthenticated && isSemanticReady,
    placeholderData: (prev) => prev,
    staleTime: 1000 * 15,
  })

  const handleChange = useCallback((val: string) => setQuery(val), [])
  const handleClear = useCallback(() => {
    setQuery("")
    setFilters({})
    inputRef.current?.focus()
  }, [])

  const filterCount =
    (filters.tags?.length ?? 0) +
    (filters.path ? 1 : 0) +
    (filters.from_date ? 1 : 0) +
    (filters.to_date ? 1 : 0) +
    (filters.only_mine ? 1 : 0) +
    (filters.visibility ? 1 : 0)

  // Track whether the hero search bar has scrolled out of view
  const heroSearchRef = useRef<HTMLDivElement>(null)
  const isSearchStickyRef = useRef(false)

  useEffect(() => {
    const check = () => {
      const el = heroSearchRef.current
      if (!el) return
      const rect = el.getBoundingClientRect()
      // The header is at top: 1rem (16px).
      // We want to trigger when the search bar's top reaches 16px.
      const isSticky = rect.top <= 16

      if (isSticky !== isSearchStickyRef.current) {
        isSearchStickyRef.current = isSticky
        // Dispatch event to tell header to hide
        window.dispatchEvent(new CustomEvent("search-sticky", { detail: isSticky }))
      }
    }
    window.addEventListener("scroll", check, { passive: true })
    check() // initial check
    return () => {
      window.removeEventListener("scroll", check)
      window.dispatchEvent(new CustomEvent("search-sticky", { detail: false })) // reset on unmount
    }
  }, [])

  const keywordResults: SearchResult[] = keywordData?.results ?? []
  const semanticResults: SearchResult[] = semanticData?.results ?? []

  const showResults = isAuthenticated && trimmedQuery.length > 0

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
        minHeight: "calc(100vh - 5rem)",
      }}
    >
      {/* Hero — fixed height, centers content */}
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          padding: "4rem 1rem 1rem",
          flexShrink: 0,
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
      </div>

      {/* Sticky Search bar — naturally follows hero, sticks to top when scrolling */}
      <div
        ref={heroSearchRef}
        style={{
          position: "sticky",
          top: "1rem", // Stops exactly where the header was
          zIndex: 40,
          width: "100%",
          display: "flex",
          justifyContent: "center",
          padding: "0 1rem 2rem", // Padding acts as a visual spacer below
          pointerEvents: "none", // Let clicks pass through the padding area to results
        }}
      >
        <div style={{ width: "100%", maxWidth: "640px", pointerEvents: "auto" }}>
          {isAuthenticated ? (
            <div className="relative">
              <SearchInput
                ref={inputRef}
                value={query}
                onChange={handleChange}
                onClear={handleClear}
                autoFocus
                filterCount={filterCount}
                onFilterClick={() => setIsFilterOpen(!isFilterOpen)}
              />
              <SearchFilters
                filters={filters}
                onFiltersChange={setFilters}
                isOpen={isFilterOpen}
                onClose={() => setIsFilterOpen(false)}
              />
            </div>
          ) : (
            <SearchInput
              value=""
              onChange={() => undefined}
              disabled
              placeholder="Sign in to search your knowledge base..."
            />
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
              isKeywordLoading={isKeywordFetching}
              isSemanticLoading={isSemanticFetching}
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
