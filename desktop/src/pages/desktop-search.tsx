import { useState, useRef, useCallback, useEffect, useMemo } from "react"
import { useQuery } from "@tanstack/react-query"
import { getCurrentWindow } from "@tauri-apps/api/window"
import { motion, AnimatePresence } from "framer-motion"
import { invoke } from "@tauri-apps/api/core"
import { listen } from "@tauri-apps/api/event"
import { Plus, BookOpen, Filter, X } from "lucide-react"
import { DesktopHeader } from "@/components/desktop-header"
import { SettingsPanel } from "@/components/settings-panel"
import { SearchInput } from "@/components/search/search-input"
import { SearchFilters, countActiveFilters } from "@/components/search/search-filters"
import { SearchResults } from "@/components/search/search-results"
import { AuthModal } from "@/components/auth/auth-modal"
import { FolderTree, CreateEntryDialog } from "@/components/entries"
import { EntryPreviewModal } from "@/components/search/entry-preview-modal"
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { useDebounce } from "@/hooks/use-debounce"
import { useKeyboardNav, type SearchColumn } from "@/hooks/use-keyboard-nav"
import { keywordSearch, semanticSearch } from "@/api/search"
import { listEntries } from "@/api/entries"
import { clearBaseUrlCache } from "@/api/client"
import { useAuth } from "@/providers/auth-provider"
import type { SearchFilters as SearchFiltersType, ListEntriesParams, SearchResult } from "@/types/api"

const DEBOUNCE_MS = 350
const MIN_KEYWORD_LENGTH = 1
const MIN_SEMANTIC_WORDS = 2

interface AppConfig {
  backend_url: string
}

type ViewMode = "search" | "entries"

export function DesktopSearch() {
  const { isAuthenticated, isLoading: authLoading, connect } = useAuth()
  const [viewMode, setViewMode] = useState<ViewMode>("search")
  const [query, setQuery] = useState("")
  const [filters, setFilters] = useState<SearchFiltersType>({})
  const [isFilterOpen, setIsFilterOpen] = useState(false)
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [authModalOpen, setAuthModalOpen] = useState(false)
  const [configLoaded, setConfigLoaded] = useState(false)
  const [backendUrl, setBackendUrl] = useState<string | null>(null)
  const [previewEntryId, setPreviewEntryId] = useState<string | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const [activeColumn, setActiveColumn] = useState<SearchColumn>("keyword")
  const [isSearchActive, setIsSearchActive] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const heroSearchRef = useRef<HTMLDivElement>(null)
  const scrollContainerRef = useRef<HTMLDivElement>(null)
  const isSearchStickyRef = useRef(false)

  const handleClearSearch = useCallback(() => {
    setQuery("")
    setFilters({})
    inputRef.current?.focus()
  }, [])

  const handleClose = useCallback(async () => {
    await getCurrentWindow().hide()
  }, [])

  const debouncedQuery = useDebounce(query, DEBOUNCE_MS)

  useEffect(() => {
    setIsSearchActive(query.length > 0)
  }, [query])

  useEffect(() => {
    const unlisten = listen("tauri://focus", () => {
      setTimeout(() => inputRef.current?.focus(), 50)
    })
    return () => { void unlisten.then(fn => fn()) }
  }, [])

  useEffect(() => {
    invoke<AppConfig>("get_config")
      .then((config) => {
        setBackendUrl(config.backend_url)
        setConfigLoaded(true)
      })
      .catch(() => {
        setBackendUrl(null)
        setConfigLoaded(true)
      })
  }, [])

  useEffect(() => {
    if (configLoaded && !backendUrl) {
      setSettingsOpen(true)
    }
  }, [configLoaded, backendUrl])

  useEffect(() => {
    if (!authLoading && configLoaded && backendUrl && !isAuthenticated && !settingsOpen) {
      setAuthModalOpen(true)
    }
  }, [authLoading, configLoaded, backendUrl, isAuthenticated, settingsOpen])

  const wasAuthenticatedRef = useRef(isAuthenticated)
  useEffect(() => {
    if (wasAuthenticatedRef.current && !isAuthenticated) {
      setViewMode("search")
      setQuery("")
      setFilters({})
      setIsFilterOpen(false)
      setAuthModalOpen(false)
    }
    wasAuthenticatedRef.current = isAuthenticated
  }, [isAuthenticated])

  useEffect(() => {
    const container = scrollContainerRef.current
    if (!container || viewMode !== "search") return
    
    let lastSticky = false
    
    const check = () => {
      const el = heroSearchRef.current
      if (!el) return
      
      const rect = el.getBoundingClientRect()
      const isSticky = rect.top <= 24

      if (isSticky !== lastSticky) {
        lastSticky = isSticky
        isSearchStickyRef.current = isSticky
        window.dispatchEvent(new CustomEvent("search-sticky", { detail: isSticky }))
      }
    }
    
    container.addEventListener("scroll", check, { passive: true })
    requestAnimationFrame(check)
    
    return () => {
      container.removeEventListener("scroll", check)
      window.dispatchEvent(new CustomEvent("search-sticky", { detail: false }))
    }
  }, [viewMode])

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
    queryFn: () => keywordSearch({ query: trimmedQuery, limit: 8, ...filters }),
    enabled: isAuthenticated && isQueryReady && viewMode === "search",
    placeholderData: (prev) => prev,
    staleTime: 1000 * 15,
  })

  const {
    data: semanticData,
    isFetching: isSemanticFetching,
    error: semanticError,
  } = useQuery({
    queryKey: ["search", "semantic", trimmedQuery, filters],
    queryFn: () => semanticSearch({ query: trimmedQuery, limit: 8, ...filters }),
    enabled: isAuthenticated && isSemanticReady && viewMode === "search",
    placeholderData: (prev) => prev,
    staleTime: 1000 * 15,
  })

  const { data: entriesData, isLoading: entriesLoading } = useQuery({
    queryKey: ["entries", "count", filters],
    queryFn: () =>
      listEntries({
        path: "/",
        limit: 0,
        own: filters.only_mine,
        visibility: filters.visibility,
        tags: filters.tags,
        from_date: filters.from_date,
        to_date: filters.to_date,
      }),
    enabled: isAuthenticated,
  })

  const keywordResults = keywordData?.results ?? []
  const semanticResults = semanticData?.results ?? []
  const totalEntries = entriesData?.total ?? 0

  const handlePreview = useCallback((result: SearchResult) => {
    setPreviewEntryId(result.entry_id)
  }, [])

  const handleColumnChange = useCallback((column: SearchColumn) => {
    setActiveColumn(column)
  }, [])

  const { selectedIndex, activeColumn: navActiveColumn, selectedRef } = useKeyboardNav({
    keywordResults,
    semanticResults,
    onPreview: handlePreview,
    onColumnChange: handleColumnChange,
    searchInputRef: inputRef,
  })

  useEffect(() => {
    setActiveColumn(navActiveColumn)
  }, [navActiveColumn])

  const handleChange = useCallback((val: string) => setQuery(val), [])

  const filterCount = countActiveFilters(filters)

  const showResults = isAuthenticated && trimmedQuery.length > 0

  const handleSettingsSave = useCallback((newUrl: string) => {
    setBackendUrl(newUrl)
    clearBaseUrlCache()
    setSettingsOpen(false)
    if (newUrl && !isAuthenticated) {
      setAuthModalOpen(true)
    }
  }, [isAuthenticated])

  const handleConnectClick = useCallback(() => {
    setAuthModalOpen(true)
  }, [])

  const entriesFilters: ListEntriesParams = useMemo(() => ({
    own: filters.only_mine,
    visibility: filters.visibility,
    tags: filters.tags,
    from_date: filters.from_date,
    to_date: filters.to_date,
  }), [filters.only_mine, filters.visibility, filters.tags, filters.from_date, filters.to_date])

  if (authLoading || !configLoaded) {
    return (
      <div className="h-screen w-screen overflow-hidden rounded-xl relative">
        <div className="desktop-background" />
        <div className="flex h-full items-center justify-center relative z-10">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        </div>
      </div>
    )
  }

  return (
    <div className="h-screen w-screen overflow-hidden rounded-xl relative flex flex-col">
      <div className="desktop-background" />
      
      <motion.div
        animate={{ 
          opacity: isSearchActive && viewMode === "search" ? 0 : 1,
          pointerEvents: isSearchActive && viewMode === "search" ? "none" : "auto"
        }}
        transition={{ duration: 0.3, ease: "easeOut" }}
      >
        <DesktopHeader
          viewMode={viewMode}
          onViewModeChange={setViewMode}
          onSettingsClick={() => setSettingsOpen(true)}
          onConnectClick={handleConnectClick}
        />
      </motion.div>

      <button
        onClick={isSearchActive && viewMode === "search" ? handleClearSearch : handleClose}
        aria-label={isSearchActive && viewMode === "search" ? "Clear search" : "Close"}
        className="fixed top-4 right-4 z-[70] flex h-9 w-9 items-center justify-center rounded-xl transition-all hover:opacity-60 backdrop-blur-md"
        style={{
          background: "var(--glass-bg)",
          border: "1px solid var(--glass-border)",
          color: "var(--glass-icon)",
        }}
      >
        <X className="h-4 w-4" />
      </button>

      <div 
        ref={scrollContainerRef}
        className="scroll-container flex-1 min-h-0 overflow-y-auto"
        style={{ paddingTop: isSearchActive && viewMode === "search" ? "0px" : "52px" }}
      >
        {viewMode === "search" && (
          <div style={{ display: "flex", flexDirection: "column", minHeight: "100%" }}>
            <motion.div
              animate={{
                opacity: isSearchActive ? 0 : 1,
                height: isSearchActive ? 0 : "auto",
              }}
              transition={{ duration: 0.3, ease: [0.4, 0, 0.2, 1] }}
              style={{
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                justifyContent: "center",
                padding: isSearchActive ? "0" : "2rem 1rem 0.5rem",
                flexShrink: 0,
                overflow: "hidden",
              }}
            >
              <div style={{ textAlign: "center", marginBottom: "1rem" }}>
                <h1
                  style={{
                    fontSize: "1.5rem",
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
                    marginTop: "0.125rem",
                    fontSize: "0.625rem",
                    color: "var(--glass-text-muted)",
                  }}
                >
                  Discover insights across your knowledge base
                </p>
              </div>
            </motion.div>

            <motion.div
              ref={heroSearchRef}
              animate={{
                top: isSearchActive ? "1rem" : "1rem",
              }}
              transition={{ duration: 0.3, ease: [0.4, 0, 0.2, 1] }}
              style={{
                position: "sticky",
                zIndex: 50,
                width: "100%",
                display: "flex",
                justifyContent: "center",
                padding: "0 1rem 1rem",
                pointerEvents: "none",
              }}
            >
              <div style={{ width: "100%", maxWidth: "600px", pointerEvents: "auto" }}>
                {isAuthenticated ? (
                  <div className="relative">
                    <SearchInput
                      ref={inputRef}
                      value={query}
                      onChange={handleChange}
                      autoFocus
                      filterCount={filterCount}
                      onFilterClick={() => setIsFilterOpen(!isFilterOpen)}
                    />
                    <SearchFilters
                      filters={filters}
                      onFiltersChange={setFilters}
                      isOpen={isFilterOpen}
                      onClose={() => setIsFilterOpen(false)}
                      showPath={true}
                    />
                  </div>
                ) : (
                  <SearchInput
                    value=""
                    onChange={() => undefined}
                    disabled
                    placeholder={backendUrl ? "Sign in to search your knowledge base..." : "Configure backend URL first..."}
                  />
                )}
              </div>
            </motion.div>

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
                    margin: "0 auto",
                    padding: "0 1rem 1rem",
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
                    selectedIndex={selectedIndex}
                    activeColumn={activeColumn}
                    selectedRef={selectedRef}
                  />
                </motion.div>
              )}
            </AnimatePresence>


          </div>
        )}

        {viewMode === "entries" && isAuthenticated && (
          <div className="px-4 py-4">
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="rounded-xl overflow-hidden"
              style={{
                background: "var(--glass-bg)",
                border: "1px solid var(--glass-border)",
              }}
            >
              <div className="px-4 py-3">
                <div className="flex items-center justify-between">
                  <h1 className="text-lg font-bold" style={{ color: "var(--glass-text)" }}>
                    {totalEntries === 1 ? "1 Entry" : `${totalEntries} Entries`}
                  </h1>
                  <div className="flex items-center gap-1">
                    <DropdownMenu open={isFilterOpen} onOpenChange={setIsFilterOpen}>
                      <DropdownMenuTrigger asChild>
                        <button
                          className="relative flex items-center justify-center h-8 w-8 rounded-md transition-opacity hover:opacity-60 cursor-pointer"
                          style={{ color: "var(--glass-text)" }}
                        >
                          <Filter className="h-4 w-4" />
                          {filterCount > 0 && (
                            <span
                              className="absolute -top-1 -right-1 flex h-4 min-w-4 items-center justify-center rounded-full text-[10px] font-semibold"
                              style={{ background: "var(--color-primary)", color: "var(--color-primary-foreground)" }}
                            >
                              {filterCount}
                            </span>
                          )}
                        </button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end" className="p-0 border-0 shadow-none bg-transparent">
                        <SearchFilters
                          filters={filters}
                          onFiltersChange={setFilters}
                          isOpen={isFilterOpen}
                          onClose={() => setIsFilterOpen(false)}
                          showPath={false}
                          title="Filters"
                        />
                      </DropdownMenuContent>
                    </DropdownMenu>
                    <button
                      onClick={() => setCreateOpen(true)}
                      className="flex items-center gap-1.5 h-8 px-3 rounded-md text-xs font-medium transition-opacity hover:opacity-60 cursor-pointer"
                      style={{ color: "var(--glass-text)" }}
                    >
                      <Plus className="h-4 w-4" />
                      New
                    </button>
                  </div>
                </div>
              </div>

              {entriesLoading ? (
                <div className="space-y-2 p-4">
                  {[65, 78, 55, 70, 60].map((w, i) => (
                    <div
                      key={i}
                      className="h-8 rounded-md animate-pulse"
                      style={{
                        background: "var(--glass-bg)",
                        width: `${w}%`,
                        marginLeft: `${(i % 3) * 8}px`,
                      }}
                    />
                  ))}
                </div>
              ) : totalEntries === 0 ? (
                <div className="flex flex-col items-center justify-center py-12">
                  <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-2xl mx-auto" style={{ background: "var(--glass-hover)" }}>
                    <BookOpen className="h-6 w-6" style={{ color: "var(--glass-icon)" }} />
                  </div>
                  <h2 className="text-base font-semibold mb-1" style={{ color: "var(--glass-text)" }}>
                    No entries yet
                  </h2>
                  <p className="text-xs mb-4" style={{ color: "var(--glass-text-muted)" }}>
                    Create your first knowledge base entry
                  </p>
                  <button
                    onClick={() => setCreateOpen(true)}
                    className="flex items-center gap-1.5 h-8 px-3 rounded-md text-xs font-medium transition-opacity hover:opacity-60"
                    style={{
                      background: "var(--glass-hover)",
                      color: "var(--glass-text)",
                    }}
                  >
                    <Plus className="h-4 w-4" />
                    New Entry
                  </button>
                </div>
              ) : (
                <FolderTree
                  onEntrySelect={setPreviewEntryId}
                  filters={entriesFilters}
                />
              )}
            </motion.div>
          </div>
        )}

        {viewMode === "entries" && !isAuthenticated && (
          <div className="flex min-h-[60vh] flex-col items-center justify-center gap-4 text-center">
            <div
              className="rounded-2xl p-6"
              style={{
                background: "var(--glass-bg)",
                border: "1px solid var(--glass-border)",
              }}
            >
              <BookOpen className="h-10 w-10 mx-auto mb-3" style={{ color: "var(--glass-icon)" }} />
              <h2 className="text-base font-semibold mb-1" style={{ color: "var(--glass-text)" }}>
                Sign in to view entries
              </h2>
              <p className="text-xs mb-4" style={{ color: "var(--glass-text-muted)" }}>
                Access your knowledge base
              </p>
            </div>
          </div>
        )}
      </div>

      <SettingsPanel
        isOpen={settingsOpen}
        onClose={() => {
          if (backendUrl) {
            setSettingsOpen(false)
          }
        }}
        onSave={handleSettingsSave}
        onTokenChange={connect}
        isAuthenticated={isAuthenticated}
      />
      <AuthModal
        isOpen={authModalOpen}
        onClose={() => setAuthModalOpen(false)}
      />
      <CreateEntryDialog open={createOpen} onOpenChange={setCreateOpen} defaultPath="/" />
      <EntryPreviewModal entryId={previewEntryId} onClose={() => setPreviewEntryId(null)} />
    </div>
  )
}
