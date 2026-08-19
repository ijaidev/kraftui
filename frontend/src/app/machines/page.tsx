import type { Metadata } from "next"

import { MachinesView } from "@/components/dashboard/list-views"

export const metadata: Metadata = {
  title: "Machines · KraftUI",
}

export default function MachinesPage() {
  return <MachinesView />
}
