import { Skeleton } from "@/components/ui/skeleton"

export function TableSkeleton({ columns }: { columns: number }) {
  return (
    <div className="flex flex-col gap-2" aria-hidden="true">
      {Array.from({ length: 4 }, (_, row) => (
        <div className="flex gap-4" key={row}>
          {Array.from({ length: columns }, (_, column) => (
            <Skeleton className="h-4 flex-1" key={column} />
          ))}
        </div>
      ))}
    </div>
  )
}
