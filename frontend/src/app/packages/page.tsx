import type { Metadata } from "next"

import { PackagesView } from "@/components/dashboard/list-views"

export const metadata: Metadata = {
  title: "Packages · KraftUI",
}

export default function PackagesPage() {
  return <PackagesView />
}
