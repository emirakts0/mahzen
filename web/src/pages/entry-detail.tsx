import { useState } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { useEntry, useDeleteEntry } from "@/hooks/use-entries"
import { EntryDeleteDialog } from "@/components/entries/entry-delete-dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { Skeleton } from "@/components/ui/skeleton"
import { VISIBILITY_LABELS } from "@/lib/constants"
import {
  Pencil,
  Trash2,
  ArrowLeft,
  Clock,
  Eye,
  EyeOff,
  Folder,
  Sparkles,
} from "lucide-react"
import { toast } from "sonner"

export function EntryDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data, isLoading, error } = useEntry(id!)
  const deleteEntry = useDeleteEntry()
  const [showDelete, setShowDelete] = useState(false)

  if (isLoading) {
    return (
      <div className="mx-auto max-w-3xl space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-64 w-full" />
      </div>
    )
  }

  if (error || !data?.entry) {
    return (
      <div className="mx-auto max-w-3xl space-y-4">
        <p className="text-destructive-foreground">Entry not found.</p>
        <Button variant="outline" onClick={() => navigate("/entries")}>
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back to entries
        </Button>
      </div>
    )
  }

  const entry = data.entry
  const isPrivate = entry.visibility !== "public"
  const created = new Date(entry.created_at)
  const updated = new Date(entry.updated_at)

  const handleDelete = () => {
    deleteEntry.mutate(entry.id, {
      onSuccess: () => {
        toast.success("Entry deleted")
        navigate("/entries")
      },
      onError: () => {
        toast.error("Failed to delete entry")
      },
    })
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      {/* Top bar */}
      <div className="flex items-center justify-between">
        <Button variant="ghost" size="sm" onClick={() => navigate("/entries")}>
          <ArrowLeft className="mr-2 h-4 w-4" />
          Entries
        </Button>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" asChild>
            <Link to={`/entries/${entry.id}/edit`}>
              <Pencil className="mr-2 h-4 w-4" />
              Edit
            </Link>
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowDelete(true)}
          >
            <Trash2 className="mr-2 h-4 w-4" />
            Delete
          </Button>
        </div>
      </div>

      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold">{entry.title}</h1>
        <div className="mt-2 flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
          <span className="flex items-center gap-1">
            {isPrivate ? (
              <EyeOff className="h-3.5 w-3.5" />
            ) : (
              <Eye className="h-3.5 w-3.5" />
            )}
            {VISIBILITY_LABELS[entry.visibility]}
          </span>
          {entry.path && entry.path !== "/" && (
            <span className="flex items-center gap-1">
              <Folder className="h-3.5 w-3.5" />
              {entry.path}
            </span>
          )}
          <span className="flex items-center gap-1">
            <Clock className="h-3.5 w-3.5" />
            Created {created.toLocaleDateString()}
          </span>
          {entry.updated_at !== entry.created_at && (
            <span className="flex items-center gap-1">
              <Clock className="h-3.5 w-3.5" />
              Updated {updated.toLocaleDateString()}
            </span>
          )}
        </div>
      </div>

      {/* Tags */}
      {entry.tags && entry.tags.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {entry.tags.map((tag) => (
            <Badge key={tag} variant="secondary">
              {tag}
            </Badge>
          ))}
        </div>
      )}

      {/* Summary */}
      {entry.summary && (
        <div className="rounded-lg border bg-muted/50 p-4">
          <div className="mb-1 flex items-center gap-1.5 text-xs font-medium text-muted-foreground">
            <Sparkles className="h-3 w-3" />
            AI Summary
          </div>
          <p className="text-sm">{entry.summary}</p>
        </div>
      )}

      <Separator />

      {/* Content */}
      <div className="whitespace-pre-wrap font-mono text-sm leading-relaxed">
        {entry.content}
      </div>

      {/* Delete dialog */}
      <EntryDeleteDialog
        open={showDelete}
        onOpenChange={setShowDelete}
        onConfirm={handleDelete}
        isDeleting={deleteEntry.isPending}
        entryTitle={entry.title}
      />
    </div>
  )
}
