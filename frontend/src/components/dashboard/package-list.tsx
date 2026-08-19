import { Badge } from "@/components/ui/badge"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

import { MonoText } from "@/components/dashboard/mono-text"
import { TableSkeleton } from "@/components/dashboard/table-skeleton"
import { ViewAllLink } from "@/components/dashboard/view-all-link"
import type { Package } from "@/lib/api/types.gen"
import type { ResourceResult } from "@/lib/runtime"
import { formatPlatArch, summarizePackages } from "@/lib/status"

export function PackageSection({
  packages,
  heading = "h2",
  limit,
  viewAllHref,
}: {
  packages?: ResourceResult<Package[]>
  heading?: "h1" | "h2"
  limit?: number
  viewAllHref?: string
}) {
  const meta = packages?.ok ? summarizePackages(packages.data) : undefined
  const Heading = heading

  return (
    <section className="flex flex-col gap-4" aria-labelledby="packages-heading">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <Heading
          id="packages-heading"
          className="font-heading text-sm font-medium tracking-tight text-pretty"
        >
          Packages
        </Heading>
        {meta ? (
          <p className="font-mono text-xs text-muted-foreground tabular-nums">
            {meta}
          </p>
        ) : null}
      </div>
      <PackageBody packages={packages} limit={limit} />
      {viewAllHref ? <ViewAllLink href={viewAllHref} /> : null}
    </section>
  )
}

function PackageBody({
  packages,
  limit,
}: {
  packages?: ResourceResult<Package[]>
  limit?: number
}) {
  if (packages === undefined) {
    return <TableSkeleton columns={4} />
  }

  if (!packages.ok) {
    return (
      <Alert variant="destructive">
        <AlertTitle>Could not load packages.</AlertTitle>
        <AlertDescription>{packages.message}</AlertDescription>
      </Alert>
    )
  }

  if (packages.data.length === 0) {
    return <p className="text-sm text-muted-foreground">No local packages.</p>
  }

  const rows =
    limit === undefined ? packages.data : packages.data.slice(0, limit)

  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
          <TableHead>Name</TableHead>
          <TableHead>Version</TableHead>
          <TableHead>Type</TableHead>
          <TableHead>Platform</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {rows.map((item) => (
          <TableRow
            key={`${item.name}-${item.version}-${item.type}-${item.platform ?? ""}-${item.architecture ?? ""}`}
          >
            <TableCell className="max-w-56 min-w-0 font-medium">
              <MonoText value={item.name} className="font-medium" />
            </TableCell>
            <TableCell className="max-w-28 min-w-0">
              <MonoText value={item.version} />
            </TableCell>
            <TableCell>
              <Badge variant="outline" translate="no">
                {item.type}
              </Badge>
            </TableCell>
            <TableCell className="max-w-36 min-w-0">
              <MonoText
                value={formatPlatArch(item.platform, item.architecture)}
              />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
