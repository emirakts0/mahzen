import { forwardRef } from "react"
import { Search, X, Filter } from "lucide-react"
import { cn } from "@/lib/utils"

interface SearchInputProps {
  value: string
  onChange: (value: string) => void
  onClear?: () => void
  placeholder?: string
  disabled?: boolean
  className?: string
  hint?: string | null
  autoFocus?: boolean
  filterCount?: number
  onFilterClick?: () => void
}

export const SearchInput = forwardRef<HTMLInputElement, SearchInputProps>(
  ({ value, onChange, onClear, placeholder = "Search your knowledge base...", disabled = false, className, hint, autoFocus, filterCount, onFilterClick }, ref) => {
    const hasFilters = filterCount !== undefined && filterCount > 0

    return (
      <div className={cn("relative w-full", className)}>
        <div className="relative flex items-center">
          <Search
            className="absolute left-4 z-10 h-5 w-5 shrink-0"
            style={{ color: "var(--glass-text-muted)" }}
          />
          <input
            ref={ref}
            type="text"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            placeholder={placeholder}
            disabled={disabled}
            autoFocus={autoFocus}
            spellCheck={false}
            autoComplete="off"
            className={cn(
              "h-12 w-full rounded-xl pl-11 pr-20 text-base shadow-sm outline-none transition-all backdrop-blur-md",
              "disabled:cursor-not-allowed disabled:opacity-50",
            )}
            style={{
              background: "var(--glass-bg)",
              border: "1px solid var(--glass-border)",
              color: "var(--glass-text)",
            }}
          />
          <div className="absolute right-2 flex items-center gap-1">
            {onFilterClick && (
              <button
                type="button"
                onClick={onFilterClick}
                className="flex items-center gap-1 rounded-md px-2 py-1 transition-all hover:opacity-80 cursor-pointer"
                style={{
                  color: hasFilters ? "var(--glass-text)" : "var(--glass-text-muted)",
                }}
              >
                {hasFilters && (
                  <span className="text-xs font-semibold tabular-nums">{filterCount}</span>
                )}
                <Filter className="h-4 w-4" />
              </button>
            )}
            {value && (
              <button
                type="button"
                onClick={onClear}
                aria-label="Clear search"
                className="flex h-6 w-6 items-center justify-center rounded-md transition-colors"
                style={{ color: "var(--glass-icon)" }}
              >
                <X className="h-4 w-4" />
              </button>
            )}
          </div>
        </div>

        {hint && (
          <p className="mt-2 text-center text-xs" style={{ color: "var(--glass-text-muted)" }}>{hint}</p>
        )}
      </div>
    )
  }
)

SearchInput.displayName = "SearchInput"
