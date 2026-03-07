import { useState } from "react"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { Link } from "react-router"
import { motion } from "framer-motion"
import { Plus, BookOpen, RefreshCw, FolderTree as FolderTreeIcon } from "lucide-react"
import { listEntries } from "@/api/entries"
import { Button } from "@/components/ui/button"
import { useAuth } from "@/hooks/use-auth"
import { FolderTree, CreateEntryDialog } from "@/components/entries"
import { EntryPreviewModal } from "@/components/search/entry-preview-modal"

export default function EntriesPage() {
  const { isAuthenticated } = useAuth()
  const [createOpen, setCreateOpen] = useState(false)
  const [previewEntryId, setPreviewEntryId] = useState<string | null>(null)
  const queryClient = useQueryClient()

  // Use same query key as FolderTree to share cache
  const { data, isLoading, isFetching } = useQuery({
    queryKey: ["entries", "/"],
    queryFn: () => listEntries({ path: "/", limit: 0 }),
    enabled: isAuthenticated,
  })

  const total = data?.total ?? 0
  const isRefetching = isFetching && !isLoading

  const handleRefresh = () => {
    queryClient.invalidateQueries({ queryKey: ["entries"] })
  }

  if (!isAuthenticated) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-4 text-center">
        <div className="backdrop-blur-xl rounded-2xl p-8" style={{ background: "var(--glass-bg)", border: "1px solid var(--glass-border)" }}>
          <BookOpen className="h-12 w-12 mx-auto mb-4 text-primary" />
          <h2 className="text-xl font-semibold mb-2" style={{ color: "var(--glass-text)" }}>Sign in to view entries</h2>
          <p className="text-sm mb-6" style={{ color: "var(--glass-text-muted)" }}>Access your knowledge base from anywhere</p>
          <div className="flex gap-3 justify-center">
            <Button asChild>
              <Link to="?auth=login" replace>Sign in</Link>
            </Button>
            <Button variant="outline" asChild>
              <Link to="?auth=signup" replace>Create account</Link>
            </Button>
          </div>
        </div>
      </div>
    )
  }



  return (
    <div className="min-h-screen">
      <div className="mx-auto w-full max-w-5xl px-4 py-6">
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-6"
        >
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold flex items-center gap-2" style={{ color: "var(--glass-text)" }}>
                <FolderTreeIcon className="h-6 w-6 text-primary" />
                Entries
              </h1>
              <p className="text-sm mt-0.5" style={{ color: "var(--glass-text-muted)" }}>
                {total} {total === 1 ? "entry" : "entries"} in your knowledge base
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="icon"
                onClick={handleRefresh}
                disabled={isRefetching}
                className="backdrop-blur-md"
                style={{ background: "var(--glass-bg)", border: "1px solid var(--glass-border)", color: "var(--glass-text)" }}
              >
                <RefreshCw className={`h-4 w-4 ${isRefetching ? "animate-spin" : ""}`} />
              </Button>
              <Button
                onClick={() => setCreateOpen(true)}
                className="backdrop-blur-md"
                style={{ background: "var(--glass-bg)", border: "1px solid var(--glass-border)", color: "var(--glass-text)" }}
              >
                <Plus className="mr-2 h-4 w-4" />
                New Entry
              </Button>
            </div>
          </div>
        </motion.div>

        {isLoading ? (
          <div className="space-y-2">
            {[1, 2, 3, 4, 5].map((i) => (
              <motion.div
                key={i}
                className="h-10 rounded-md animate-pulse"
                style={{ background: "var(--glass-bg)", width: `${50 + Math.random() * 50}%`, marginLeft: `${Math.random() * 20}px` }}
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: i * 0.05 }}
              />
            ))}
          </div>
        ) : total === 0 ? (
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="flex flex-col items-center justify-center py-20"
          >
            <div className="backdrop-blur-xl rounded-2xl p-8 text-center" style={{ background: "var(--glass-bg)", border: "1px solid var(--glass-border)" }}>
              <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl mx-auto bg-primary/10">
                <BookOpen className="h-8 w-8 text-primary" />
              </div>
              <h2 className="text-xl font-semibold mb-2" style={{ color: "var(--glass-text)" }}>No entries yet</h2>
              <p className="text-sm mb-6" style={{ color: "var(--glass-text-muted)" }}>Create your first knowledge base entry</p>
              <Button
                onClick={() => setCreateOpen(true)}
                className="backdrop-blur-md"
                style={{ background: "var(--glass-bg)", border: "1px solid var(--glass-border)", color: "var(--glass-text)" }}
              >
                <Plus className="mr-2 h-4 w-4" />
                New Entry
              </Button>
            </div>
          </motion.div>
        ) : (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.1 }}
            className="backdrop-blur-xl rounded-xl overflow-hidden"
            style={{ background: "var(--glass-bg)", border: "1px solid var(--glass-border)" }}
          >
            <FolderTree
              onEntrySelect={setPreviewEntryId}
            />
          </motion.div>
        )}
      </div>

      <CreateEntryDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        defaultPath="/"
      />

      <EntryPreviewModal
        entryId={previewEntryId}
        onClose={() => setPreviewEntryId(null)}
      />
    </div>
  )
}
