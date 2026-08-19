import type { Metadata } from "next"

import { NetworksView } from "@/components/dashboard/list-views"

export const metadata: Metadata = {
  title: "Networks · KraftUI",
}

export default function NetworksPage() {
  return <NetworksView />
}
