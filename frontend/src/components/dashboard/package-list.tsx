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
import type { Package } from "@/lib/api/types.gen"
import type { ResourceResult } from "@/lib/runtime"
import { formatPlatArch, summarizePackages } from "@/lib/status"

const VISIBLE_PACKAGES = 8

export function PackageSection({
  packages,
}: {
  packages?: ResourceResult<Package[]>
}) {
  const meta = packages?.ok ? summarizePackages(packages.data) : undefined

  return (
    <section className="flex flex-col gap-4" aria-labelledby="packages-heading">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <h2
          id="packages-heading"
          className="font-heading text-sm font-medium tracking-tight text-pretty"
        >
          Packages
        </h2>
        {meta ? (
          <p className="font-mono text-xs text-muted-foreground tabular-nums">
            {meta}
          </p>
        ) : null}
      </div>
      <PackageBody packages={packages} />
    </section>
  )
}

function PackageBody({ packages }: { packages?: ResourceResult<Package[]> }) {
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

  const visible = packages.data.slice(0, VISIBLE_PACKAGES)
  const remaining = packages.data.length - visible.length

  return (
    <div className="flex flex-col gap-2">
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
          {visible.map((item) => (
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
      {remaining > 0 ? (
        <p className="text-xs text-muted-foreground tabular-nums">
          +{remaining} more locally
        </p>
      ) : null}
    </div>
  )
}
