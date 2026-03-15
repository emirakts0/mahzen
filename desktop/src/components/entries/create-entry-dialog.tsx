import { useState } from "react"
import { useMutation, useQueryClient, useQuery } from "@tanstack/react-query"
import { Plus, Loader2 } from "lucide-react"
import { toast } from "sonner"
import { createEntry } from "@/api/entries"
import { listTags, createTag } from "@/api/tags"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { cn } from "@/lib/utils"
import type { CreateEntryRequest } from "@/types/api"

interface CreateEntryDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  defaultPath?: string
}

export function CreateEntryDialog({ open, onOpenChange, defaultPath = "/" }: CreateEntryDialogProps) {
  const queryClient = useQueryClient()
  const [title, setTitle] = useState("")
  const [content, setContent] = useState("")
  const [path, setPath] = useState(defaultPath)
  const [visibility, setVisibility] = useState<"public" | "private">("private")
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>([])
  const [newTagName, setNewTagName] = useState("")

  const { data: tagsData } = useQuery({
    queryKey: ["tags"],
    queryFn: () => listTags({ limit: 100 }),
  })

  const createTagMutation = useMutation({
    mutationFn: (name: string) => createTag({ name }),
    onSuccess: (newTag) => {
      void queryClient.invalidateQueries({ queryKey: ["tags"] })
      setSelectedTagIds((prev) => [...prev, newTag.id])
      setNewTagName("")
      toast.success("Tag created")
    },
    onError: (err: Error) => {
      toast.error(err.message || "Failed to create tag")
    },
  })

  const entryMutation = useMutation({
    mutationFn: (data: CreateEntryRequest) => createEntry(data),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["entries"] })
      void queryClient.invalidateQueries({ queryKey: ["folders"] })
      toast.success("Entry created")
      onOpenChange(false)
      resetForm()
    },
    onError: (err: Error) => {
      toast.error(err.message || "Failed to create entry")
    },
  })

  const resetForm = () => {
    setTitle("")
    setContent("")
    setPath(defaultPath)
    setVisibility("private")
    setSelectedTagIds([])
    setNewTagName("")
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    entryMutation.mutate({ title, content, path, visibility, tag_ids: selectedTagIds })
  }

  const toggleTag = (id: string) => {
    setSelectedTagIds((prev) =>
      prev.includes(id) ? prev.filter((t) => t !== id) : [...prev, id],
    )
  }

  const handleCreateTag = () => {
    const name = newTagName.trim()
    if (name) {
      createTagMutation.mutate(name)
    }
  }

  const availableTags = tagsData?.tags ?? []

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!entryMutation.isPending) { onOpenChange(v); if (!v) resetForm() } }}>
      <DialogContent className="max-w-md max-h-[75vh] flex flex-col overflow-hidden backdrop-blur-xl p-0" style={{ background: "var(--glass-bg)", border: "1px solid var(--glass-border)" }}>
        <DialogHeader className="px-4 pt-4 pb-2 shrink-0">
          <DialogTitle className="text-sm" style={{ color: "var(--glass-text)" }}>New Entry</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-3 overflow-y-auto px-4 pb-4 flex-1">
          <div className="space-y-1.5">
            <Label htmlFor="entry-title" className="text-[10px]" style={{ color: "var(--glass-text)" }}>Title</Label>
            <Input
              id="entry-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="My Note"
              disabled={entryMutation.isPending}
              className="backdrop-blur-md h-8 text-xs"
              style={{ background: "var(--glass-bg)", borderColor: "var(--glass-border)", color: "var(--glass-text)" }}
            />
          </div>

          <div className="grid grid-cols-2 gap-2">
            <div className="space-y-1.5">
              <Label htmlFor="entry-path" className="text-[10px]" style={{ color: "var(--glass-text)" }}>Path</Label>
              <Input
                id="entry-path"
                value={path}
                onChange={(e) => setPath(e.target.value)}
                placeholder="/notes/work"
                disabled={entryMutation.isPending}
                className="backdrop-blur-md h-8 text-xs"
                style={{ background: "var(--glass-bg)", borderColor: "var(--glass-border)", color: "var(--glass-text)" }}
              />
            </div>
            <div className="space-y-1.5">
              <Label className="text-[10px]" style={{ color: "var(--glass-text)" }}>Visibility</Label>
              <Select
                value={visibility}
                onValueChange={(v) => setVisibility(v as "public" | "private")}
                disabled={entryMutation.isPending}
              >
                <SelectTrigger className="backdrop-blur-md h-8 text-xs" style={{ background: "var(--glass-bg)", borderColor: "var(--glass-border)", color: "var(--glass-text)" }}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="public">Public</SelectItem>
                  <SelectItem value="private">Private</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="entry-content" className="text-[10px]" style={{ color: "var(--glass-text)" }}>Content</Label>
            <Textarea
              id="entry-content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="Write your content here..."
              rows={4}
              disabled={entryMutation.isPending}
              className="resize-none backdrop-blur-md text-xs"
              style={{ background: "var(--glass-bg)", borderColor: "var(--glass-border)", color: "var(--glass-text)" }}
            />
          </div>

          {availableTags.length > 0 && (
            <div className="space-y-1.5">
              <Label className="text-[10px]" style={{ color: "var(--glass-text)" }}>Tags</Label>
              <div className="flex flex-wrap gap-1">
                {availableTags.map((t) => (
                  <button
                    key={t.id}
                    type="button"
                    onClick={() => toggleTag(t.id)}
                    className={cn(
                      "rounded-full border px-2 py-0.5 text-[10px] font-medium transition-colors backdrop-blur-sm",
                      selectedTagIds.includes(t.id)
                        ? "border-primary bg-primary text-primary-foreground"
                        : "border-[var(--glass-border)] bg-[var(--glass-bg)] text-[var(--glass-text-muted)] hover:border-primary/50",
                    )}
                  >
                    {t.name}
                  </button>
                ))}
              </div>
            </div>
          )}

          <div className="space-y-1.5">
            <Label className="text-[10px]" style={{ color: "var(--glass-text)" }}>Add Tag</Label>
            <div className="flex gap-2">
              <Input
                value={newTagName}
                onChange={(e) => setNewTagName(e.target.value)}
                placeholder="New tag name"
                onKeyDown={(e) => e.key === "Enter" && (e.preventDefault(), handleCreateTag())}
                className="backdrop-blur-md h-8 text-xs"
                style={{ background: "var(--glass-bg)", borderColor: "var(--glass-border)", color: "var(--glass-text)" }}
              />
              <Button
                type="button"
                onClick={handleCreateTag}
                disabled={!newTagName.trim() || createTagMutation.isPending}
                size="sm"
                className="shrink-0 h-8"
              >
                {createTagMutation.isPending ? <Loader2 className="h-3 w-3 animate-spin" /> : <Plus className="h-3 w-3" />}
              </Button>
            </div>
          </div>

          <DialogFooter className="shrink-0 pt-2 border-t" style={{ borderColor: "var(--glass-border)" }}>
            <Button type="button" variant="outline" onClick={() => { onOpenChange(false); resetForm() }} disabled={entryMutation.isPending} size="sm" className="h-8 text-xs">
              Cancel
            </Button>
            <Button type="submit" disabled={entryMutation.isPending} size="sm" className="h-8 text-xs">
              {entryMutation.isPending ? (
                <>
                  <Loader2 className="mr-1.5 h-3 w-3 animate-spin" />
                  Creating...
                </>
              ) : (
                <>
                  <Plus className="mr-1.5 h-3 w-3" />
                  Create
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
