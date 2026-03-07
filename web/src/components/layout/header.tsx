import { Link, useLocation, useNavigate, useSearchParams } from "react-router"
import { useAuth } from "@/hooks/use-auth"
import { useTheme } from "@/hooks/use-theme"
import { useState, useEffect } from "react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { BookOpen, Tag, Search, LogOut, Moon, Sun, LogIn, UserPlus } from "lucide-react"

const NAV_ITEMS = [
  { to: "/search", label: "Search", icon: Search },
  { to: "/entries", label: "Entries", icon: BookOpen },
  { to: "/tags", label: "Tags", icon: Tag },
] as const

export function Header() {
  const location = useLocation()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const { user, isAuthenticated, logout } = useAuth()
  const { theme, toggleTheme } = useTheme()
  const [isHidden, setIsHidden] = useState(false)

  const authType = searchParams.get("auth")
  const isLoginOpen = authType === "login"
  const isSignupOpen = authType === "signup"

  useEffect(() => {
    const handleEvent = (e: Event) => {
      const customEvent = e as CustomEvent<boolean>
      setIsHidden(customEvent.detail)
    }
    window.addEventListener("search-sticky", handleEvent)
    return () => window.removeEventListener("search-sticky", handleEvent)
  }, [])

  const handleLogout = async () => {
    await logout()
    void navigate("/login")
  }

  return (
    /* Floating island wrapper — centered, no full-width bar */
    <div 
      className="fixed top-4 inset-x-0 z-[60] flex justify-center px-4 pointer-events-none transition-transform duration-300 ease-out"
      style={{
        transform: isHidden ? "translateY(-150%)" : "translateY(0)",
      }}
    >
      <header
        className="pointer-events-auto flex items-center gap-2 rounded-2xl px-3 py-2 shadow-lg backdrop-blur-md"
        style={{
          background: "var(--glass-bg)",
          border: "1px solid var(--glass-border)",
        }}
      >
        {isAuthenticated ? (
          <>
            {/* Nav items */}
            {NAV_ITEMS.map(({ to, label, icon: Icon }) => {
              const isActive =
                location.pathname === to || location.pathname.startsWith(to + "/")
              return (
                <Link
                  key={to}
                  to={to}
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
                </Link>
              )
            })}

            <div className="mx-1 h-4 w-px" style={{ background: "var(--glass-divider)" }} />

            {/* Dark mode toggle */}
            <button
              onClick={toggleTheme}
              aria-label="Toggle theme"
              className="flex h-7 w-7 items-center justify-center rounded-lg transition-colors"
              style={{ color: "var(--glass-icon)", background: "transparent" }}
              onMouseEnter={e => { (e.currentTarget as HTMLElement).style.background = "var(--glass-bg)" }}
              onMouseLeave={e => { (e.currentTarget as HTMLElement).style.background = "transparent" }}
            >
              {theme === "dark" ? <Sun className="h-3.5 w-3.5" /> : <Moon className="h-3.5 w-3.5" />}
            </button>

            {/* User avatar dropdown */}
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
          <>
            {/* Unauthenticated: dark toggle + login/signup */}
            <button
              onClick={toggleTheme}
              aria-label="Toggle theme"
              className="flex h-7 w-7 items-center justify-center rounded-lg transition-colors"
              style={{ color: "var(--glass-icon)", background: "transparent" }}
              onMouseEnter={e => { (e.currentTarget as HTMLElement).style.background = "var(--glass-bg)" }}
              onMouseLeave={e => { (e.currentTarget as HTMLElement).style.background = "transparent" }}
            >
              {theme === "dark" ? <Sun className="h-3.5 w-3.5" /> : <Moon className="h-3.5 w-3.5" />}
            </button>

            <div className="mx-1 h-4 w-px" style={{ background: "var(--glass-divider)" }} />

            <button
              onClick={() => {
                setSearchParams((prev) => { prev.set("auth", "login"); return prev }, { replace: true })
              }}
              className="flex h-7 items-center gap-1.5 rounded-xl px-3 text-xs font-medium transition-colors"
              style={{
                background: isLoginOpen ? "var(--glass-hover)" : "transparent",
                color: isLoginOpen ? "var(--glass-text)" : "var(--glass-text-muted)",
              }}
              onMouseEnter={e => { if (!isLoginOpen) (e.currentTarget as HTMLElement).style.background = "var(--glass-bg)" }}
              onMouseLeave={e => { if (!isLoginOpen) (e.currentTarget as HTMLElement).style.background = "transparent" }}
            >
              <LogIn className="h-3.5 w-3.5" />
              Login
            </button>
            <button
              onClick={() => {
                setSearchParams((prev) => { prev.set("auth", "signup"); return prev }, { replace: true })
              }}
              className="flex h-7 items-center gap-1.5 rounded-xl px-3 text-xs font-medium transition-colors"
              style={{
                background: isSignupOpen || (!isLoginOpen && !isSignupOpen) ? "var(--glass-hover)" : "transparent",
                border: isSignupOpen || (!isLoginOpen && !isSignupOpen) ? "1px solid var(--glass-border)" : "1px solid transparent",
                color: isSignupOpen || (!isLoginOpen && !isSignupOpen) ? "var(--glass-text)" : "var(--glass-text-muted)",
              }}
              onMouseEnter={e => { if (isLoginOpen) (e.currentTarget as HTMLElement).style.background = "var(--glass-bg)" }}
              onMouseLeave={e => { if (isLoginOpen) (e.currentTarget as HTMLElement).style.background = "transparent" }}
            >
              <UserPlus className="h-3.5 w-3.5" />
              Sign up
            </button>
          </>
        )}
      </header>
    </div>
  )
}
