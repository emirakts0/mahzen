import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { BookOpen, Search, LogOut, Settings, LogIn } from "lucide-react"
import { useAuth } from "@/providers/auth-provider"

const NAV_ITEMS = [
  { mode: "search" as const, label: "Search", icon: Search },
  { mode: "entries" as const, label: "Entries", icon: BookOpen },
]

interface DesktopHeaderProps {
  viewMode: "search" | "entries"
  onViewModeChange: (mode: "search" | "entries") => void
  onSettingsClick: () => void
  onConnectClick: () => void
}

export function DesktopHeader({
  viewMode,
  onViewModeChange,
  onSettingsClick,
  onConnectClick,
}: DesktopHeaderProps) {
  const { user, isAuthenticated, logout } = useAuth()

  const handleLogout = async () => {
    await logout()
  }

  return (
    <div 
      className="fixed top-4 inset-x-0 z-[60] flex justify-center px-4 pointer-events-none"
    >
      <header
        data-tauri-drag-region
        className="pointer-events-auto flex items-center gap-2 rounded-2xl px-3 py-2 shadow-lg backdrop-blur-md"
        style={{
          background: "var(--glass-bg)",
          border: "1px solid var(--glass-border)",
        }}
      >
        {isAuthenticated ? (
          <>
            {NAV_ITEMS.map(({ mode, label, icon: Icon }) => {
              const isActive = viewMode === mode
              return (
                <button
                  key={mode}
                  onClick={() => onViewModeChange(mode)}
                  className="flex items-center gap-1.5 rounded-xl px-3 py-1.5 text-xs font-medium transition-colors"
                  style={{
                    background: isActive ? "var(--glass-hover)" : "transparent",
                    color: isActive ? "var(--glass-text)" : "var(--glass-text-muted)",
                  }}
                  onMouseEnter={e => { if (!isActive) (e.currentTarget as HTMLElement).style.background = "var(--glass-bg)" }}
                  onMouseLeave={e => { if (!isActive) (e.currentTarget as HTMLElement).style.background = "transparent" }}
                >
                  <Icon className="h-3.5 w-3.5" />
                  {label}
                </button>
              )
            })}

            <div className="mx-1 h-4 w-px" style={{ background: "var(--glass-divider)" }} />

            <DropdownMenu>
               <DropdownMenuTrigger asChild>
                 <button
                   aria-label="User menu"
                   className="flex h-7 w-7 items-center justify-center rounded-lg text-xs font-semibold transition-colors"
                   style={{ background: "var(--glass-bg)", border: "1px solid var(--glass-border)", color: "var(--glass-text)" }}
                   onMouseEnter={e => { (e.currentTarget as HTMLElement).style.background = "var(--glass-hover)" }}
                   onMouseLeave={e => { (e.currentTarget as HTMLElement).style.background = "var(--glass-bg)" }}
                 >
                   {user?.display_name?.[0]?.toUpperCase() ??
                     user?.email?.[0]?.toUpperCase() ??
                     "U"}
                 </button>
               </DropdownMenuTrigger>
               <DropdownMenuContent 
                 align="end" 
                 className="w-48 rounded-xl border-0 shadow-xl backdrop-blur-xl"
                 style={{
                   background: "var(--glass-bg)",
                   border: "1px solid var(--glass-border)",
                 }}
               >
                 <div className="px-2 py-1.5">
                   <p className="text-xs font-medium" style={{ color: "var(--glass-text)" }}>{user?.display_name || "User"}</p>
                   <p className="truncate text-xs" style={{ color: "var(--glass-text-muted)" }}>{user?.email}</p>
                 </div>
                 <DropdownMenuSeparator style={{ background: "var(--glass-divider)" }} />
                 <DropdownMenuItem
                   className="flex cursor-pointer items-center gap-2"
                   style={{ color: "var(--glass-text)" }}
                   onClick={onSettingsClick}
                 >
                   <Settings className="h-3.5 w-3.5" />
                   Settings
                 </DropdownMenuItem>
                 <DropdownMenuItem
                   className="flex cursor-pointer items-center gap-2 focus:bg-destructive/10"
                   style={{ color: "var(--glass-error, #ef4444)" }}
                   onClick={() => void handleLogout()}
                 >
                   <LogOut className="h-3.5 w-3.5" />
                   Log out
                 </DropdownMenuItem>
               </DropdownMenuContent>
             </DropdownMenu>
          </>
        ) : (
          <button
            onClick={onConnectClick}
            className="flex h-7 items-center gap-1.5 rounded-xl px-3 text-xs font-medium transition-colors"
            style={{
              background: "var(--glass-hover)",
              border: "1px solid var(--glass-border)",
              color: "var(--glass-text)",
            }}
          >
            <LogIn className="h-3.5 w-3.5" />
            Connect
          </button>
        )}
      </header>
    </div>
  )
}
