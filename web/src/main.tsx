import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import "./index.css"
import App from "./App.tsx"

// Apply stored theme before first render to avoid flash
const stored = localStorage.getItem("theme") ?? "system"
const resolved =
  stored === "system"
    ? window.matchMedia("(prefers-color-scheme: dark)").matches
      ? "dark"
      : "light"
    : stored
document.documentElement.classList.toggle("dark", resolved === "dark")

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>
)
