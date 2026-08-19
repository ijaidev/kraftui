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
import { StatusBadge } from "@/components/dashboard/status-badge"
import type { ResourceResult } from "@/lib/runtime"

export type ResourceRow = {
  key: string
  name: string
  status: string
  driver: string
  detail?: string
}

export function ResourceSection({
  title,
  empty,
  errorTitle,
  items,
}: {
  title: string
  empty: string
  errorTitle: string
  items?: ResourceResult<ResourceRow[]>
}) {
  const headingId = `${title.toLowerCase()}-heading`

  return (
    <section className="flex min-w-0 flex-col gap-4" aria-labelledby={headingId}>
      <h2
        id={headingId}
        className="font-heading text-sm font-medium tracking-tight text-pretty"
      >
        {title}
      </h2>
      <ResourceBody
        empty={empty}
        errorTitle={errorTitle}
        items={items}
        title={title}
      />
    </section>
  )
}

function ResourceBody({
  title,
  empty,
  errorTitle,
  items,
}: {
  title: string
  empty: string
  errorTitle: string
  items?: ResourceResult<ResourceRow[]>
}) {
  if (items === undefined) {
    return <TableSkeleton columns={4} />
  }

  if (!items.ok) {
    return (
      <Alert variant="destructive">
        <AlertTitle>{errorTitle}</AlertTitle>
        <AlertDescription>{items.message}</AlertDescription>
      </Alert>
    )
  }

  if (items.data.length === 0) {
    return <p className="text-sm text-muted-foreground">{empty}</p>
  }

  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
          <TableHead>Name</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Driver</TableHead>
          <TableHead>{title === "Networks" ? "Network" : "Source"}</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {items.data.map((item) => (
          <TableRow key={item.key}>
            <TableCell className="max-w-36 min-w-0 font-medium">
              <span className="block truncate" translate="no">
                {item.name}
              </span>
            </TableCell>
            <TableCell>
              <StatusBadge status={item.status} />
            </TableCell>
            <TableCell className="max-w-28 min-w-0">
              <MonoText value={item.driver} />
            </TableCell>
            <TableCell className="max-w-40 min-w-0">
              <MonoText value={item.detail} />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
