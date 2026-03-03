import { Link } from "react-router-dom"
import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { VISIBILITY_LABELS } from "@/lib/constants"
import type { Entry } from "@/lib/types"
import { Clock, Eye, EyeOff, Folder } from "lucide-react"

interface EntryCardProps {
  entry: Entry
}

export function EntryCard({ entry }: EntryCardProps) {
  const isPrivate = entry.visibility !== "public"
  const date = new Date(entry.updated_at || entry.created_at)

  return (
    <Link to={`/entries/${entry.id}`}>
      <Card className="transition-colors hover:bg-accent/50">
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between gap-2">
            <CardTitle className="text-base leading-snug">
              {entry.title}
            </CardTitle>
            <div className="flex items-center gap-1 text-muted-foreground">
              {isPrivate ? (
                <EyeOff className="h-3.5 w-3.5" />
              ) : (
                <Eye className="h-3.5 w-3.5" />
              )}
              <span className="text-xs">
                {VISIBILITY_LABELS[entry.visibility]}
              </span>
            </div>
          </div>
          {entry.path && entry.path !== "/" && (
            <CardDescription className="flex items-center gap-1 text-xs">
              <Folder className="h-3 w-3" />
              {entry.path}
            </CardDescription>
          )}
        </CardHeader>
        <CardContent className="pt-0">
          {entry.summary && (
            <p className="mb-3 text-sm text-muted-foreground line-clamp-2">
              {entry.summary}
            </p>
          )}
          {!entry.summary && entry.content && (
            <p className="mb-3 text-sm text-muted-foreground line-clamp-2">
              {entry.content}
            </p>
          )}
          <div className="flex items-center justify-between">
            <div className="flex flex-wrap gap-1">
              {entry.tags?.map((tag) => (
                <Badge key={tag} variant="secondary" className="text-xs">
                  {tag}
                </Badge>
              ))}
            </div>
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <Clock className="h-3 w-3" />
              {date.toLocaleDateString()}
            </div>
          </div>
        </CardContent>
      </Card>
    </Link>
  )
}
