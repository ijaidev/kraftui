import type { ReactNode } from "react"
import Link from "next/link"

import { ModeToggle } from "@/components/mode-toggle"
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
} from "@/components/ui/navigation-menu"

export function SiteNav({ children }: { children?: ReactNode }) {
  return (
    <header className="border-b border-border">
      <div className="mx-auto flex w-full max-w-[1080px] items-center justify-between gap-4 px-6 py-3">
        <div className="flex min-w-0 items-center gap-4">
          <NavigationMenu>
            <NavigationMenuList>
              <NavigationMenuItem>
                <NavigationMenuLink
                  render={<Link href="/" />}
                  className="font-heading text-sm font-medium tracking-tight"
                >
                  <span translate="no">kraftui</span>
                </NavigationMenuLink>
              </NavigationMenuItem>
            </NavigationMenuList>
          </NavigationMenu>
          {children}
        </div>
        <ModeToggle />
      </div>
    </header>
  )
}
