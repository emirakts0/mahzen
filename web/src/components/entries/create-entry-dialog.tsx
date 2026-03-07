import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { Plus, Loader2 } from "lucide-react"
import { toast } from "sonner"
import { createEntry } from "@/api/entries"
import { listTags } from "@/api/tags"
import { useQuery } from "@tanstack/react-query"
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

  const { data: tagsData } = useQuery({
    queryKey: ["tags"],
    queryFn: () => listTags({ limit: 100 }),
  })

  const mutation = useMutation({
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

  const availableTags = tagsData?.tags ?? []

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!mutation.isPending) { onOpenChange(v); if (!v) resetForm() } }}>
      <DialogContent className="max-w-lg backdrop-blur-xl" style={{ background: "var(--glass-bg)", border: "1px solid var(--glass-border)" }}>
        <DialogHeader>
          <DialogTitle style={{ color: "var(--glass-text)" }}>New Entry</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="entry-title" style={{ color: "var(--glass-text)" }}>Title</Label>
            <Input
              id="entry-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="My Note"
              disabled={mutation.isPending}
              className="backdrop-blur-md"
              style={{ background: "var(--glass-bg)", borderColor: "var(--glass-border)", color: "var(--glass-text)" }}
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label htmlFor="entry-path" style={{ color: "var(--glass-text)" }}>Path</Label>
              <Input
                id="entry-path"
                value={path}
                onChange={(e) => setPath(e.target.value)}
                placeholder="/notes/work"
                disabled={mutation.isPending}
                className="backdrop-blur-md"
                style={{ background: "var(--glass-bg)", borderColor: "var(--glass-border)", color: "var(--glass-text)" }}
              />
            </div>
            <div className="space-y-2">
              <Label style={{ color: "var(--glass-text)" }}>Visibility</Label>
              <Select
                value={visibility}
                onValueChange={(v) => setVisibility(v as "public" | "private")}
                disabled={mutation.isPending}
              >
                <SelectTrigger className="backdrop-blur-md" style={{ background: "var(--glass-bg)", borderColor: "var(--glass-border)", color: "var(--glass-text)" }}>
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
            <Label htmlFor="entry-content" style={{ color: "var(--glass-text)" }}>Content</Label>
            <Textarea
              id="entry-content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="Write your content here..."
              rows={6}
              disabled={mutation.isPending}
              className="resize-none backdrop-blur-md"
              style={{ background: "var(--glass-bg)", borderColor: "var(--glass-border)", color: "var(--glass-text)" }}
            />
          </div>

          {availableTags.length > 0 && (
            <div className="space-y-2">
              <Label style={{ color: "var(--glass-text)" }}>Tags</Label>
              <div className="flex flex-wrap gap-1.5">
                {availableTags.map((t) => (
                  <button
                    key={t.id}
                    type="button"
                    onClick={() => toggleTag(t.id)}
                    className={cn(
                      "rounded-full border px-2.5 py-0.5 text-xs font-medium transition-colors backdrop-blur-sm",
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

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => { onOpenChange(false); resetForm() }} disabled={mutation.isPending}>
              Cancel
            </Button>
            <Button type="submit" disabled={mutation.isPending}>
              {mutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Creating...
                </>
              ) : (
                <>
                  <Plus className="mr-2 h-4 w-4" />
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
