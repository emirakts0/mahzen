import { useState, type FormEvent } from "react"
import { useTags, useCreateTag, useDeleteTag } from "@/hooks/use-tags"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Plus, Trash2 } from "lucide-react"
import { toast } from "sonner"

export function TagsPage() {
  const { data, isLoading } = useTags({ limit: 100 })
  const createTag = useCreateTag()
  const deleteTag = useDeleteTag()
  const [newTagName, setNewTagName] = useState("")
  const [deleteTarget, setDeleteTarget] = useState<{
    id: string
    name: string
  } | null>(null)

  const handleCreate = (e: FormEvent) => {
    e.preventDefault()
    if (!newTagName.trim()) return
    createTag.mutate(
      { name: newTagName.trim() },
      {
        onSuccess: () => {
          toast.success("Tag created")
          setNewTagName("")
        },
        onError: () => {
          toast.error("Failed to create tag")
        },
      }
    )
  }

  const handleDelete = () => {
    if (!deleteTarget) return
    deleteTag.mutate(deleteTarget.id, {
      onSuccess: () => {
        toast.success("Tag deleted")
        setDeleteTarget(null)
      },
      onError: () => {
        toast.error("Failed to delete tag")
      },
    })
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <h1 className="text-2xl font-bold">Tags</h1>

      {/* Create form */}
      <form onSubmit={handleCreate} className="flex gap-2">
        <Input
          placeholder="New tag name"
          value={newTagName}
          onChange={(e) => setNewTagName(e.target.value)}
          className="max-w-sm"
        />
        <Button type="submit" disabled={createTag.isPending || !newTagName.trim()}>
          <Plus className="mr-2 h-4 w-4" />
          Create
        </Button>
      </form>

      {/* Tag list */}
      {isLoading ? (
        <div className="flex flex-wrap gap-2">
          {Array.from({ length: 8 }).map((_, i) => (
            <Skeleton key={i} className="h-8 w-20 rounded-md" />
          ))}
        </div>
      ) : !data?.tags || data.tags.length === 0 ? (
        <p className="text-muted-foreground">No tags yet. Create one above.</p>
      ) : (
        <div className="flex flex-wrap gap-2">
          {data.tags.map((tag) => (
            <div
              key={tag.id}
              className="group flex items-center gap-1"
            >
              <Badge variant="secondary" className="text-sm py-1.5 px-3">
                {tag.name}
              </Badge>
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity"
                onClick={() =>
                  setDeleteTarget({ id: tag.id, name: tag.name })
                }
              >
                <Trash2 className="h-3 w-3" />
              </Button>
            </div>
          ))}
        </div>
      )}

      {data && (
        <p className="text-sm text-muted-foreground">
          {data.total} tag{data.total !== 1 ? "s" : ""} total
        </p>
      )}

      {/* Delete dialog */}
      <Dialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete tag</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete the tag &quot;
              {deleteTarget?.name}&quot;? This will remove it from all entries.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={deleteTag.isPending}
            >
              {deleteTag.isPending ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
