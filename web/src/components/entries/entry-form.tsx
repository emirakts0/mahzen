import { useState, type FormEvent } from "react"
import { useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import type {
  CreateEntryRequest,
  Entry,
  UpdateEntryRequest,
  Visibility,
} from "@/lib/types"

interface EntryFormProps {
  entry?: Entry
  onSubmit: (data: CreateEntryRequest | UpdateEntryRequest) => void
  isSubmitting: boolean
  submitLabel?: string
}

export function EntryForm({
  entry,
  onSubmit,
  isSubmitting,
  submitLabel = "Save",
}: EntryFormProps) {
  const navigate = useNavigate()
  const [title, setTitle] = useState(entry?.title ?? "")
  const [content, setContent] = useState(entry?.content ?? "")
  const [path, setPath] = useState(entry?.path ?? "/")
  const [visibility, setVisibility] = useState<Visibility>(
    entry?.visibility ?? "private"
  )

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    onSubmit({
      title,
      content,
      path: path || "/",
      visibility,
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="title">Title</Label>
        <Input
          id="title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Entry title"
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="content">Content</Label>
        <Textarea
          id="content"
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="Write your content..."
          required
          rows={12}
          className="font-mono"
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="path">Path</Label>
          <Input
            id="path"
            value={path}
            onChange={(e) => setPath(e.target.value)}
            placeholder="/"
          />
          <p className="text-xs text-muted-foreground">
            Hierarchical path, e.g. /notes/work
          </p>
        </div>

        <div className="space-y-2">
          <Label>Visibility</Label>
          <Select
            value={visibility}
            onValueChange={(v) => setVisibility(v as Visibility)}
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

      <div className="flex justify-end gap-3">
        <Button
          type="button"
          variant="outline"
          onClick={() => navigate(-1)}
        >
          Cancel
        </Button>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Saving..." : submitLabel}
        </Button>
      </div>
    </form>
  )
}
