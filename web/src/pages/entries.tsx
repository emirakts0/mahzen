import { useState } from "react"
import { useSearchParams, Link } from "react-router-dom"
import { useEntries } from "@/hooks/use-entries"
import { EntryList } from "@/components/entries/entry-list"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Plus, Filter } from "lucide-react"
import { DEFAULT_PAGE_SIZE } from "@/lib/constants"

export function EntriesPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const pathFilter = searchParams.get("path") ?? undefined
  const page = parseInt(searchParams.get("page") ?? "1", 10)
  const offset = (page - 1) * DEFAULT_PAGE_SIZE

  const [pathInput, setPathInput] = useState(pathFilter ?? "")

  const { data, isLoading } = useEntries({
    path: pathFilter,
    limit: DEFAULT_PAGE_SIZE,
    offset,
  })

  const totalPages = data ? Math.ceil(data.total / DEFAULT_PAGE_SIZE) : 0

  const handlePathFilter = () => {
    const params = new URLSearchParams(searchParams)
    if (pathInput) {
      params.set("path", pathInput)
    } else {
      params.delete("path")
    }
    params.delete("page")
    setSearchParams(params)
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Entries</h1>
        <Button asChild>
          <Link to="/entries/new">
            <Plus className="mr-2 h-4 w-4" />
            New Entry
          </Link>
        </Button>
      </div>

      {/* Path filter */}
      <div className="flex gap-2">
        <Input
          placeholder="Filter by path, e.g. /notes"
          value={pathInput}
          onChange={(e) => setPathInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handlePathFilter()}
          className="max-w-sm"
        />
        <Button variant="outline" size="icon" onClick={handlePathFilter}>
          <Filter className="h-4 w-4" />
        </Button>
        {pathFilter && (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              setPathInput("")
              const params = new URLSearchParams(searchParams)
              params.delete("path")
              params.delete("page")
              setSearchParams(params)
            }}
          >
            Clear
          </Button>
        )}
      </div>

      {pathFilter && (
        <p className="text-sm text-muted-foreground">
          Showing entries under <code className="font-mono">{pathFilter}</code>
          {data && ` (${data.total} total)`}
        </p>
      )}

      <EntryList entries={data?.entries} isLoading={isLoading} />

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => {
              const params = new URLSearchParams(searchParams)
              params.set("page", String(page - 1))
              setSearchParams(params)
            }}
          >
            Previous
          </Button>
          <span className="text-sm text-muted-foreground">
            Page {page} of {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages}
            onClick={() => {
              const params = new URLSearchParams(searchParams)
              params.set("page", String(page + 1))
              setSearchParams(params)
            }}
          >
            Next
          </Button>
        </div>
      )}
    </div>
  )
}
