"use client"

import { useEffect, useState } from "react"

import { StatusPip } from "@/components/dashboard/status-pip"
import { Separator } from "@/components/ui/separator"
import { getHealth } from "@/lib/api/sdk.gen"
import type { Health } from "@/lib/api/types.gen"
import type { ResourceResult } from "@/lib/runtime"

export function RuntimeStatus() {
  const [health, setHealth] = useState<ResourceResult<Health>>()

  useEffect(() => {
    let cancelled = false

    getHealth()
      .then((result) => {
        if (cancelled) {
          return
        }
        if (result.data) {
          setHealth({ ok: true, data: result.data })
          return
        }
        setHealth({
          ok: false,
          message: result.error?.message ?? "Could not reach KraftUI.",
        })
      })
      .catch(() => {
        if (!cancelled) {
          setHealth({ ok: false, message: "Could not reach KraftUI." })
        }
      })

    return () => {
      cancelled = true
    }
  }, [])

  const ready = health?.ok === true
  const label =
    health === undefined ? "Checking…" : ready ? "Ready" : "Unavailable"

  return (
    <div className="flex min-w-0 items-center gap-3 text-xs" aria-live="polite">
      <span className="inline-flex items-center gap-1.5">
        <StatusPip live={ready} />
        <span>{label}</span>
      </span>
      {ready ? (
        <>
          <Separator orientation="vertical" className="hidden h-3 sm:block" />
          <p
            className="hidden min-w-0 truncate font-mono text-muted-foreground tabular-nums sm:block"
            translate="no"
          >
            kraft {health.data.kraftVersion}
            <span className="text-border"> · </span>
            kraftui {health.data.version}
          </p>
        </>
      ) : null}
    </div>
  )
}
