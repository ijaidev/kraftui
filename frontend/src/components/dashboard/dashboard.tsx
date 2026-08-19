"use client"

import { useEffect, useState } from "react"
import { IconAlertTriangle } from "@tabler/icons-react"

import { MachineSection } from "@/components/dashboard/machine-table"
import { PackageSection } from "@/components/dashboard/package-list"
import {
  ResourceSection,
  type ResourceRow,
} from "@/components/dashboard/resource-table"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Separator } from "@/components/ui/separator"
import type { Network, Volume } from "@/lib/api/types.gen"
import {
  fetchRuntime,
  isRuntimeUnreachable,
  type ResourceResult,
  type RuntimeSnapshot,
} from "@/lib/runtime"

export function Dashboard() {
  const [snapshot, setSnapshot] = useState<RuntimeSnapshot | null>(null)

  useEffect(() => {
    let cancelled = false

    fetchRuntime().then((next) => {
      if (!cancelled) {
        setSnapshot(next)
      }
    })

    return () => {
      cancelled = true
    }
  }, [])

  const unreachable = snapshot !== null && isRuntimeUnreachable(snapshot)

  return (
    <div className="flex min-h-full flex-1 flex-col">
      <main
        id="main"
        className="mx-auto flex w-full max-w-[1080px] flex-1 scroll-mt-6 flex-col gap-10 px-6 py-10"
        aria-busy={snapshot === null}
      >
        {unreachable ? (
          <Alert variant="destructive">
            <IconAlertTriangle />
            <AlertTitle>Kraft is unavailable.</AlertTitle>
            <AlertDescription>
              {kraftUnavailableCopy(snapshot.health)}
            </AlertDescription>
          </Alert>
        ) : (
          <>
            <MachineSection machines={snapshot?.machines} />
            <Separator />
            <div className="grid gap-10 md:grid-cols-2">
              <ResourceSection
                title="Networks"
                empty="No networks."
                errorTitle="Could not load networks."
                items={mapNetworks(snapshot?.networks)}
              />
              <ResourceSection
                title="Volumes"
                empty="No volumes."
                errorTitle="Could not load volumes."
                items={mapVolumes(snapshot?.volumes)}
              />
            </div>
            <Separator />
            <PackageSection packages={snapshot?.packages} />
          </>
        )}
      </main>
    </div>
  )
}

function kraftUnavailableCopy(health: RuntimeSnapshot["health"]): string {
  if (!health.ok && health.message === "Could not reach KraftUI.") {
    return "Could not reach KraftUI. Start the backend with just dev."
  }
  return "Check that kraft 0.12.14 is on PATH, then restart KraftUI."
}

function mapNetworks(
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

function mapVolumes(
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
