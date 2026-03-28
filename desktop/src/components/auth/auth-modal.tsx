import { useState, useEffect } from "react";
import { X, Loader2 } from "lucide-react";
import { useAuth } from "@/providers/auth-provider";

interface AuthModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function AuthModal({ isOpen, onClose }: AuthModalProps) {
  const [serverUrl, setServerUrl] = useState("https://localhost:8080");
  const [accessToken, setAccessToken] = useState("");
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const { connect } = useAuth();

  useEffect(() => {
    if (isOpen) {
      setError("");
    }
  }, [isOpen]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);

    try {
      await connect(serverUrl, accessToken);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Connection failed");
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
            Connect to Server
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
          <div>
            <label
              className="mb-1 block text-xs"
              style={{ color: "var(--glass-text-muted)" }}
            >
              Server URL
            </label>
            <input
              type="url"
              value={serverUrl}
              onChange={(e) => setServerUrl(e.target.value)}
              required
              className="h-9 w-full rounded-md border px-3 text-sm outline-none"
              style={{
                background: "var(--glass-hover)",
                borderColor: "var(--glass-border)",
                color: "var(--glass-text)",
              }}
              placeholder="https://your-server.com"
            />
          </div>

          <div>
            <label
              className="mb-1 block text-xs"
              style={{ color: "var(--glass-text-muted)" }}
            >
              Access Token
            </label>
            <input
              type="password"
              value={accessToken}
              onChange={(e) => setAccessToken(e.target.value)}
              required
              className="h-9 w-full rounded-md border px-3 text-sm outline-none font-mono"
              style={{
                background: "var(--glass-hover)",
                borderColor: "var(--glass-border)",
                color: "var(--glass-text)",
              }}
              placeholder="mah_..."
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
            ) : (
              "Connect"
            )}
          </button>
        </form>
      </div>
    </div>
  );
}
