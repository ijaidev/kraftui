import { Badge } from "@/components/ui/badge"

import { StatusPip } from "@/components/dashboard/status-pip"
import { isLiveStatus, statusTone } from "@/lib/status"

export function StatusBadge({ status }: { status: string }) {
  const tone = statusTone(status)

  return (
    <Badge variant={tone}>
      <StatusPip live={isLiveStatus(status)} />
      <span translate="no">{status}</span>
    </Badge>
  )
}
