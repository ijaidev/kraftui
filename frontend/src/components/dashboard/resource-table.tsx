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
import { ViewAllLink } from "@/components/dashboard/view-all-link"
import type { Network, Volume } from "@/lib/api/types.gen"
import type { ResourceResult } from "@/lib/runtime"

export type ResourceRow = {
  key: string
  name: string
  status: string
  driver: string
  detail?: string
}

export function mapNetworks(
  networks?: ResourceResult<Network[]>
): ResourceResult<ResourceRow[]> | undefined {
  if (networks === undefined) {
    return undefined
  }
  if (!networks.ok) {
    return networks
  }
  return {
    ok: true,
    data: networks.data.map((network) => ({
      key: network.machineId ?? `${network.name}-${network.driver}`,
      name: network.name,
      status: network.status,
      driver: network.driver,
      detail: network.network,
    })),
  }
}

export function mapVolumes(
  volumes?: ResourceResult<Volume[]>
): ResourceResult<ResourceRow[]> | undefined {
  if (volumes === undefined) {
    return undefined
  }
  if (!volumes.ok) {
    return volumes
  }
  return {
    ok: true,
    data: volumes.data.map((volume) => ({
      key: volume.id ?? `${volume.name}-${volume.driver}`,
      name: volume.name,
      status: volume.status,
      driver: volume.driver,
      detail: volume.source,
    })),
  }
}

export function ResourceSection({
  title,
  empty,
  errorTitle,
  items,
  heading = "h2",
  limit,
  viewAllHref,
}: {
  title: string
  empty: string
  errorTitle: string
  items?: ResourceResult<ResourceRow[]>
  heading?: "h1" | "h2"
  limit?: number
  viewAllHref?: string
}) {
  const headingId = `${title.toLowerCase()}-heading`
  const Heading = heading

  return (
    <section className="flex min-w-0 flex-col gap-4" aria-labelledby={headingId}>
      <Heading
        id={headingId}
        className="font-heading text-sm font-medium tracking-tight text-pretty"
      >
        {title}
      </Heading>
      <ResourceBody
        empty={empty}
        errorTitle={errorTitle}
        items={items}
        limit={limit}
        title={title}
      />
      {viewAllHref ? <ViewAllLink href={viewAllHref} /> : null}
    </section>
  )
}

function ResourceBody({
  title,
  empty,
  errorTitle,
  items,
  limit,
}: {
  title: string
  empty: string
  errorTitle: string
  items?: ResourceResult<ResourceRow[]>
  limit?: number
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

  const rows = limit === undefined ? items.data : items.data.slice(0, limit)

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
        {rows.map((item) => (
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
