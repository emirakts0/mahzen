import { createContext } from "react"

export type Theme = "light" | "dark"

export interface ThemeContextValue {
  theme: Theme
  toggleTheme: () => void
  setTheme: (theme: Theme) => void
  bgAnimation: boolean
  toggleBgAnimation: () => void
}

export const ThemeContext = createContext<ThemeContextValue | null>(null)
