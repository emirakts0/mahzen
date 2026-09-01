import { useState, useEffect } from "react"
import { useSearchParams } from "react-router"
import { useAuth } from "@/hooks/use-auth"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { ApiRequestError } from "@/api/client"
import { Dialog, DialogContent, DialogTitle, DialogDescription, DialogHeader } from "@/components/ui/dialog"
import { Eye, EyeOff } from "lucide-react"

export function AuthModal() {
  const [searchParams, setSearchParams] = useSearchParams()
  const { login, register } = useAuth()
  
  const authType = searchParams.get("auth") // 'login' | 'signup' | null
  const isOpen = authType === "login" || authType === "signup"

  const [identifier, setIdentifier] = useState("")
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [displayName, setDisplayName] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)

  // Clear form when modal opens/closes
  useEffect(() => {
    if (isOpen) {
      setError(null)
      setPassword("")
      setConfirmPassword("")
      setShowPassword(false)
      setShowConfirmPassword(false)
    }
  }, [isOpen, authType])

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      searchParams.delete("auth")
      setSearchParams(searchParams, { replace: true })
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    if (authType === "signup") {
      if (password.length < 8) {
        setError("Password must be at least 8 characters.")
        return
      }
      if (password !== confirmPassword) {
        setError("Passwords do not match.")
        return
      }
      if (!/^[a-z][a-z0-9_]{2,31}$/.test(username.trim().toLowerCase())) {
        setError("Username must be 3-32 characters: lowercase letters, digits and underscores, starting with a letter.")
        return
      }
    }

    setIsLoading(true)

    try {
      if (authType === "signup") {
        await register({ username: username.trim().toLowerCase(), email: identifier.trim(), display_name: displayName, password })
      } else {
        await login({ identifier: identifier.trim(), password })
      }
      handleOpenChange(false)
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.message)
      } else {
        setError("An unexpected error occurred. Please try again.")
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <DialogContent 
        showCloseButton={false}
        className="sm:max-w-[425px] rounded-3xl p-8 border-0 shadow-2xl backdrop-blur-xl"
        style={{
          background: "var(--glass-bg)",
          border: "1px solid var(--glass-border)",
        }}
        onInteractOutside={(e) => {
          // Prevent closing when clicking the header buttons
          if ((e.target as Element).closest('header')) {
            e.preventDefault()
          }
        }}
      >
        <DialogHeader className="mb-2 text-center">
          <DialogTitle 
            className="text-3xl font-bold tracking-tight mb-2 text-center"
            style={{ color: "var(--glass-text)" }}
          >
            {authType === "signup" ? "Create an account" : "Welcome back"}
          </DialogTitle>
          <DialogDescription className="text-center" style={{ color: "var(--glass-text-muted)" }}>
            {authType === "signup" 
              ? "Start building your personal knowledge base" 
              : "Sign in to access your knowledge base"}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
          {authType === "signup" && (
            <div className="space-y-2">
              <Label htmlFor="display-name" style={{ color: "var(--glass-text)" }}>
                Display name
              </Label>
              <Input
                id="display-name"
                type="text"
                placeholder="Jane Doe"
                autoComplete="name"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                disabled={isLoading}
                className="h-11 rounded-xl bg-transparent transition-colors"
                style={{
                  borderColor: "var(--glass-border)",
                  color: "var(--glass-text)",
                }}
              />
            </div>
          )}

          {authType === "signup" && (
            <div className="space-y-2">
              <Label htmlFor="username" style={{ color: "var(--glass-text)" }}>
                Username
              </Label>
              <Input
                id="username"
                type="text"
                placeholder="jane_doe"
                autoComplete="username"
                required
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                disabled={isLoading}
                className="h-11 rounded-xl bg-transparent transition-colors"
                style={{
                  borderColor: "var(--glass-border)",
                  color: "var(--glass-text)",
                }}
              />
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="identifier" style={{ color: "var(--glass-text)" }}>
              {authType === "signup" ? "Email" : "Email or username"}
            </Label>
            <Input
              id="identifier"
              type="text"
              placeholder={authType === "signup" ? "you@example.com" : "you@example.com or username"}
              autoComplete={authType === "signup" ? "email" : "username"}
              required
              value={identifier}
              onChange={(e) => setIdentifier(e.target.value)}
              disabled={isLoading}
              className="h-11 rounded-xl bg-transparent transition-colors"
              style={{
                borderColor: "var(--glass-border)",
                color: "var(--glass-text)",
              }}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="password" style={{ color: "var(--glass-text)" }}>
              Password
            </Label>
            <div className="relative">
              <Input
                id="password"
                type={showPassword ? "text" : "password"}
                placeholder="••••••••"
                autoComplete={authType === "signup" ? "new-password" : "current-password"}
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={isLoading}
                className="h-11 rounded-xl bg-transparent transition-colors pr-10"
                style={{
                  borderColor: "var(--glass-border)",
                  color: "var(--glass-text)",
                }}
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                style={{ color: "var(--glass-text-muted)" }}
                tabIndex={-1}
              >
                {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
          </div>

          {authType === "signup" && (
            <div className="space-y-2">
              <Label htmlFor="confirm-password" style={{ color: "var(--glass-text)" }}>
                Confirm Password
              </Label>
              <div className="relative">
                <Input
                  id="confirm-password"
                  type={showConfirmPassword ? "text" : "password"}
                  placeholder="••••••••"
                  autoComplete="new-password"
                  required
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  disabled={isLoading}
                  className="h-11 rounded-xl bg-transparent transition-colors pr-10"
                  style={{
                    borderColor: "var(--glass-border)",
                    color: "var(--glass-text)",
                  }}
                />
                <button
                  type="button"
                  onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                  style={{ color: "var(--glass-text-muted)" }}
                  tabIndex={-1}
                >
                  {showConfirmPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
            </div>
          )}

          {error && (
            <div
              className="rounded-xl p-3 text-sm"
              style={{
                background: "var(--glass-error-bg)",
                border: "1px solid var(--glass-error-border)",
                color: "var(--glass-error)",
              }}
            >
              {error}
            </div>
          )}

          <Button
            type="submit"
            className="h-11 w-full rounded-xl font-medium transition-opacity hover:opacity-90"
            disabled={isLoading}
            style={{
              background: "var(--glass-text)",
              color: "var(--background)",
            }}
          >
            {isLoading 
              ? (authType === "signup" ? "Creating account..." : "Signing in...") 
              : (authType === "signup" ? "Create account" : "Sign in")}
          </Button>
        </form>

        <div
          className="mt-6 text-center text-sm"
          style={{ color: "var(--glass-text-muted)" }}
        >
          {authType === "signup" ? "Already have an account?" : "Don't have an account?"}{" "}
          <button
            type="button"
            onClick={() => {
              setSearchParams({ auth: authType === "signup" ? "login" : "signup" }, { replace: true })
            }}
            className="font-medium hover:underline underline-offset-4 transition-colors"
            style={{ color: "var(--glass-text)" }}
          >
            {authType === "signup" ? "Sign in" : "Sign up"}
          </button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
