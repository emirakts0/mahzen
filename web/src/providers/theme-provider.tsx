import { useEffect, useState } from "react"
import type { ReactNode } from "react"
import { ThemeContext } from "@/context/theme-context"
import type { Theme } from "@/context/theme-context"

export type { Theme }

const THEME_KEY = "mahzen_theme"
const BG_ANIM_KEY = "mahzen_bg_animation"

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<Theme>(() => {
    const stored = localStorage.getItem(THEME_KEY)
    if (stored === "dark" || stored === "light") return stored
    return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light"
  })

  const [bgAnimation, setBgAnimation] = useState<boolean>(() => {
    const stored = localStorage.getItem(BG_ANIM_KEY)
    if (stored !== null) return stored === "true"
    return true
  })

  useEffect(() => {
    const root = document.documentElement
    if (theme === "dark") {
      root.classList.add("dark")
    } else {
      root.classList.remove("dark")
    }
    localStorage.setItem(THEME_KEY, theme)
  }, [theme])

  useEffect(() => {
    localStorage.setItem(BG_ANIM_KEY, String(bgAnimation))
  }, [bgAnimation])

  const setTheme = (t: Theme) => setThemeState(t)
  const toggleTheme = () => setThemeState((prev) => (prev === "dark" ? "light" : "dark"))
  const toggleBgAnimation = () => setBgAnimation((prev) => !prev)

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme, setTheme, bgAnimation, toggleBgAnimation }}>
      {children}
    </ThemeContext.Provider>
  )
}
