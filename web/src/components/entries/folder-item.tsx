import { motion } from "framer-motion"
import { Folder, FolderOpen, ChevronRight } from "lucide-react"
import { cn } from "@/lib/utils"
import type { TreeNode } from "./tree-utils"

interface FolderItemProps {
  node: TreeNode
  depth: number
  isExpanded: boolean
  count?: number
  onToggle: () => void
  children?: React.ReactNode
}

export function FolderItem({
  node,
  depth,
  isExpanded,
  count,
  onToggle,
  children,
}: FolderItemProps) {
  return (
    <div>
      <motion.button
        onClick={onToggle}
        className={cn(
          "w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-left transition-all duration-200"
        )}
        style={{
          paddingLeft: depth * 16 + 8,
          background: "transparent",
          color: "var(--glass-text)",
        }}
        whileHover={{ background: "var(--glass-hover)" }}
        whileTap={{ scale: 0.99 }}
      >
        <motion.div
          animate={{ rotate: isExpanded ? 90 : 0 }}
          transition={{ duration: 0.2 }}
          className="shrink-0"
          style={{ color: "var(--glass-icon)" }}
        >
          <ChevronRight className="h-3.5 w-3.5" />
        </motion.div>

        <motion.div
          initial={false}
          animate={{ scale: isExpanded ? 1.1 : 1 }}
          transition={{ duration: 0.15 }}
        >
          {isExpanded ? (
            <FolderOpen className="h-4 w-4" style={{ color: "var(--glass-icon)" }} />
          ) : (
            <Folder className="h-4 w-4" style={{ color: "var(--glass-icon)" }} />
          )}
        </motion.div>

        <span className="text-sm font-medium truncate">
          {node.name || "root"}
        </span>

        {count !== undefined && count > 0 && (
          <span className="ml-auto text-xs" style={{ color: "var(--glass-text-muted)" }}>
            {count}
          </span>
        )}
      </motion.button>

      <div
        style={{
          display: "grid",
          gridTemplateRows: isExpanded ? "1fr" : "0fr",
          opacity: isExpanded ? 1 : 0,
          transition: "grid-template-rows 0.2s ease-out, opacity 0.15s ease-out",
        }}
      >
        <div className="overflow-hidden" style={{ minHeight: 0 }}>
          {children}
        </div>
      </div>
    </div>
  )
}
