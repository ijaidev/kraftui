import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { cn } from "@/lib/utils"

const TOOLTIP_MIN_LENGTH = 24

export function MonoText({
  value,
  className,
}: {
  value?: string
  className?: string
}) {
  const text = value?.trim() || "—"
  const showTooltip = Boolean(value && value.length >= TOOLTIP_MIN_LENGTH)

  const content = (
    <span
      className={cn("block truncate font-mono text-xs tabular-nums", className)}
      translate="no"
    >
      {text}
    </span>
  )

  if (!showTooltip) {
    return content
  }

  return (
    <Tooltip>
      <TooltipTrigger
        className="block max-w-full min-w-0 text-left"
        render={<span />}
      >
        {content}
      </TooltipTrigger>
      <TooltipContent>{value}</TooltipContent>
    </Tooltip>
  )
}
