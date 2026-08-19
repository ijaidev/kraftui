import type { Metadata } from "next"

import { VolumesView } from "@/components/dashboard/list-views"

export const metadata: Metadata = {
  title: "Volumes · KraftUI",
}

export default function VolumesPage() {
  return <VolumesView />
}
