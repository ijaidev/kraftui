import type { ReactNode } from "react"
import { IconAlertTriangle } from "@tabler/icons-react"

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { kraftUnavailableCopy } from "@/lib/runtime"

export function PageFrame({
  children,
  busy,
}: {
  children: ReactNode
  busy?: boolean
}) {
  return (
    <div className="flex min-h-full flex-1 flex-col">
      <main
        id="main"
        className="mx-auto flex w-full max-w-[1080px] flex-1 scroll-mt-6 flex-col gap-10 px-6 py-10"
        aria-busy={busy}
      >
        {children}
      </main>
    </div>
  )
}

export function KraftUnavailableAlert({ message }: { message?: string }) {
  return (
    <Alert variant="destructive">
      <IconAlertTriangle />
      <AlertTitle>Kraft is unavailable.</AlertTitle>
      <AlertDescription>{kraftUnavailableCopy(message)}</AlertDescription>
    </Alert>
  )
}
