import { useState, useCallback } from "react"
import { useQuery } from "@tanstack/react-query"
import { motion } from "framer-motion"
import { listEntries } from "@/api/entries"
import { FolderItem } from "./folder-item"
import { EntryItem } from "./entry-item"
import type { Entry } from "@/types/api"

interface FolderTreeProps {
  onEntrySelect: (entryId: string) => void
  ownOnly?: boolean
}

export function FolderTree({ onEntrySelect, ownOnly = false }: FolderTreeProps) {
  const [expandedFolders, setExpandedFolders] = useState<Set<string>>(new Set(["/"]))

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
          ownOnly={ownOnly}
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
  ownOnly: boolean
  onToggle: (path: string) => void
  onEntrySelect: (entryId: string) => void
}

function FolderNode({
  path,
  depth,
  expandedFolders,
  folderCount,
  ownOnly,
  onToggle,
  onEntrySelect,
}: FolderNodeProps) {
  const isExpanded = expandedFolders.has(path)

  const { data, isLoading } = useQuery({
    queryKey: ["entries", path, ownOnly],
    queryFn: () => listEntries({ path, limit: 100, own: ownOnly }),
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
            {isLoading && (
              <div className="space-y-1 pl-4">
                {[1, 2].map((i) => (
                  <motion.div
                    key={i}
                    className="h-6 rounded animate-pulse"
                    style={{ background: "var(--glass-bg)", width: `${50 + Math.random() * 30}%` }}
                  />
                ))}
              </div>
            )}

            {!isLoading && folders.map((folder) => (
              <FolderNode
                key={folder.path}
                path={folder.path}
                depth={depth + 1}
                expandedFolders={expandedFolders}
                folderCount={folder.count}
                ownOnly={ownOnly}
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
          {isLoading && (
            <div className="space-y-1 pl-4">
              {[1, 2, 3].map((i) => (
                <motion.div
                  key={i}
                  className="h-6 rounded animate-pulse"
                  style={{ background: "var(--glass-bg)", width: `${50 + Math.random() * 30}%` }}
                />
              ))}
            </div>
          )}

          {!isLoading && folders.map((folder) => (
            <FolderNode
              key={folder.path}
              path={folder.path}
              depth={depth}
              expandedFolders={expandedFolders}
              folderCount={folder.count}
              ownOnly={ownOnly}
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
      )}
    </div>
  )
}
