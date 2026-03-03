import { useParams, useNavigate } from "react-router-dom"
import { useEntry, useUpdateEntry } from "@/hooks/use-entries"
import { EntryForm } from "@/components/entries/entry-form"
import { Skeleton } from "@/components/ui/skeleton"
import { Button } from "@/components/ui/button"
import type { UpdateEntryRequest } from "@/lib/types"
import { ArrowLeft } from "lucide-react"
import { toast } from "sonner"

export function EntryEditPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data, isLoading, error } = useEntry(id!)
  const updateEntry = useUpdateEntry()

  if (isLoading) {
    return (
      <div className="mx-auto max-w-3xl space-y-4">
        <Skeleton className="h-8 w-48" />
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

  const handleSubmit = (formData: UpdateEntryRequest) => {
    updateEntry.mutate(
      { id: id!, data: formData },
      {
        onSuccess: () => {
          toast.success("Entry updated")
          navigate(`/entries/${id}`)
        },
        onError: () => {
          toast.error("Failed to update entry")
        },
      }
    )
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <h1 className="text-2xl font-bold">Edit Entry</h1>
      <EntryForm
        entry={data.entry}
        onSubmit={handleSubmit as (data: unknown) => void}
        isSubmitting={updateEntry.isPending}
        submitLabel="Update"
      />
    </div>
  )
}
