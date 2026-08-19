import { IconServer } from "@tabler/icons-react"

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

import { MonoText } from "@/components/dashboard/mono-text"
import { StatusBadge } from "@/components/dashboard/status-badge"
import { TableSkeleton } from "@/components/dashboard/table-skeleton"
import type { Machine } from "@/lib/api/types.gen"
import type { ResourceResult } from "@/lib/runtime"
import { formatPlatArch, machineAddress, summarizeStatuses } from "@/lib/status"

export function MachineSection({
  machines,
}: {
  machines?: ResourceResult<Machine[]>
}) {
  const meta = machines?.ok ? summarizeStatuses(machines.data) : undefined

  return (
    <section className="flex flex-col gap-4" aria-labelledby="machines-heading">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <h1
          id="machines-heading"
          className="font-heading text-sm font-medium tracking-tight text-pretty"
        >
          Machines
        </h1>
        {meta ? (
          <p className="font-mono text-xs text-muted-foreground tabular-nums">
            {meta}
          </p>
        ) : null}
      </div>
      <MachineBody machines={machines} />
    </section>
  )
}

function MachineBody({ machines }: { machines?: ResourceResult<Machine[]> }) {
  if (machines === undefined) {
    return <TableSkeleton columns={6} />
  }

  if (!machines.ok) {
    return (
      <Alert variant="destructive">
        <AlertTitle>Could not load machines.</AlertTitle>
        <AlertDescription>{machines.message}</AlertDescription>
      </Alert>
    )
  }

  if (machines.data.length === 0) {
    return (
      <Empty className="border">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <IconServer />
          </EmptyMedia>
          <EmptyTitle>No machines on this host.</EmptyTitle>
          <EmptyDescription>
            Start one with{" "}
            <span className="font-mono" translate="no">
              kraft run
            </span>
            .
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
          <TableHead>Name</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Platform</TableHead>
          <TableHead className="hidden md:table-cell">Memory</TableHead>
          <TableHead>Address</TableHead>
          <TableHead className="hidden lg:table-cell">Kernel</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {machines.data.map((machine) => (
          <TableRow
            key={machine.id ?? `${machine.name}-${machine.platform}-${machine.architecture}`}
          >
            <TableCell className="max-w-40 min-w-0 font-medium">
              <span className="block truncate" translate="no">
                {machine.name}
              </span>
            </TableCell>
            <TableCell>
              <StatusBadge status={machine.status} />
            </TableCell>
            <TableCell className="max-w-36 min-w-0">
              <MonoText
                value={formatPlatArch(machine.platform, machine.architecture)}
              />
            </TableCell>
            <TableCell className="hidden max-w-24 min-w-0 md:table-cell">
              <MonoText value={machine.memory} />
            </TableCell>
            <TableCell className="max-w-40 min-w-0">
              <MonoText value={machineAddress(machine)} />
            </TableCell>
            <TableCell className="hidden max-w-48 min-w-0 text-muted-foreground lg:table-cell">
              <MonoText value={machine.kernel} />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
