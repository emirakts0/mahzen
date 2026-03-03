import { Link, useLocation } from "react-router-dom"
import {
  BookOpen,
  Tags,
  Search,
  Plus,
  Home,
} from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"

const navigation = [
  { name: "Home", href: "/", icon: Home },
  { name: "Entries", href: "/entries", icon: BookOpen },
  { name: "Tags", href: "/tags", icon: Tags },
  { name: "Search", href: "/search", icon: Search },
]

export function Sidebar() {
  const location = useLocation()

  return (
    <aside className="flex h-full w-64 flex-col border-r bg-sidebar text-sidebar-foreground">
      {/* Logo */}
      <div className="flex h-14 items-center px-4">
        <Link to="/" className="flex items-center gap-2 font-semibold text-lg">
          <BookOpen className="h-5 w-5" />
          Mahzen
        </Link>
      </div>

      <Separator />

      {/* New Entry */}
      <div className="p-3">
        <Button asChild className="w-full justify-start gap-2">
          <Link to="/entries/new">
            <Plus className="h-4 w-4" />
            New Entry
          </Link>
        </Button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 px-3">
        {navigation.map((item) => {
          const isActive =
            item.href === "/"
              ? location.pathname === "/"
              : location.pathname.startsWith(item.href)
          return (
            <Link
              key={item.name}
              to={item.href}
              className={cn(
                "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
                isActive
                  ? "bg-sidebar-accent text-sidebar-accent-foreground"
                  : "text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
              )}
            >
              <item.icon className="h-4 w-4" />
              {item.name}
            </Link>
          )
        })}
      </nav>
    </aside>
  )
}
