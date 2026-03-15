import { useState, useCallback, useMemo } from "react"
import { useQuery } from "@tanstack/react-query"
import { listEntries } from "@/api/entries"
import { FolderItem } from "./folder-item"
import { EntryItem } from "./entry-item"
import type { Entry, ListEntriesParams } from "@/types/api"

const SKELETON_WIDTHS = [65, 78, 55, 70, 60]

function SkeletonLoader({ count = 2 }: { count?: number }) {
  return (
    <div className="space-y-1 pl-4">
      {SKELETON_WIDTHS.slice(0, count).map((w, i) => (
        <div
          key={i}
          className="h-6 rounded animate-pulse"
          style={{ background: "var(--glass-bg)", width: `${w}%` }}
        />
      ))}
    </div>
  )
}

interface FolderTreeProps {
  onEntrySelect: (entryId: string) => void
  filters?: ListEntriesParams
}

export function FolderTree({ onEntrySelect, filters = {} }: FolderTreeProps) {
  const [expandedFolders, setExpandedFolders] = useState<Set<string>>(new Set(["/"]))

  const stableFilters = useMemo(() => ({
    own: filters.own,
    visibility: filters.visibility,
    tags: filters.tags,
    from_date: filters.from_date,
    to_date: filters.to_date,
  }), [filters.own, filters.visibility, filters.tags, filters.from_date, filters.to_date])

  const toggleFolder = useCallback((path: string) => {
    setExpandedFolders((prev) => {
      const next = new Set(prev)
      if (next.has(path)) {
        next.delete(path)
      } else {
        next.add(path)
      }
      return next
    })
  }, [])

  return (
    <>
      <div className="py-2">
        <FolderNode
          path="/"
          depth={0}
          expandedFolders={expandedFolders}
          onToggle={toggleFolder}
          onEntrySelect={onEntrySelect}
          filters={stableFilters}
        />
      </div>
    </>
  )
}

interface FolderNodeProps {
  path: string
  depth: number
  expandedFolders: Set<string>
  folderCount?: number
  filters: ListEntriesParams
  onToggle: (path: string) => void
  onEntrySelect: (entryId: string) => void
}

function FolderNode({
  path,
  depth,
  expandedFolders,
  folderCount,
  filters,
  onToggle,
  onEntrySelect,
}: FolderNodeProps) {
  const isExpanded = expandedFolders.has(path)

  const queryKey = useMemo(() => [
    "entries",
    path,
    filters?.own,
    filters?.visibility,
    filters?.tags,
    filters?.from_date,
    filters?.to_date,
  ], [path, filters?.own, filters?.visibility, filters?.tags, filters?.from_date, filters?.to_date])

  const { data, isLoading, error } = useQuery({
    queryKey,
    queryFn: () => listEntries({ ...filters, path, limit: 100 }),
    enabled: isExpanded,
    staleTime: 30000,
  })

  const entries = data?.entries ?? []
  const folders = data?.folders ?? []
  const folderEntries = entries.filter((e: Entry) => e.path === path)

  const node = {
    name: path === "/" ? "" : path.split("/").pop() || path,
    path,
    isFolder: true,
    children: [],
  }

  return (
    <div>
      {path !== "/" && (
        <FolderItem
          node={node}
          depth={depth}
          isExpanded={isExpanded}
          count={folderCount}
          onToggle={() => onToggle(path)}
        >
          <div className="space-y-0.5">
            {isLoading && <SkeletonLoader count={2} />}
            {error && (
              <div className="px-4 py-2 text-xs" style={{ color: "var(--glass-error, #ef4444)" }}>
                Failed to load entries
              </div>
            )}

            {!isLoading && !error && folders.map((folder) => (
              <FolderNode
                key={folder.path}
                path={folder.path}
                depth={depth + 1}
                expandedFolders={expandedFolders}
                folderCount={folder.count}
                filters={filters}
                onToggle={onToggle}
                onEntrySelect={onEntrySelect}
              />
            ))}

            {!isLoading && folderEntries.map((entry: Entry) => (
              <EntryItem
                key={entry.id}
                entry={entry}
                depth={depth + 1}
                onSelect={() => onEntrySelect(entry.id)}
              />
            ))}
          </div>
        </FolderItem>
      )}

      {path === "/" && (
        <div className="space-y-0.5">
          {isLoading && <SkeletonLoader count={3} />}
          {error && (
            <div className="px-4 py-2 text-xs" style={{ color: "var(--glass-error, #ef4444)" }}>
              Failed to load entries
            </div>
          )}

          {!isLoading && !error && folders.map((folder) => (
            <FolderNode
              key={folder.path}
              path={folder.path}
              depth={depth}
              expandedFolders={expandedFolders}
              folderCount={folder.count}
              filters={filters}
              onToggle={onToggle}
              onEntrySelect={onEntrySelect}
            />
          ))}

          {!isLoading && !error && folderEntries.map((entry: Entry) => (
            <EntryItem
              key={entry.id}
              entry={entry}
              depth={depth + 1}
              onSelect={() => onEntrySelect(entry.id)}
            />
          ))}
        </div>
      )}
    </div>
  )
}
