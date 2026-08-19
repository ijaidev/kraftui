export type StatusTone = "live" | "destructive" | "secondary"

const LIVE_STATUSES = new Set(["running", "up", "bound", "restarting"])
const DESTRUCTIVE_STATUSES = new Set(["failed", "errored"])

export function statusTone(status: string): StatusTone {
  const normalized = status.trim().toLowerCase()
  if (LIVE_STATUSES.has(normalized)) {
    return "live"
  }
  if (DESTRUCTIVE_STATUSES.has(normalized)) {
    return "destructive"
  }
  return "secondary"
}

export function isLiveStatus(status: string): boolean {
  return statusTone(status) === "live"
}

export function formatPlatArch(platform?: string, architecture?: string): string {
  if (platform && architecture) {
    return `${platform}/${architecture}`
  }
  return platform || architecture || "—"
}

export function machineAddress(machine: { ip?: string; ports?: string }): string {
  const ports = machine.ports?.trim()
  if (ports) {
    return ports
  }
  const ip = machine.ip?.trim()
  if (ip) {
    return ip
  }
  return "—"
}

export function summarizeStatuses(items: { status: string }[]): string {
  if (items.length === 0) {
    return ""
  }

  const counts = new Map<string, number>()
  for (const item of items) {
    const key = item.status.trim().toLowerCase() || "unknown"
    counts.set(key, (counts.get(key) ?? 0) + 1)
  }

  const preferred = ["running", "up", "bound", "restarting", "paused", "suspended", "created", "exited", "down", "failed", "errored", "unknown"]
  const seen = new Set<string>()
  const parts: string[] = []

  for (const key of preferred) {
    const count = counts.get(key)
    if (count) {
      parts.push(`${count} ${key}`)
      seen.add(key)
    }
  }

  for (const [key, count] of counts) {
    if (!seen.has(key)) {
      parts.push(`${count} ${key}`)
    }
  }

  return parts.join(" · ")
}

export function summarizePackages(packages: { type: string }[]): string {
  const counts = new Map<string, number>()
  for (const item of packages) {
    const key = item.type.trim().toLowerCase() || "other"
    counts.set(key, (counts.get(key) ?? 0) + 1)
  }

  const parts = [`${packages.length} local`]
  const labels: Record<string, [string, string]> = {
    app: ["app", "apps"],
    lib: ["lib", "libs"],
    core: ["core", "core"],
  }
  for (const kind of ["app", "lib", "core"]) {
    const count = counts.get(kind)
    if (count) {
      const [singular, plural] = labels[kind]
      parts.push(`${count} ${count === 1 ? singular : plural}`)
    }
  }

  return parts.join(" · ")
}
