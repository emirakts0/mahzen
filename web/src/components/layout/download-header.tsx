import { useState, useEffect } from "react"
import { AnimatePresence, motion } from "framer-motion"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { SiLinux, SiApple } from "@icons-pack/react-simple-icons"

function WindowsIcon({ className, style }: { className?: string; style?: React.CSSProperties }) {
  return (
    <svg className={className} style={style} viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" fill="currentColor">
      <path d="M0 0h11.4v11.4H0zM12.6 0H24v11.4H12.6zM0 12.6h11.4V24H0zM12.6 12.6H24V24H12.6z" />
    </svg>
  )
}

const OS_ICONS = [WindowsIcon, SiLinux, SiApple]

const PLATFORMS = [
  { id: "windows", label: "Windows", ext: ".exe", Icon: WindowsIcon, disabled: false },
  { id: "linux", label: "Linux", ext: ".AppImage", Icon: SiLinux, disabled: false },
  { id: "macos", label: "macOS", ext: ".dmg", Icon: SiApple, disabled: true },
] as const

export function DownloadHeader() {
  const [activeIcon, setActiveIcon] = useState(0)

  useEffect(() => {
    const timer = setInterval(() => {
      setActiveIcon((prev) => (prev + 1) % OS_ICONS.length)
    }, 2000)
    return () => clearInterval(timer)
  }, [])

  const handleDownload = async (platform: string) => {
    const res = await fetch(`/v1/downloads/${platform}`)
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: "Download failed" }))
      alert(err.error)
      return
    }
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    const disposition = res.headers.get("Content-Disposition")
    const match = disposition?.match(/filename="(.+)"/)
    a.download = match?.[1] ?? `mahzen-${platform}`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="pointer-events-auto hidden md:block">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <button
            className="flex items-center gap-2 rounded-2xl px-3.5 py-2 shadow-lg backdrop-blur-md transition-colors cursor-pointer"
            style={{
              background: "var(--glass-bg)",
              border: "1px solid var(--glass-border)",
            }}
            onMouseEnter={(e) => {
              ;(e.currentTarget as HTMLElement).style.background = "var(--glass-hover)"
            }}
            onMouseLeave={(e) => {
              ;(e.currentTarget as HTMLElement).style.background = "var(--glass-bg)"
            }}
          >
            <div className="flex h-7 items-center gap-2">
              <div className="relative flex h-4 w-4 items-center justify-center overflow-hidden">
              <AnimatePresence mode="wait">
                <motion.div
                  key={activeIcon}
                  initial={{ y: 7, opacity: 0 }}
                  animate={{ y: 0, opacity: 1 }}
                  exit={{ y: -7, opacity: 0 }}
                  transition={{ duration: 0.25 }}
                  className="flex items-center justify-center"
                >
                  {(() => {
                    const Icon = OS_ICONS[activeIcon]
                    return <Icon className="h-4 w-4" style={{ color: "var(--glass-icon)" }} />
                  })()}
                </motion.div>
              </AnimatePresence>
            </div>
              <span className="text-xs font-medium" style={{ color: "var(--glass-text)" }}>
                Download
              </span>
            </div>
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent
          align="end"
          className="w-52 rounded-xl border-0 shadow-xl backdrop-blur-xl"
          style={{
            background: "var(--glass-bg)",
            border: "1px solid var(--glass-border)",
          }}
        >
          {PLATFORMS.map(({ id, label, ext, Icon, disabled }) => (
            <DropdownMenuItem
              key={id}
              disabled={disabled}
              className="flex cursor-pointer items-center gap-2.5 px-2.5 py-2"
              style={{ color: "var(--glass-text)" }}
              onSelect={(e) => {
                e.preventDefault()
                if (!disabled) void handleDownload(id)
              }}
            >
              <Icon className="h-4 w-4" style={{ color: "var(--glass-icon)" }} />
              <span className="flex-1 text-xs font-medium">{label}</span>
              {disabled ? (
                <span
                  className="rounded-md px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wider"
                  style={{
                    background: "var(--glass-hover)",
                    color: "var(--glass-text-muted)",
                  }}
                >
                  Soon
                </span>
              ) : (
                <span className="text-[10px] font-mono" style={{ color: "var(--glass-text-muted)" }}>
                  {ext}
                </span>
              )}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
