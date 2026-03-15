import { motion, AnimatePresence } from "framer-motion"
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
          "w-full flex items-center gap-1.5 px-2 py-1 rounded-md text-left transition-all duration-200"
        )}
        style={{
          paddingLeft: depth * 12 + 8,
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
          <ChevronRight className="h-3 w-3" />
        </motion.div>

        <motion.div
          initial={false}
          animate={{ scale: isExpanded ? 1.1 : 1 }}
          transition={{ duration: 0.15 }}
        >
          {isExpanded ? (
            <FolderOpen className="h-3.5 w-3.5" style={{ color: "var(--glass-icon)" }} />
          ) : (
            <Folder className="h-3.5 w-3.5" style={{ color: "var(--glass-icon)" }} />
          )}
        </motion.div>

        <span className="text-xs font-medium truncate">
          {node.name || "root"}
        </span>

        {count !== undefined && count > 0 && (
          <span className="ml-auto text-[10px]" style={{ color: "var(--glass-text-muted)" }}>
            {count}
          </span>
        )}
      </motion.button>

      <AnimatePresence initial={false}>
        {isExpanded && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2, ease: "easeOut" }}
            className="overflow-hidden"
          >
            {children}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
