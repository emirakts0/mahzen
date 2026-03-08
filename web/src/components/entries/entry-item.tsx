import { motion } from "framer-motion"
import { Eye, EyeOff } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { getFileIcon } from "./file-icons"
import type { Entry } from "@/types/api"

interface EntryItemProps {
  entry: Entry
  depth: number
  onSelect: () => void
}

export function EntryItem({ entry, depth, onSelect }: EntryItemProps) {
  const FileIcon = getFileIcon(entry.file_type)

  return (
    <motion.button
      onClick={onSelect}
      className="w-full flex items-center gap-2 px-2 py-1.5 rounded-md cursor-pointer transition-all duration-200"
      style={{
        paddingLeft: depth * 16 + 24,
        background: "transparent",
        color: "var(--glass-text)",
      }}
      whileHover={{ background: "var(--glass-hover)" }}
      whileTap={{ scale: 0.99 }}
    >
      <FileIcon className="h-4 w-4 shrink-0" style={{ color: "var(--glass-icon)" }} />

      <div className="flex-1 min-w-0 flex items-center gap-2">
        <span className="text-sm font-medium truncate">
          {entry.title || "(Untitled)"}
        </span>
        {entry.visibility === "public" ? (
          <Eye className="h-3 w-3 text-green-700 shrink-0" />
        ) : (
          <EyeOff className="h-3 w-3 text-rose-600 shrink-0" />
        )}
      </div>

      {entry.tags && entry.tags.length > 0 && (
        <div className="flex gap-1 shrink-0">
          {entry.tags.slice(0, 2).map((tag) => (
            <Badge
              key={tag}
              variant="secondary"
              className="text-[10px] px-1.5 py-0 h-4"
            >
              {tag}
            </Badge>
          ))}
          {entry.tags.length > 2 && (
            <span className="text-[10px]" style={{ color: "var(--glass-text-muted)" }}>
              +{entry.tags.length - 2}
            </span>
          )}
        </div>
      )}

      {entry.file_type && (
        <span
          className="text-[10px] font-mono px-1.5 py-0.5 rounded shrink-0"
          style={{
            background: "var(--glass-hover)",
            color: "var(--glass-text-muted)",
          }}
        >
          .{entry.file_type}
        </span>
      )}
    </motion.button>
  )
}
