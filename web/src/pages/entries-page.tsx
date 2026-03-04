import { useState } from "react"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { Link } from "react-router"
import { motion, AnimatePresence } from "framer-motion"
import { Plus, BookOpen, ChevronRight, Eye, EyeOff, Loader2, AlertCircle, Tag } from "lucide-react"
import { toast } from "sonner"
import { listEntries, createEntry, deleteEntry } from "@/api/entries"
import { listTags } from "@/api/tags"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useAuth } from "@/hooks/use-auth"
import { cn } from "@/lib/utils"
import type { CreateEntryRequest, Entry } from "@/types/api"

const PAGE_SIZE = 20

const containerVariants = {
  hidden: {},
  visible: { transition: { staggerChildren: 0.05 } },
}

const itemVariants = {
  hidden: { opacity: 0, y: 10 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.2 } },
}

function EntryCardSkeleton() {
  return (
    <div className="rounded-lg border border-border bg-card p-4 space-y-2">
      <div className="flex items-center justify-between">
        <Skeleton className="h-5 w-48" />
        <Skeleton className="h-4 w-16" />
      </div>
      <Skeleton className="h-3 w-32" />
      <Skeleton className="h-3 w-full" />
      <div className="flex gap-1 pt-1">
        <Skeleton className="h-5 w-12 rounded-full" />
        <Skeleton className="h-5 w-14 rounded-full" />
      </div>
    </div>
  )
}

function EmptyState({ onNew }: { onNew: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-24 text-center">
      <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/10">
        <BookOpen className="h-8 w-8 text-primary" />
      </div>
      <h2 className="text-xl font-semibold text-foreground">No entries yet</h2>
      <p className="mt-1 text-sm text-muted-foreground">Create your first knowledge base entry</p>
      <Button className="mt-6" onClick={onNew}>
        <Plus className="mr-2 h-4 w-4" />
        New Entry
      </Button>
    </div>
  )
}

interface CreateEntryDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  availableTags: Array<{ id: string; name: string }>
  onCreated: () => void
}

function CreateEntryDialog({ open, onOpenChange, availableTags, onCreated }: CreateEntryDialogProps) {
  const queryClient = useQueryClient()
  const [title, setTitle] = useState("")
  const [content, setContent] = useState("")
  const [path, setPath] = useState("/")
  const [visibility, setVisibility] = useState<"public" | "private">("private")
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>([])

  const mutation = useMutation({
    mutationFn: (data: CreateEntryRequest) => createEntry(data),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["entries"] })
      toast.success("Entry created")
      onCreated()
      resetForm()
    },
    onError: (err: Error) => {
      toast.error(err.message || "Failed to create entry")
    },
  })

  const resetForm = () => {
    setTitle("")
    setContent("")
    setPath("/")
    setVisibility("private")
    setSelectedTagIds([])
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    mutation.mutate({ title, content, path, visibility, tag_ids: selectedTagIds })
  }

  const toggleTag = (id: string) => {
    setSelectedTagIds((prev) =>
      prev.includes(id) ? prev.filter((t) => t !== id) : [...prev, id],
    )
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!mutation.isPending) { onOpenChange(v); if (!v) resetForm() } }}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>New Entry</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="entry-title">Title</Label>
            <Input
              id="entry-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="My Note"
              disabled={mutation.isPending}
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label htmlFor="entry-path">Path</Label>
              <Input
                id="entry-path"
                value={path}
                onChange={(e) => setPath(e.target.value)}
                placeholder="/notes/work"
                disabled={mutation.isPending}
              />
            </div>
            <div className="space-y-2">
              <Label>Visibility</Label>
              <Select
                value={visibility}
                onValueChange={(v) => setVisibility(v as "public" | "private")}
                disabled={mutation.isPending}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="public">Public</SelectItem>
                  <SelectItem value="private">Private</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="entry-content">Content</Label>
            <Textarea
              id="entry-content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="Write your content here..."
              rows={6}
              disabled={mutation.isPending}
              className="resize-none"
            />
          </div>

          {availableTags.length > 0 && (
            <div className="space-y-2">
              <Label>Tags</Label>
              <div className="flex flex-wrap gap-1.5">
                {availableTags.map((t) => (
                  <button
                    key={t.id}
                    type="button"
                    onClick={() => toggleTag(t.id)}
                    className={cn(
                      "rounded-full border px-2.5 py-0.5 text-xs font-medium transition-colors",
                      selectedTagIds.includes(t.id)
                        ? "border-primary bg-primary text-primary-foreground"
                        : "border-border bg-background text-muted-foreground hover:border-primary/50",
                    )}
                  >
                    {t.name}
                  </button>
                ))}
              </div>
            </div>
          )}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => { onOpenChange(false); resetForm() }} disabled={mutation.isPending}>
              Cancel
            </Button>
            <Button type="submit" disabled={mutation.isPending}>
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

function EntryCard({ entry }: { entry: Entry }) {
  return (
    <Link to={`/entries/${entry.id}`} className="block group">
      <div className="rounded-lg border border-border bg-card p-4 transition-colors hover:border-primary/40 hover:bg-accent/40">
        <div className="flex items-start justify-between gap-2">
          <h3 className="font-semibold text-foreground group-hover:text-primary transition-colors line-clamp-1">
            {entry.title || "(Untitled)"}
          </h3>
          <div className="flex items-center gap-1 shrink-0">
            {entry.visibility === "public" ? (
              <Eye className="h-3.5 w-3.5 text-muted-foreground" />
            ) : (
              <EyeOff className="h-3.5 w-3.5 text-muted-foreground" />
            )}
            <span className="text-xs text-muted-foreground">{entry.visibility}</span>
          </div>
        </div>

        <div className="mt-1 flex items-center gap-1 text-xs text-muted-foreground">
          <span className="truncate">{entry.path || "/"}</span>
          <ChevronRight className="h-3 w-3 shrink-0" />
        </div>

        {entry.summary && (
          <p className="mt-1.5 text-sm text-muted-foreground line-clamp-2">{entry.summary}</p>
        )}

        {entry.tags && entry.tags.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1">
            {entry.tags.map((tag) => (
              <Badge key={tag} variant="secondary" className="text-xs px-1.5 py-0">
                {tag}
              </Badge>
            ))}
          </div>
        )}

        <p className="mt-2 text-xs text-muted-foreground">
          {new Date(entry.updated_at).toLocaleDateString()}
        </p>
      </div>
    </Link>
  )
}

export default function EntriesPage() {
  const { isAuthenticated } = useAuth()
  const [createOpen, setCreateOpen] = useState(false)

  const { data, isLoading, error } = useQuery({
    queryKey: ["entries"],
    queryFn: () => listEntries({ limit: PAGE_SIZE, offset: 0 }),
    enabled: isAuthenticated,
  })

  const { data: tagsData } = useQuery({
    queryKey: ["tags"],
    queryFn: () => listTags({ limit: 100 }),
    enabled: isAuthenticated,
  })

  const queryClient = useQueryClient()

  const deleteMutation = useMutation({
    mutationFn: deleteEntry,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["entries"] })
      toast.success("Entry deleted")
    },
    onError: (err: Error) => toast.error(err.message || "Failed to delete entry"),
  })
  void deleteMutation // suppress unused warning — used via entry detail page

  if (!isAuthenticated) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-4 text-center">
        <BookOpen className="h-12 w-12 text-muted-foreground/40" />
        <p className="text-muted-foreground">Sign in to view your entries</p>
        <div className="flex gap-3">
          <Button asChild><Link to="/login">Sign in</Link></Button>
          <Button variant="outline" asChild><Link to="/signup">Create account</Link></Button>
        </div>
      </div>
    )
  }

  const entries = data?.entries ?? []
  const availableTags = tagsData?.tags ?? []

  return (
    <div className="mx-auto w-full max-w-4xl px-4 py-8">
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Entries</h1>
          {data && (
            <p className="mt-0.5 text-sm text-muted-foreground">
              {data.total} {data.total === 1 ? "entry" : "entries"}
            </p>
          )}
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          New Entry
        </Button>
      </div>

      {/* Loading */}
      {isLoading && (
        <div className="space-y-3">
          {Array.from({ length: 5 }).map((_, i) => (
            <EntryCardSkeleton key={i} />
          ))}
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="flex flex-col items-center justify-center py-20 gap-3 text-center">
          <AlertCircle className="h-10 w-10 text-destructive/60" />
          <p className="text-sm text-muted-foreground">{error instanceof Error ? error.message : "Failed to load entries"}</p>
        </div>
      )}

      {/* Empty state */}
      {!isLoading && !error && entries.length === 0 && (
        <EmptyState onNew={() => setCreateOpen(true)} />
      )}

      {/* Entry list */}
      {!isLoading && entries.length > 0 && (
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          className="space-y-3"
        >
          <AnimatePresence>
            {entries.map((entry) => (
              <motion.div key={entry.id} variants={itemVariants} layout>
                <EntryCard entry={entry} />
              </motion.div>
            ))}
          </AnimatePresence>
        </motion.div>
      )}

      {/* Tags shortcut if no tags yet */}
      {!isLoading && availableTags.length === 0 && entries.length > 0 && (
        <p className="mt-6 text-center text-xs text-muted-foreground">
          <Tag className="inline h-3 w-3 mr-1" />
          Create tags on the{" "}
          <Link to="/tags" className="underline underline-offset-2 hover:text-foreground">
            Tags page
          </Link>{" "}
          to organize entries.
        </p>
      )}

      <CreateEntryDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        availableTags={availableTags}
        onCreated={() => setCreateOpen(false)}
      />
    </div>
  )
}
