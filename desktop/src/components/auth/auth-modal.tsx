import { useState, useEffect } from "react";
import { X, Loader2 } from "lucide-react";
import { useAuth } from "@/providers/auth-provider";

interface AuthModalProps {
  isOpen: boolean;
  onClose: () => void;
  defaultMode?: "login" | "signup";
}

export function AuthModal({ isOpen, onClose, defaultMode = "login" }: AuthModalProps) {
  const [mode, setMode] = useState<"login" | "signup">(defaultMode);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const { login, register } = useAuth();

  useEffect(() => {
    if (isOpen) {
      setMode(defaultMode);
    }
  }, [isOpen, defaultMode]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);

    try {
      if (mode === "login") {
        await login({ email, password });
      } else {
        await register({ email, password, display_name: displayName || undefined });
      }
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setIsLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div
      className="absolute inset-0 z-50 flex items-center justify-center"
      style={{ background: "rgba(0, 0, 0, 0.5)", backdropFilter: "blur(4px)" }}
    >
      <div
        className="w-80 rounded-xl border p-4"
        style={{
          background: "var(--glass-bg)",
          borderColor: "var(--glass-border)",
        }}
      >
        <div className="flex items-center justify-between mb-4">
          <h2
            className="text-sm font-semibold"
            style={{ color: "var(--glass-text)" }}
          >
            {mode === "login" ? "Sign In" : "Create Account"}
          </h2>
          <button
            onClick={onClose}
            className="flex h-6 w-6 items-center justify-center rounded-md hover:opacity-80"
            style={{ color: "var(--glass-icon)" }}
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-3">
          {mode === "signup" && (
            <div>
              <label
                className="mb-1 block text-xs"
                style={{ color: "var(--glass-text-muted)" }}
              >
                Display Name
              </label>
              <input
                type="text"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                className="h-9 w-full rounded-md border px-3 text-sm outline-none"
                style={{
                  background: "var(--glass-hover)",
                  borderColor: "var(--glass-border)",
                  color: "var(--glass-text)",
                }}
                placeholder="Your name"
              />
            </div>
          )}

          <div>
            <label
              className="mb-1 block text-xs"
              style={{ color: "var(--glass-text-muted)" }}
            >
              Email
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className="h-9 w-full rounded-md border px-3 text-sm outline-none"
              style={{
                background: "var(--glass-hover)",
                borderColor: "var(--glass-border)",
                color: "var(--glass-text)",
              }}
              placeholder="you@example.com"
            />
          </div>

          <div>
            <label
              className="mb-1 block text-xs"
              style={{ color: "var(--glass-text-muted)" }}
            >
              Password
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={8}
              className="h-9 w-full rounded-md border px-3 text-sm outline-none"
              style={{
                background: "var(--glass-hover)",
                borderColor: "var(--glass-border)",
                color: "var(--glass-text)",
              }}
              placeholder="********"
            />
          </div>

          {error && (
            <p
              className="text-xs"
              style={{ color: "var(--color-destructive)" }}
            >
              {error}
            </p>
          )}

          <button
            type="submit"
            disabled={isLoading}
            className="flex h-9 w-full items-center justify-center rounded-md text-sm font-medium transition-opacity disabled:opacity-50"
            style={{
              background: "var(--color-primary)",
              color: "var(--color-primary-foreground)",
            }}
          >
            {isLoading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : mode === "login" ? (
              "Sign In"
            ) : (
              "Create Account"
            )}
          </button>
        </form>

        <div className="mt-3 text-center">
          <button
            onClick={() => setMode(mode === "login" ? "signup" : "login")}
            className="text-xs transition-opacity hover:opacity-80"
            style={{ color: "var(--glass-text-muted)" }}
          >
            {mode === "login"
              ? "Don't have an account? Sign up"
              : "Already have an account? Sign in"}
          </button>
        </div>
      </div>
    </div>
  );
}
