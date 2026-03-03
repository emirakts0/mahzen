import { EntryCard } from "./entry-card"
import { Skeleton } from "@/components/ui/skeleton"
import type { Entry } from "@/lib/types"

interface EntryListProps {
  entries: Entry[] | undefined
  isLoading: boolean
  emptyMessage?: string
}

export function EntryList({
  entries,
  isLoading,
  emptyMessage = "No entries found.",
}: EntryListProps) {
  if (isLoading) {
    return (
      <div className="grid gap-3">
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-32 w-full rounded-xl" />
        ))}
      </div>
    )
  }

  if (!entries || entries.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
        <p>{emptyMessage}</p>
      </div>
    )
  }

  return (
    <div className="grid gap-3">
      {entries.map((entry) => (
        <EntryCard key={entry.id} entry={entry} />
      ))}
    </div>
  )
}
