import { useNavigate } from "react-router-dom"
import { useCreateEntry } from "@/hooks/use-entries"
import { EntryForm } from "@/components/entries/entry-form"
import type { CreateEntryRequest } from "@/lib/types"
import { toast } from "sonner"

export function NewEntryPage() {
  const navigate = useNavigate()
  const createEntry = useCreateEntry()

  const handleSubmit = (data: CreateEntryRequest) => {
    createEntry.mutate(data, {
      onSuccess: (res) => {
        toast.success("Entry created")
        navigate(`/entries/${res.entry.id}`)
      },
      onError: () => {
        toast.error("Failed to create entry")
      },
    })
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <h1 className="text-2xl font-bold">New Entry</h1>
      <EntryForm
        onSubmit={handleSubmit as (data: unknown) => void}
        isSubmitting={createEntry.isPending}
        submitLabel="Create"
      />
    </div>
  )
}
