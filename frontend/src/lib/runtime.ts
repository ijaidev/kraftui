import { getHealth, listMachines, listNetworks, listPackages, listVolumes } from "@/lib/api/sdk.gen"
import type { Error, Health, Machine, Network, Package, Volume } from "@/lib/api/types.gen"

export type ResourceResult<T> =
  | { ok: true; data: T }
  | { ok: false; message: string }

export type RuntimeSnapshot = {
  health: ResourceResult<Health>
  machines: ResourceResult<Machine[]>
  networks: ResourceResult<Network[]>
  volumes: ResourceResult<Volume[]>
  packages: ResourceResult<Package[]>
}

type SdkResult<T> = {
  data?: T
  error?: Error
  response?: Response
}

async function read<T>(request: Promise<SdkResult<T>>): Promise<ResourceResult<T>> {
  try {
    const result = await request
    if (result.data !== undefined) {
      return { ok: true, data: result.data }
    }
    if (result.error?.message) {
      return { ok: false, message: result.error.message }
    }
    if (!result.response) {
      return { ok: false, message: "Could not reach KraftUI." }
    }
    return { ok: false, message: "Kraft command failed." }
  } catch {
    return { ok: false, message: "Could not reach KraftUI." }
  }
}

const UNREACHABLE_MESSAGE = "Could not reach KraftUI."

export async function fetchMachines(): Promise<ResourceResult<Machine[]>> {
  return read(listMachines({ query: { all: true } }))
}

export async function fetchNetworks(): Promise<ResourceResult<Network[]>> {
  return read(listNetworks())
}

export async function fetchVolumes(): Promise<ResourceResult<Volume[]>> {
  return read(listVolumes())
}

export async function fetchPackages(): Promise<ResourceResult<Package[]>> {
  return read(listPackages({ query: { limit: 100 } }))
}

export async function fetchRuntime(): Promise<RuntimeSnapshot> {
  const [health, machines, networks, volumes, packages] = await Promise.all([
    read(getHealth()),
    fetchMachines(),
    fetchNetworks(),
    fetchVolumes(),
    fetchPackages(),
  ])

  return { health, machines, networks, volumes, packages }
}

export function isKraftUnreachable(
  result: ResourceResult<unknown>
): result is { ok: false; message: string } {
  return !result.ok && result.message === UNREACHABLE_MESSAGE
}

export function isRuntimeUnreachable(snapshot: RuntimeSnapshot): boolean {
  return (
    !snapshot.health.ok &&
    !snapshot.machines.ok &&
    !snapshot.networks.ok &&
    !snapshot.volumes.ok &&
    !snapshot.packages.ok
  )
}

export function kraftUnavailableCopy(message?: string): string {
  if (message === UNREACHABLE_MESSAGE) {
    return "Could not reach KraftUI. Start the backend with just dev."
  }
  return "Check that kraft 0.12.14 is on PATH, then restart KraftUI."
}
