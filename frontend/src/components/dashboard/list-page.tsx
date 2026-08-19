"use client"

import { useEffect, useState, type ReactNode } from "react"

import {
  KraftUnavailableAlert,
  PageFrame,
} from "@/components/dashboard/page-frame"
import {
  isKraftUnreachable,
  type ResourceResult,
} from "@/lib/runtime"

export function ListPage<T>({
  load,
  children,
}: {
  load: () => Promise<ResourceResult<T>>
  children: (result?: ResourceResult<T>) => ReactNode
}) {
  const [result, setResult] = useState<ResourceResult<T>>()

  useEffect(() => {
    let cancelled = false

    load().then((next) => {
      if (!cancelled) {
        setResult(next)
      }
    })

    return () => {
      cancelled = true
    }
  }, [load])

  const unreachable = result !== undefined && isKraftUnreachable(result)

  return (
    <PageFrame busy={result === undefined}>
      {unreachable ? (
        <KraftUnavailableAlert message={result.message} />
      ) : (
        children(result)
      )}
    </PageFrame>
  )
}
