import { useState } from "react"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { Link } from "react-router"
import { motion, AnimatePresence } from "framer-motion"
import { Plus, Tag, Trash2, Loader2, AlertCircle } from "lucide-react"
import { toast } from "sonner"
import { listTags, createTag, deleteTag } from "@/api/tags"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useAuth } from "@/hooks/use-auth"
import type { Tag as TagType } from "@/types/api"

const containerVariants = {
  hidden: {},
  visible: { transition: { staggerChildren: 0.04 } },
}

const itemVariants = {
  hidden: { opacity: 0, scale: 0.95 },
  visible: { opacity: 1, scale: 1, transition: { duration: 0.15 } },
}

function TagCardSkeleton() {
  return (
    <div className="flex items-center justify-between rounded-lg border border-border bg-card px-4 py-3">
      <div className="space-y-1">
        <Skeleton className="h-4 w-24" />
        <Skeleton className="h-3 w-16" />
      </div>
      <Skeleton className="h-8 w-8 rounded-md" />
    </div>
  )
}

function EmptyState({ onNew }: { onNew: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-24 text-center">
      <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/10">
        <Tag className="h-8 w-8 text-primary" />
      </div>
      <h2 className="text-xl font-semibold text-foreground">No tags yet</h2>
      <p className="mt-1 text-sm text-muted-foreground">
        Create tags to organize your entries
      </p>
      <Button className="mt-6" onClick={onNew}>
        <Plus className="mr-2 h-4 w-4" />
        New Tag
      </Button>
    </div>
  )
}

interface CreateTagDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

function CreateTagDialog({ open, onOpenChange }: CreateTagDialogProps) {
  const queryClient = useQueryClient()
  const [name, setName] = useState("")

  const mutation = useMutation({
    mutationFn: () => createTag({ name }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["tags"] })
      toast.success("Tag created")
      setName("")
      onOpenChange(false)
    },
    onError: (err: Error) => toast.error(err.message || "Failed to create tag"),
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return
    mutation.mutate()
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!mutation.isPending) { onOpenChange(v); if (!v) setName("") } }}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>New Tag</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="tag-name">Name</Label>
            <Input
              id="tag-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. machine-learning"
              autoFocus
              disabled={mutation.isPending}
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => { onOpenChange(false); setName("") }} disabled={mutation.isPending}>
              Cancel
            </Button>
            <Button type="submit" disabled={mutation.isPending || !name.trim()}>
              {mutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Creating…
                </>
              ) : (
                "Create"
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

interface DeleteTagDialogProps {
  tag: TagType | null
  onOpenChange: (open: boolean) => void
}

function DeleteTagDialog({ tag, onOpenChange }: DeleteTagDialogProps) {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: () => deleteTag(tag!.id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["tags"] })
      toast.success("Tag deleted")
      onOpenChange(false)
    },
    onError: (err: Error) => toast.error(err.message || "Failed to delete tag"),
  })

  return (
    <Dialog open={!!tag} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Delete Tag</DialogTitle>
        </DialogHeader>
        <p className="text-sm text-muted-foreground">
          Delete tag{" "}
          <span className="font-medium text-foreground">"{tag?.name}"</span>? This will detach
          it from all entries. This action cannot be undone.
        </p>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={mutation.isPending}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={() => mutation.mutate()} disabled={mutation.isPending}>
            {mutation.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Deleting…
              </>
            ) : (
              "Delete"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function TagRow({ tag, onDelete }: { tag: TagType; onDelete: (tag: TagType) => void }) {
  return (
    <div className="flex items-center justify-between rounded-lg border border-border bg-card px-4 py-3 hover:border-primary/30 transition-colors">
      <div>
        <p className="font-medium text-foreground">{tag.name}</p>
        <p className="text-xs text-muted-foreground font-mono">{tag.slug}</p>
      </div>
      <Button
        variant="ghost"
        size="icon"
        className="h-8 w-8 text-muted-foreground hover:text-destructive"
        onClick={() => onDelete(tag)}
        aria-label={`Delete tag ${tag.name}`}
      >
        <Trash2 className="h-4 w-4" />
      </Button>
    </div>
  )
}

export default function TagsPage() {
  const { isAuthenticated } = useAuth()
  const [createOpen, setCreateOpen] = useState(false)
  const [tagToDelete, setTagToDelete] = useState<TagType | null>(null)

  const { data, isLoading, error } = useQuery({
    queryKey: ["tags"],
    queryFn: () => listTags({ limit: 100 }),
    enabled: isAuthenticated,
  })

  if (!isAuthenticated) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-4 text-center">
        <Tag className="h-12 w-12 text-muted-foreground/40" />
        <p className="text-muted-foreground">Sign in to manage your tags</p>
        <div className="flex gap-3">
          <Button asChild><Link to="/login">Sign in</Link></Button>
          <Button variant="outline" asChild><Link to="/signup">Create account</Link></Button>
        </div>
      </div>
    )
  }

  const tags = data?.tags ?? []

  return (
    <div className="mx-auto w-full max-w-2xl px-4 py-8">
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Tags</h1>
          {data && (
            <p className="mt-0.5 text-sm text-muted-foreground">
              {data.total} {data.total === 1 ? "tag" : "tags"}
            </p>
          )}
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          New Tag
        </Button>
      </div>

      {/* Loading */}
      {isLoading && (
        <div className="space-y-2">
          {Array.from({ length: 6 }).map((_, i) => (
            <TagCardSkeleton key={i} />
          ))}
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="flex flex-col items-center justify-center py-20 gap-3 text-center">
          <AlertCircle className="h-10 w-10 text-destructive/60" />
          <p className="text-sm text-muted-foreground">
            {error instanceof Error ? error.message : "Failed to load tags"}
          </p>
        </div>
      )}

      {/* Empty state */}
      {!isLoading && !error && tags.length === 0 && (
        <EmptyState onNew={() => setCreateOpen(true)} />
      )}

      {/* Tag list */}
      {!isLoading && tags.length > 0 && (
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          className="space-y-2"
        >
          <AnimatePresence>
            {tags.map((tag) => (
              <motion.div key={tag.id} variants={itemVariants} layout>
                <TagRow tag={tag} onDelete={setTagToDelete} />
              </motion.div>
            ))}
          </AnimatePresence>
        </motion.div>
      )}

      <CreateTagDialog open={createOpen} onOpenChange={setCreateOpen} />
      <DeleteTagDialog tag={tagToDelete} onOpenChange={(open) => { if (!open) setTagToDelete(null) }} />
    </div>
  )
}
