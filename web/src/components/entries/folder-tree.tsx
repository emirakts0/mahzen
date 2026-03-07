import { useState, useCallback } from "react"
import { useQuery } from "@tanstack/react-query"
import { motion } from "framer-motion"
import { listEntries } from "@/api/entries"
import { FolderItem } from "./folder-item"
import { EntryItem } from "./entry-item"
import { EntryPreviewModal } from "@/components/search/entry-preview-modal"
import type { Entry } from "@/types/api"

interface FolderTreeProps {
  selectedPath: string | null
  onPathSelect: (path: string) => void
}

export function FolderTree({ selectedPath, onPathSelect }: FolderTreeProps) {
  const [expandedFolders, setExpandedFolders] = useState<Set<string>>(new Set(["/"]))
  const [previewEntryId, setPreviewEntryId] = useState<string | null>(null)

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

  const handleEntrySelect = useCallback((entryId: string) => {
    setPreviewEntryId(entryId)
  }, [])

  return (
    <>
      <div className="py-2">
        <FolderNode
          path="/"
          depth={0}
          expandedFolders={expandedFolders}
          selectedPath={selectedPath}
          onToggle={toggleFolder}
          onPathSelect={onPathSelect}
          onEntrySelect={handleEntrySelect}
        />
      </div>

      <EntryPreviewModal
        entryId={previewEntryId}
        onClose={() => setPreviewEntryId(null)}
      />
    </>
  )
}

interface FolderNodeProps {
  path: string
  depth: number
  expandedFolders: Set<string>
  selectedPath: string | null
  onToggle: (path: string) => void
  onPathSelect: (path: string) => void
  onEntrySelect: (entryId: string) => void
}

function FolderNode({
  path,
  depth,
  expandedFolders,
  selectedPath,
  onToggle,
  onPathSelect,
  onEntrySelect,
}: FolderNodeProps) {
  const isExpanded = expandedFolders.has(path)
  const isSelected = selectedPath === path

  const { data, isLoading } = useQuery({
    queryKey: ["entries", path],
    queryFn: () => listEntries({ path, limit: 100 }),
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
          isSelected={isSelected}
          onToggle={() => onToggle(path)}
          onSelect={() => onPathSelect(path)}
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

            {!isLoading && folders.map((folderPath) => (
              <FolderNode
                key={folderPath}
                path={folderPath}
                depth={depth + 1}
                expandedFolders={expandedFolders}
                selectedPath={selectedPath}
                onToggle={onToggle}
                onPathSelect={onPathSelect}
                onEntrySelect={onEntrySelect}
              />
            ))}

            {!isLoading && folderEntries.map((entry: Entry) => (
              <EntryItem
                key={entry.id}
                entry={entry}
                depth={depth + 1}
                isSelected={false}
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

          {!isLoading && folders.map((folderPath) => (
            <FolderNode
              key={folderPath}
              path={folderPath}
              depth={depth}
              expandedFolders={expandedFolders}
              selectedPath={selectedPath}
              onToggle={onToggle}
              onPathSelect={onPathSelect}
              onEntrySelect={onEntrySelect}
            />
          ))}

          {!isLoading && folderEntries.map((entry: Entry) => (
            <EntryItem
              key={entry.id}
              entry={entry}
              depth={depth + 1}
              isSelected={false}
              onSelect={() => onEntrySelect(entry.id)}
            />
          ))}
        </div>
      )}
    </div>
  )
}
