import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"

function ResultCardSkeleton({ className }: { className?: string }) {
  return (
    <div className={cn("rounded-xl border border-border bg-card/80 backdrop-blur-sm p-4 shadow-sm", className)}>
      <Skeleton className="h-4 w-3/4" />
      <Skeleton className="mt-1.5 h-3 w-1/3" />
      <div className="mt-3 space-y-2">
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-3 w-2/3" />
      </div>
      <div className="mt-3 flex gap-2">
        <Skeleton className="h-5 w-12 rounded-full" />
        <Skeleton className="h-5 w-16 rounded-full" />
      </div>
    </div>
  )
}

interface SearchSkeletonProps {
  count?: number
}

export function SearchSkeleton({ count = 3 }: SearchSkeletonProps) {
  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
      <div className="space-y-3">
        {Array.from({ length: count }).map((_, i) => (
          <ResultCardSkeleton key={i} />
        ))}
      </div>
      <div className="space-y-3">
        {Array.from({ length: count }).map((_, i) => (
          <ResultCardSkeleton key={i} />
        ))}
      </div>
    </div>
  )
}
