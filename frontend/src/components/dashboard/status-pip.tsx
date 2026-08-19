import { cn } from "@/lib/utils"

export function StatusPip({ live }: { live: boolean }) {
  return (
    <span
      aria-hidden="true"
      className={cn(
        "inline-block size-1.5 shrink-0 rounded-full",
        live ? "bg-live motion-safe:animate-pulse" : "bg-muted-foreground/40"
      )}
    />
  )
}
