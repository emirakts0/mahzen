import { useEffect, useState } from "react"
import { Link } from "react-router-dom"
import { useKeywordSearch, useSemanticSearch } from "@/hooks/use-search"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { Separator } from "@/components/ui/separator"
import { Search as SearchIcon, Zap, Brain } from "lucide-react"

function useDebounce<T>(value: T, delay: number): T {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(id)
  }, [value, delay])
  return debounced
}

export function SearchPage() {
  const [inputValue, setInputValue] = useState("")
  const query = useDebounce(inputValue.trim(), 300)

  const keyword = useKeywordSearch({ query, limit: 20 })
  const semantic = useSemanticSearch({ query, limit: 10 })

  // Normalize raw Typesense TextMatch scores (large integers) to 0-1 using
  // log-scale relative to the top result in the set.
  const normalizedKeywordResults = (() => {
    const results = keyword.data?.results
    if (!results || results.length === 0) return results

    const scores = results.map((r) => r.score).filter((s) => s > 0)
    if (scores.length === 0) return results

    const maxLog = Math.log1p(Math.max(...scores))
    if (maxLog === 0) return results

    return results.map((r) => ({
      ...r,
      score: r.score > 0 ? Math.log1p(r.score) / maxLog : 0,
    }))
  })()

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <h1 className="text-2xl font-bold">Search</h1>

      <div className="relative">
        <SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search entries..."
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          className="pl-9"
          autoFocus
        />
      </div>

      {query && (
        <>
          {/* Keyword results */}
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Zap className="h-4 w-4 text-yellow-500" />
              <h2 className="text-lg font-semibold">Keyword</h2>
              {keyword.isLoading && (
                <Badge variant="outline" className="text-xs">Searching...</Badge>
              )}
              {keyword.data && (
                <Badge variant="secondary" className="text-xs">
                  {keyword.data.total} found
                </Badge>
              )}
            </div>

            {keyword.isLoading ? (
              <div className="space-y-2">
                {Array.from({ length: 3 }).map((_, i) => (
                  <Skeleton key={i} className="h-20 w-full rounded-xl" />
                ))}
              </div>
            ) : keyword.data?.results.length === 0 ? (
              <p className="text-sm text-muted-foreground">No matches found.</p>
            ) : (
              <div className="space-y-2">
                {normalizedKeywordResults?.map((result) => (
                  <SearchResultCard key={result.entry_id} result={result} />
                ))}
              </div>
            )}
          </div>

          <Separator />

          {/* Semantic results */}
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Brain className="h-4 w-4 text-purple-500" />
              <h2 className="text-lg font-semibold">Semantic</h2>
              {semantic.isLoading && (
                <Badge variant="outline" className="text-xs">Computing...</Badge>
              )}
              {semantic.isError && (
                <Badge variant="outline" className="text-xs text-muted-foreground">
                  Unavailable
                </Badge>
              )}
              {semantic.data && (
                <Badge variant="secondary" className="text-xs">
                  {semantic.data.total} found
                </Badge>
              )}
            </div>

            {semantic.isLoading ? (
              <div className="space-y-2">
                {Array.from({ length: 3 }).map((_, i) => (
                  <Skeleton key={i} className="h-20 w-full rounded-xl" />
                ))}
              </div>
            ) : semantic.isError ? (
              <p className="text-sm text-muted-foreground">
                Semantic search requires AI to be configured.
              </p>
            ) : semantic.data?.results.length === 0 ? (
              <p className="text-sm text-muted-foreground">No matches found.</p>
            ) : (
              <div className="space-y-2">
                {semantic.data?.results.map((result) => (
                  <SearchResultCard key={result.entry_id} result={result} />
                ))}
              </div>
            )}
          </div>
        </>
      )}
    </div>
  )
}

function SearchResultCard({
  result,
}: {
  result: { entry_id: string; title: string; snippet: string; score: number; highlights: string[] }
}) {
  const scoreLabel = (() => {
    if (!result.score || result.score <= 0) return null
    return `${Math.round(result.score * 100)}% match`
  })()

  return (
    <Link to={`/entries/${result.entry_id}`}>
      <Card className="transition-colors hover:bg-accent/50">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm">{result.title}</CardTitle>
            {scoreLabel && (
              <Badge variant="secondary" className="text-xs shrink-0 ml-auto">
                {scoreLabel}
              </Badge>
            )}
          </div>
        </CardHeader>
        <CardContent className="pt-0">
          {result.snippet && (
            <p className="text-sm text-muted-foreground line-clamp-2">
              {result.snippet}
            </p>
          )}
        </CardContent>
      </Card>
    </Link>
  )
}
