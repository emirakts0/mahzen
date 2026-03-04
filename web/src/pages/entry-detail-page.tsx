import { useState } from "react"
import { useParams, useNavigate, Link } from "react-router"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { motion } from "framer-motion"
import {
  ArrowLeft,
  Edit,
  Trash2,
  Eye,
  EyeOff,
  Copy,
  Download,
  Loader2,
  AlertCircle,
  Tag,
  X,
} from "lucide-react"
import { toast } from "sonner"
import { getEntry, updateEntry, deleteEntry, attachTag, detachTag } from "@/api/entries"
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
import type { Entry, UpdateEntryRequest } from "@/types/api"

// ─────────────────────────────────────────────
// Edit Entry Dialog
// ─────────────────────────────────────────────

interface EditEntryDialogProps {
  entry: Entry
  availableTags: Array<{ id: string; name: string; slug: string }>
  open: boolean
  onOpenChange: (open: boolean) => void
}

function EditEntryDialog({ entry, availableTags, open, onOpenChange }: EditEntryDialogProps) {
  const queryClient = useQueryClient()
  const [title, setTitle] = useState(entry.title)
  const [content, setContent] = useState(entry.content)
  const [path, setPath] = useState(entry.path)
  const [visibility, setVisibility] = useState<"public" | "private">(entry.visibility)

  // Resolve current tag IDs from tag names (best effort)
  const currentTagIds = (entry.tags ?? []).flatMap((name) => {
    const found = availableTags.find((t) => t.name === name)
    return found ? [found.id] : []
  })
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>(currentTagIds)

  const mutation = useMutation({
    mutationFn: (data: UpdateEntryRequest) => updateEntry(entry.id, data),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["entry", entry.id] })
      void queryClient.invalidateQueries({ queryKey: ["entries"] })
      toast.success("Entry updated")
      onOpenChange(false)
    },
    onError: (err: Error) => toast.error(err.message || "Failed to update entry"),
  })

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
    <Dialog open={open} onOpenChange={(v) => { if (!mutation.isPending) onOpenChange(v) }}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Edit Entry</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-title">Title</Label>
            <Input
              id="edit-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="My Note"
              disabled={mutation.isPending}
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label htmlFor="edit-path">Path</Label>
              <Input
                id="edit-path"
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
            <Label htmlFor="edit-content">Content</Label>
            <Textarea
              id="edit-content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              rows={8}
              disabled={mutation.isPending}
              className="resize-none font-mono text-sm"
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
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={mutation.isPending}>
              Cancel
            </Button>
            <Button type="submit" disabled={mutation.isPending}>
              {mutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Saving…
                </>
              ) : (
                "Save"
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// ─────────────────────────────────────────────
// Tag attach/detach inline widget
// ─────────────────────────────────────────────

interface TagManagerProps {
  entry: Entry
  availableTags: Array<{ id: string; name: string; slug: string }>
}

function TagManager({ entry, availableTags }: TagManagerProps) {
  const queryClient = useQueryClient()
  const [addOpen, setAddOpen] = useState(false)

  const attachMutation = useMutation({
    mutationFn: (tagId: string) => attachTag(entry.id, tagId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["entry", entry.id] })
      toast.success("Tag attached")
      setAddOpen(false)
    },
    onError: (err: Error) => toast.error(err.message || "Failed to attach tag"),
  })

  const detachMutation = useMutation({
    mutationFn: (tagId: string) => detachTag(entry.id, tagId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["entry", entry.id] })
      toast.success("Tag removed")
    },
    onError: (err: Error) => toast.error(err.message || "Failed to remove tag"),
  })

  const currentTagNames = new Set(entry.tags ?? [])
  const unattachedTags = availableTags.filter((t) => !currentTagNames.has(t.name))

  const handleDetach = (tagName: string) => {
    const found = availableTags.find((t) => t.name === tagName)
    if (found) detachMutation.mutate(found.id)
  }

  return (
    <div className="flex flex-wrap gap-1.5 items-center">
      {(entry.tags ?? []).map((tagName) => (
        <span
          key={tagName}
          className="inline-flex items-center gap-1 rounded-full border border-border bg-secondary px-2.5 py-0.5 text-xs font-medium text-secondary-foreground"
        >
          {tagName}
          <button
            type="button"
            aria-label={`Remove tag ${tagName}`}
            onClick={() => handleDetach(tagName)}
            className="rounded-full hover:text-destructive transition-colors disabled:opacity-50"
            disabled={detachMutation.isPending}
          >
            <X className="h-3 w-3" />
          </button>
        </span>
      ))}

      {/* Add tag button */}
      {unattachedTags.length > 0 && (
        <div className="relative">
          {!addOpen ? (
            <button
              type="button"
              onClick={() => setAddOpen(true)}
              className="inline-flex items-center gap-1 rounded-full border border-dashed border-border px-2 py-0.5 text-xs text-muted-foreground hover:border-primary hover:text-primary transition-colors"
            >
              <Tag className="h-3 w-3" />
              Add tag
            </button>
          ) : (
            <div className="absolute left-0 top-6 z-10 flex flex-wrap gap-1 rounded-lg border border-border bg-popover p-2 shadow-md min-w-[160px]">
              {unattachedTags.map((t) => (
                <button
                  key={t.id}
                  type="button"
                  onClick={() => attachMutation.mutate(t.id)}
                  disabled={attachMutation.isPending}
                  className="rounded-full border border-border px-2 py-0.5 text-xs hover:border-primary hover:bg-primary/10 transition-colors disabled:opacity-50"
                >
                  {t.name}
                </button>
              ))}
              <button
                type="button"
                onClick={() => setAddOpen(false)}
                className="mt-1 w-full text-center text-xs text-muted-foreground hover:text-foreground"
              >
                Close
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// ─────────────────────────────────────────────
// Copy / Download helpers
// ─────────────────────────────────────────────

// The backend returns content inline for small entries.
// For entries > 64 KB the content was stored in S3 but is fetched
// and returned by GET /v1/entries/:id, so `content` is always populated
// on a single-entry fetch. We always show Copy if content is present.
function ContentActions({ content, title }: { content: string; title: string }) {
  const [copying, setCopying] = useState(false)

  const handleCopy = async () => {
    if (!content) return
    setCopying(true)
    try {
      await navigator.clipboard.writeText(content)
      toast.success("Copied to clipboard")
    } catch {
      toast.error("Failed to copy")
    } finally {
      setCopying(false)
    }
  }

  const handleDownload = () => {
    if (!content) return
    const blob = new Blob([content], { type: "text/plain;charset=utf-8" })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = `${title || "entry"}.txt`
    a.click()
    URL.revokeObjectURL(url)
  }

  if (!content) return null

  return (
    <div className="flex gap-2">
      <Button variant="outline" size="sm" onClick={() => void handleCopy()} disabled={copying}>
        <Copy className="mr-1.5 h-3.5 w-3.5" />
        Copy
      </Button>
      <Button variant="outline" size="sm" onClick={handleDownload}>
        <Download className="mr-1.5 h-3.5 w-3.5" />
        Download
      </Button>
    </div>
  )
}

// ─────────────────────────────────────────────
// Entry Detail Page
// ─────────────────────────────────────────────

function DetailSkeleton() {
  return (
    <div className="space-y-4">
      <Skeleton className="h-8 w-64" />
      <div className="flex gap-2">
        <Skeleton className="h-5 w-20" />
        <Skeleton className="h-5 w-16" />
      </div>
      <Skeleton className="h-4 w-40" />
      <Skeleton className="h-[300px] w-full" />
    </div>
  )
}

export default function EntryDetailPage() {
  const { entryId } = useParams<{ entryId: string }>()
  const navigate = useNavigate()
  const { isAuthenticated, user } = useAuth()
  const queryClient = useQueryClient()
  const [editOpen, setEditOpen] = useState(false)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)

  const { data: entry, isLoading, error } = useQuery({
    queryKey: ["entry", entryId],
    queryFn: () => getEntry(entryId!),
    enabled: !!entryId,
  })

  const { data: tagsData } = useQuery({
    queryKey: ["tags"],
    queryFn: () => listTags({ limit: 100 }),
    enabled: isAuthenticated,
  })

  const deleteMutation = useMutation({
    mutationFn: () => deleteEntry(entryId!),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["entries"] })
      toast.success("Entry deleted")
      void navigate("/entries")
    },
    onError: (err: Error) => toast.error(err.message || "Failed to delete entry"),
  })

  const availableTags = tagsData?.tags ?? []
  const isOwner = isAuthenticated && entry?.user_id === user?.id

  if (isLoading) {
    return (
      <div className="mx-auto w-full max-w-3xl px-4 py-8">
        <Skeleton className="mb-6 h-8 w-24" />
        <DetailSkeleton />
      </div>
    )
  }

  if (error || !entry) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-4 text-center">
        <AlertCircle className="h-12 w-12 text-destructive/60" />
        <p className="text-muted-foreground">
          {error instanceof Error ? error.message : "Entry not found"}
        </p>
        <Button variant="outline" asChild>
          <Link to="/entries">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Entries
          </Link>
        </Button>
      </div>
    )
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2 }}
      className="mx-auto w-full max-w-3xl px-4 py-8"
    >
      {/* Back link */}
      <Link
        to="/entries"
        className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        <ArrowLeft className="h-4 w-4" />
        Entries
      </Link>

      {/* Title + actions */}
      <div className="mt-4 flex items-start justify-between gap-4">
        <h1 className="text-2xl font-bold text-foreground leading-tight">
          {entry.title || "(Untitled)"}
        </h1>
        {isOwner && (
          <div className="flex shrink-0 items-center gap-2">
            <Button variant="outline" size="sm" onClick={() => setEditOpen(true)}>
              <Edit className="mr-1.5 h-3.5 w-3.5" />
              Edit
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="text-destructive hover:text-destructive"
              onClick={() => setDeleteConfirmOpen(true)}
            >
              <Trash2 className="mr-1.5 h-3.5 w-3.5" />
              Delete
            </Button>
          </div>
        )}
      </div>

      {/* Meta row */}
      <div className="mt-3 flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
        <span className="font-mono">{entry.path || "/"}</span>
        <span>·</span>
        <span className="flex items-center gap-1">
          {entry.visibility === "public" ? (
            <Eye className="h-3 w-3" />
          ) : (
            <EyeOff className="h-3 w-3" />
          )}
          {entry.visibility}
        </span>
        <span>·</span>
        <span>Updated {new Date(entry.updated_at).toLocaleDateString()}</span>
        <span>·</span>
        <span>Created {new Date(entry.created_at).toLocaleDateString()}</span>
      </div>

      {/* Tags */}
      <div className="mt-3">
        {isOwner ? (
          <TagManager entry={entry} availableTags={availableTags} />
        ) : (
          <div className="flex flex-wrap gap-1.5">
            {(entry.tags ?? []).map((tag) => (
              <Badge key={tag} variant="secondary" className="text-xs">
                {tag}
              </Badge>
            ))}
          </div>
        )}
      </div>

      {/* Summary */}
      {entry.summary && (
        <div className="mt-4 rounded-lg border border-border bg-muted/40 px-4 py-3">
          <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground mb-1">
            Summary
          </p>
          <p className="text-sm text-foreground">{entry.summary}</p>
        </div>
      )}

      {/* Content actions */}
      {entry.content && (
        <div className="mt-4">
          <ContentActions content={entry.content} title={entry.title} />
        </div>
      )}

      {/* Content body */}
      <div className="mt-4 rounded-lg border border-border bg-card p-4">
        {entry.content ? (
          <pre className="whitespace-pre-wrap break-words font-mono text-sm text-foreground">
            {entry.content}
          </pre>
        ) : (
          <p className="text-sm text-muted-foreground italic">No content</p>
        )}
      </div>

      {/* Edit dialog */}
      {isOwner && editOpen && (
        <EditEntryDialog
          entry={entry}
          availableTags={availableTags}
          open={editOpen}
          onOpenChange={setEditOpen}
        />
      )}

      {/* Delete confirm dialog */}
      <Dialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Delete Entry</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            Are you sure you want to delete{" "}
            <span className="font-medium text-foreground">{entry.title || "this entry"}</span>?
            This action cannot be undone.
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteConfirmOpen(false)} disabled={deleteMutation.isPending}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={() => deleteMutation.mutate()}
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? (
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
    </motion.div>
  )
}
