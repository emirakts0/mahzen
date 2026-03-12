import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { motion, AnimatePresence } from "framer-motion"
import { X, Tag, Folder, Lock, Unlock, User, Search } from "lucide-react"
import { listTags } from "@/api/tags"
import { DateRangePicker } from "@/components/ui/date-range-picker"
import type { SearchFilters as SearchFiltersType } from "@/types/api"

interface SearchFiltersProps {
  filters: SearchFiltersType
  onFiltersChange: (filters: SearchFiltersType) => void
  isOpen: boolean
  onClose: () => void
  showPath?: boolean
  title?: string
}

export function SearchFilters({ 
  filters, 
  onFiltersChange, 
  isOpen, 
  onClose, 
  showPath = true,
  title = "Filters"
}: SearchFiltersProps) {
  const [showAllTags, setShowAllTags] = useState(false)
  const [tagSearch, setTagSearch] = useState("")

  const { data: tagsData } = useQuery({
    queryKey: ["tags"],
    queryFn: () => listTags({ limit: 100 }),
    staleTime: 1000 * 60 * 5,
  })

  const tags = tagsData?.tags ?? []

  const filteredTags = tagSearch
    ? tags.filter((t) => t.name.toLowerCase().includes(tagSearch.toLowerCase()))
    : tags

  const visibleTags = showAllTags || tagSearch ? filteredTags : filteredTags.slice(0, 8)
  const hiddenCount = filteredTags.length - 8

  const hasActiveFilters =
    (filters.tags?.length ?? 0) > 0 ||
    (showPath && filters.path) ||
    filters.from_date ||
    filters.to_date ||
    filters.only_mine ||
    filters.visibility

  const clearFilters = () => {
    onFiltersChange({})
  }

  const toggleTag = (tagName: string) => {
    const currentTags = filters.tags ?? []
    const newTags = currentTags.includes(tagName)
      ? currentTags.filter((t) => t !== tagName)
      : [...currentTags, tagName]
    onFiltersChange({ ...filters, tags: newTags.length > 0 ? newTags : undefined })
  }

  const updateFilter = <K extends keyof SearchFiltersType>(key: K, value: SearchFiltersType[K]) => {
    onFiltersChange({ ...filters, [key]: value || undefined })
  }

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.1 }}
            className="fixed inset-0 z-40"
            onClick={onClose}
          />
          <motion.div
            initial={{ opacity: 0, y: -8, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -8, scale: 0.95 }}
            transition={{ duration: 0.15, ease: "easeOut" }}
            className="absolute right-0 top-full z-50 mt-2 w-80 rounded-xl p-4 shadow-xl backdrop-blur-xl"
            style={{
              background: "var(--glass-bg)",
              border: "1px solid var(--glass-border)",
            }}
          >
            <div className="mb-3 flex items-center justify-between">
              <span className="text-sm font-semibold" style={{ color: "var(--glass-text)" }}>
                {title}
              </span>
              {hasActiveFilters && (
                <button
                  onClick={clearFilters}
                  className="flex items-center gap-1 rounded-md px-2 py-1 text-xs transition-colors"
                  style={{ color: "var(--glass-text-muted)" }}
                >
                  <X className="h-3 w-3" />
                  Clear all
                </button>
              )}
            </div>

            <div className="space-y-4">
              {showPath && (
                <div>
                  <label
                    className="mb-1.5 flex items-center gap-1.5 text-xs font-medium"
                    style={{ color: "var(--glass-text-muted)" }}
                  >
                    <Folder className="h-3 w-3" />
                    Path
                  </label>
                  <input
                    type="text"
                    placeholder="/notes/work"
                    value={filters.path ?? ""}
                    onChange={(e) => updateFilter("path", e.target.value)}
                    className="h-8 w-full rounded-md px-2 text-sm outline-none backdrop-blur-sm"
                    style={{
                      background: "var(--glass-hover)",
                      border: "1px solid var(--glass-border)",
                      color: "var(--glass-text)",
                    }}
                  />
                </div>
              )}

              <div>
                <label
                  className="mb-1.5 block text-xs font-medium"
                  style={{ color: "var(--glass-text-muted)" }}
                >
                  Date Range
                </label>
                <DateRangePicker
                  value={{ from: filters.from_date, to: filters.to_date }}
                  onChange={(range) => {
                    onFiltersChange({
                      ...filters,
                      from_date: range.from,
                      to_date: range.to,
                    })
                  }}
                />
              </div>

              <div>
                <label
                  className="mb-1.5 flex items-center gap-1.5 text-xs font-medium"
                  style={{ color: "var(--glass-text-muted)" }}
                >
                  <Lock className="h-3 w-3" />
                  Visibility
                </label>
                <div className="flex gap-1.5">
                  {[
                    { value: "", label: "All", icon: null },
                    { value: "public", label: "Public", icon: Unlock },
                    { value: "private", label: "Private", icon: Lock },
                  ].map((option) => {
                    const Icon = option.icon
                    const isSelected = (filters.visibility ?? "") === option.value
                    return (
                      <button
                        key={option.value}
                        onClick={() => updateFilter("visibility", option.value as "public" | "private" | undefined)}
                        className="flex flex-1 items-center justify-center gap-1 rounded-md px-2 py-1.5 text-xs font-medium transition-all"
                        style={{
                          background: isSelected
                            ? "var(--glass-accent, #3b82f6)"
                            : "var(--glass-hover)",
                          color: isSelected ? "white" : "var(--glass-text)",
                          border: "1px solid var(--glass-border)",
                        }}
                      >
                        {Icon && <Icon className="h-3 w-3" />}
                        {option.label}
                      </button>
                    )
                  })}
                </div>
                <div className="mt-2 flex items-center justify-between rounded-md p-2" style={{ background: "var(--glass-hover)" }}>
                  <label
                    className="flex items-center gap-2 text-xs font-medium"
                    style={{ color: "var(--glass-text)" }}
                  >
                    <User className="h-3.5 w-3.5" />
                    Only my entries
                  </label>
                  <button
                    onClick={() => updateFilter("only_mine", !filters.only_mine)}
                    className="relative h-5 w-9 rounded-full transition-colors"
                    style={{
                      background: filters.only_mine
                        ? "var(--glass-accent, #3b82f6)"
                        : "var(--glass-border)",
                    }}
                  >
                    <motion.div
                      className="absolute top-0.5 h-4 w-4 rounded-full bg-white shadow-sm"
                      animate={{ left: filters.only_mine ? 18 : 2 }}
                      transition={{ type: "spring", stiffness: 500, damping: 30 }}
                    />
                  </button>
                </div>
              </div>

              <div>
                <label
                  className="mb-1.5 flex items-center gap-1.5 text-xs font-medium"
                  style={{ color: "var(--glass-text-muted)" }}
                >
                  <Tag className="h-3 w-3" />
                  Tags
                </label>
                <div className="relative mb-2">
                  <Search
                    className="absolute left-2 top-1/2 h-3 w-3 -translate-y-1/2"
                    style={{ color: "var(--glass-text-muted)" }}
                  />
                  <input
                    type="text"
                    placeholder="Search tags..."
                    value={tagSearch}
                    onChange={(e) => {
                      setTagSearch(e.target.value)
                      if (e.target.value) setShowAllTags(true)
                    }}
                    className="h-7 w-full rounded-md pl-7 pr-2 text-xs outline-none"
                    style={{
                      background: "var(--glass-bg)",
                      border: "1px solid var(--glass-border)",
                      color: "var(--glass-text)",
                    }}
                  />
                </div>
                <div className="flex flex-wrap gap-1.5">
                  {visibleTags.length === 0 ? (
                    <span className="text-xs" style={{ color: "var(--glass-text-muted)" }}>
                      {tagSearch ? "No matching tags" : "No tags available"}
                    </span>
                  ) : (
                    visibleTags.map((tag) => {
                      const isSelected = filters.tags?.includes(tag.name)
                      return (
                        <button
                          key={tag.id}
                          onClick={() => toggleTag(tag.name)}
                          className="rounded-md px-2 py-1 text-xs font-medium transition-all"
                          style={{
                            background: isSelected
                              ? "var(--glass-accent, #3b82f6)"
                              : "var(--glass-bg)",
                            color: isSelected ? "white" : "var(--glass-text)",
                            border: "1px solid var(--glass-border)",
                          }}
                        >
                          {tag.name}
                        </button>
                      )
                    })
                  )}
                </div>
                {!tagSearch && hiddenCount > 0 && !showAllTags && (
                  <button
                    onClick={() => setShowAllTags(true)}
                    className="mt-1.5 text-xs font-medium transition-colors"
                    style={{ color: "var(--glass-accent, #3b82f6)" }}
                  >
                    +{hiddenCount} more
                  </button>
                )}
                {showAllTags && !tagSearch && hiddenCount > 0 && (
                  <button
                    onClick={() => setShowAllTags(false)}
                    className="mt-1.5 text-xs font-medium transition-colors"
                    style={{ color: "var(--glass-text-muted)" }}
                  >
                    Show less
                  </button>
                )}
              </div>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  )
}

export function countActiveFilters(filters: SearchFiltersType, showPath = true): number {
  return (
    (filters.tags?.length ?? 0) +
    (showPath && filters.path ? 1 : 0) +
    (filters.visibility ? 1 : 0) +
    (filters.only_mine ? 1 : 0) +
    (filters.from_date ? 1 : 0) +
    (filters.to_date ? 1 : 0)
  )
}
