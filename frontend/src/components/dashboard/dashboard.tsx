"use client"

import { useEffect, useState } from "react"

import { MachineSection } from "@/components/dashboard/machine-table"
import { PackageSection } from "@/components/dashboard/package-list"
import {
  KraftUnavailableAlert,
  PageFrame,
} from "@/components/dashboard/page-frame"
import {
  ResourceSection,
  mapNetworks,
  mapVolumes,
} from "@/components/dashboard/resource-table"
import { Separator } from "@/components/ui/separator"
import {
  fetchRuntime,
  isRuntimeUnreachable,
  type RuntimeSnapshot,
} from "@/lib/runtime"

const PREVIEW_LIMIT = 4

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
    <PageFrame busy={snapshot === null}>
      {unreachable ? (
        <KraftUnavailableAlert
          message={snapshot.health.ok ? undefined : snapshot.health.message}
        />
      ) : (
        <>
          <MachineSection
            machines={snapshot?.machines}
            limit={PREVIEW_LIMIT}
            viewAllHref="/machines"
          />
          <Separator />
          <div className="grid gap-10 md:grid-cols-2">
            <ResourceSection
              title="Networks"
              empty="No networks."
              errorTitle="Could not load networks."
              items={mapNetworks(snapshot?.networks)}
              limit={PREVIEW_LIMIT}
              viewAllHref="/networks"
            />
            <ResourceSection
              title="Volumes"
              empty="No volumes."
              errorTitle="Could not load volumes."
              items={mapVolumes(snapshot?.volumes)}
              limit={PREVIEW_LIMIT}
              viewAllHref="/volumes"
            />
          </div>
          <Separator />
          <PackageSection
            packages={snapshot?.packages}
            limit={PREVIEW_LIMIT}
            viewAllHref="/packages"
          />
        </>
      )}
    </PageFrame>
  )
}
