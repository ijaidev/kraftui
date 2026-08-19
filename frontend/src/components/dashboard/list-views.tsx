"use client"

import { MachineSection } from "@/components/dashboard/machine-table"
import { PackageSection } from "@/components/dashboard/package-list"
import { ListPage } from "@/components/dashboard/list-page"
import {
  ResourceSection,
  mapNetworks,
  mapVolumes,
} from "@/components/dashboard/resource-table"
import {
  fetchMachines,
  fetchNetworks,
  fetchPackages,
  fetchVolumes,
} from "@/lib/runtime"

export function MachinesView() {
  return (
    <ListPage load={fetchMachines}>
      {(machines) => <MachineSection machines={machines} />}
    </ListPage>
  )
}

export function NetworksView() {
  return (
    <ListPage load={fetchNetworks}>
      {(networks) => (
        <ResourceSection
          title="Networks"
          empty="No networks."
          errorTitle="Could not load networks."
          heading="h1"
          items={mapNetworks(networks)}
        />
      )}
    </ListPage>
  )
}

export function VolumesView() {
  return (
    <ListPage load={fetchVolumes}>
      {(volumes) => (
        <ResourceSection
          title="Volumes"
          empty="No volumes."
          errorTitle="Could not load volumes."
          heading="h1"
          items={mapVolumes(volumes)}
        />
      )}
    </ListPage>
  )
}

export function PackagesView() {
  return (
    <ListPage load={fetchPackages}>
      {(packages) => <PackageSection packages={packages} heading="h1" />}
    </ListPage>
  )
}
