import { motion } from "framer-motion"
import { Lock, Globe, User } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { FileIcon } from "./file-icons"
import { useAuth } from "@/providers/auth-provider"
import type { Entry } from "@/types/api"

interface EntryItemProps {
  entry: Entry
  depth: number
  onSelect: () => void
}

export function EntryItem({ entry, depth, onSelect }: EntryItemProps) {
  const { user } = useAuth()
  const isOwner = user?.id === entry.user_id

  return (
    <motion.button
      onClick={onSelect}
      className="w-full flex items-center gap-1.5 px-2 py-1 rounded-md cursor-pointer transition-all duration-200"
      style={{
        paddingLeft: depth * 12 + 16,
        background: "transparent",
        color: "var(--glass-text)",
      }}
      whileHover={{ background: "var(--glass-hover)" }}
      whileTap={{ scale: 0.99 }}
    >
      <FileIcon fileType={entry.file_type} className="h-3.5 w-3.5 shrink-0" style={{ color: "var(--glass-icon)" }} />

      <div className="flex-1 min-w-0 flex items-center gap-1.5">
        <span className="text-xs font-medium truncate">
          {entry.title || "(Untitled)"}
        </span>
        <div className="flex items-center gap-1 shrink-0">
          {isOwner && (
            <User className="h-2.5 w-2.5 shrink-0" style={{ color: "var(--color-primary)" }} />
          )}
          {entry.visibility === "public" ? (
            <Globe className="h-2.5 w-2.5 text-green-700 shrink-0" />
          ) : (
            <Lock className="h-2.5 w-2.5 text-rose-600 shrink-0" />
          )}
        </div>
      </div>

      {entry.tags && entry.tags.length > 0 && (
        <div className="flex gap-0.5 shrink-0">
          {entry.tags.slice(0, 2).map((tag) => (
            <Badge
              key={tag}
              variant="secondary"
              className="text-[9px] px-1 py-0 h-3.5"
            >
              {tag}
            </Badge>
          ))}
          {entry.tags.length > 2 && (
            <span className="text-[9px]" style={{ color: "var(--glass-text-muted)" }}>
              +{entry.tags.length - 2}
            </span>
          )}
        </div>
      )}

      {entry.file_type && (
        <span
          className="text-[9px] font-mono px-1 py-0.5 rounded shrink-0"
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
