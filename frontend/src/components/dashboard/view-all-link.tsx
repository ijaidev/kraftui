import Link from "next/link"
import { IconChevronRight } from "@tabler/icons-react"

import { Button } from "@/components/ui/button"

export function ViewAllLink({ href }: { href: string }) {
  return (
    <div className="flex justify-end">
      <Button
        variant="link"
        size="xs"
        nativeButton={false}
        render={<Link href={href} />}
      >
        View all
        <IconChevronRight data-icon="inline-end" />
      </Button>
    </div>
  )
}
