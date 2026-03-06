import { Outlet } from "react-router"
import { Header } from "@/components/layout/header"
import { Toaster } from "@/components/ui/sonner"
import { TooltipProvider } from "@/components/ui/tooltip"
import GrainientBackground from "@/components/layout/grainient-background"
import { PageScrollbar } from "@/components/layout/page-scrollbar"

export default function RootLayout() {
  return (
    <TooltipProvider>
      {/* Fixed full-screen animated background */}
      <GrainientBackground />

      {/* Floating island header */}
      <Header />

      {/* Page content — sits above background */}
      <main className="relative z-10 min-h-screen pt-20">
        <Outlet />
      </main>

      {/* Overlay scrollbar — floats above Grainient, never touches the background */}
      <PageScrollbar />

      <Toaster richColors position="top-right" />
    </TooltipProvider>
  )
}
