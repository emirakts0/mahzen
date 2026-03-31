import { useState } from "react"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { motion, AnimatePresence } from "framer-motion"
import { Key, Plus, Copy, Check, Trash2, Loader2 } from "lucide-react"
import { listAccessTokens, createAccessToken, revokeAccessToken } from "@/api/access-token"
import { useAuth } from "@/hooks/use-auth"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import type { AccessToken } from "@/types/api"

const EXPIRY_OPTIONS = [
  { label: "30 days", value: "720h" },
  { label: "90 days", value: "2160h" },
  { label: "1 year", value: "8760h" },
]

function statusBadge(status: AccessToken["status"]) {
  const colors: Record<string, { bg: string; text: string }> = {
    active: { bg: "var(--glass-success-bg)", text: "var(--glass-success)" },
    revoked: { bg: "rgba(239,68,68,0.15)", text: "#ef4444" },
    expired: { bg: "rgba(156,163,175,0.15)", text: "#9ca3af" },
  }
  const c = colors[status] ?? colors.expired
  return (
    <span
      className="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium capitalize"
      style={{ background: c.bg, color: c.text }}
    >
      {status}
    </span>
  )
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  })
}

export default function TokensPage() {
  const { isAuthenticated } = useAuth()
  const queryClient = useQueryClient()
  const [createOpen, setCreateOpen] = useState(false)
  const [confirmRevokeId, setConfirmRevokeId] = useState<string | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ["access-tokens"],
    queryFn: listAccessTokens,
    enabled: isAuthenticated,
  })

  const createMut = useMutation({
    mutationFn: createAccessToken,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["access-tokens"] }),
  })

  const revokeMut = useMutation({
    mutationFn: revokeAccessToken,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["access-tokens"] })
      setConfirmRevokeId(null)
    },
  })

  if (!isAuthenticated) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-4 text-center">
        <div
          className="backdrop-blur-xl rounded-2xl p-8"
          style={{ background: "var(--glass-bg)", border: "1px solid var(--glass-border)" }}
        >
          <Key className="h-12 w-12 mx-auto mb-4 text-primary" />
          <h2 className="text-xl font-semibold mb-2" style={{ color: "var(--glass-text)" }}>
            Sign in to manage tokens
          </h2>
        </div>
      </div>
    )
  }

  const tokens = data?.tokens ?? []

  return (
    <div className="mx-auto max-w-3xl px-4 pt-24 pb-12">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold" style={{ color: "var(--glass-text)" }}>
            API Tokens
          </h1>
          <p className="text-xs mt-1" style={{ color: "var(--glass-text-muted)" }}>
            Manage access tokens for the desktop app and API clients
          </p>
        </div>
        <button
          onClick={() => setCreateOpen(true)}
          className="flex items-center gap-1.5 rounded-xl px-3 py-1.5 text-xs font-medium transition-colors"
          style={{ background: "var(--glass-hover)", color: "var(--glass-text)", border: "1px solid var(--glass-border)" }}
        >
          <Plus className="h-3.5 w-3.5" />
          Generate Token
        </button>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="h-6 w-6 animate-spin" style={{ color: "var(--glass-text-muted)" }} />
        </div>
      ) : tokens.length === 0 ? (
        <div
          className="backdrop-blur-xl rounded-2xl p-8 text-center"
          style={{ background: "var(--glass-bg)", border: "1px solid var(--glass-border)" }}
        >
          <Key className="h-8 w-8 mx-auto mb-3" style={{ color: "var(--glass-text-muted)" }} />
          <p className="text-sm" style={{ color: "var(--glass-text-muted)" }}>
            No tokens yet. Generate one to connect the desktop app.
          </p>
        </div>
      ) : (
        <div className="space-y-2">
          <AnimatePresence>
            {tokens.map((token) => (
              <motion.div
                key={token.id}
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -8 }}
                className="backdrop-blur-xl rounded-xl p-4 flex items-center gap-4"
                style={{ background: "var(--glass-bg)", border: "1px solid var(--glass-border)" }}
              >
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-sm font-medium truncate" style={{ color: "var(--glass-text)" }}>
                      {token.name}
                    </span>
                    {statusBadge(token.status)}
                  </div>
                  <div className="flex items-center gap-3 text-xs" style={{ color: "var(--glass-text-muted)" }}>
                    <span className="font-mono">{token.prefix}</span>
                    <span>Expires {formatDate(token.expires_at)}</span>
                    <span>Created {formatDate(token.created_at)}</span>
                  </div>
                </div>

                {token.status === "active" && (
                  confirmRevokeId === token.id ? (
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => revokeMut.mutate(token.id)}
                        disabled={revokeMut.isPending}
                        className="flex items-center gap-1 rounded-lg px-2 py-1 text-xs font-medium"
                        style={{ background: "rgba(239,68,68,0.15)", color: "#ef4444" }}
                      >
                        {revokeMut.isPending ? <Loader2 className="h-3 w-3 animate-spin" /> : "Confirm"}
                      </button>
                      <button
                        onClick={() => setConfirmRevokeId(null)}
                        className="rounded-lg px-2 py-1 text-xs"
                        style={{ color: "var(--glass-text-muted)" }}
                      >
                        Cancel
                      </button>
                    </div>
                  ) : (
                    <button
                      onClick={() => setConfirmRevokeId(token.id)}
                      className="flex items-center gap-1 rounded-lg px-2 py-1 text-xs transition-colors"
                      style={{ color: "var(--glass-text-muted)" }}
                      onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.color = "#ef4444" }}
                      onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.color = "var(--glass-text-muted)" }}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                      Revoke
                    </button>
                  )
                )}
              </motion.div>
            ))}
          </AnimatePresence>
        </div>
      )}

      {/* Create Token Dialog */}
      <Dialog open={createOpen} onOpenChange={(open) => { if (!open) createMut.reset() }}>
        <DialogContent
          className="rounded-xl border-0 shadow-xl backdrop-blur-xl"
          style={{ background: "var(--glass-bg)", border: "1px solid var(--glass-border)" }}
        >
          {createMut.data ? (
            <TokenCreatedDialog rawToken={createMut.data.raw_token} onClose={() => { createMut.reset(); setCreateOpen(false) }} />
          ) : (
            <CreateTokenForm
              isPending={createMut.isPending}
              onSubmit={(name, expiresIn) => createMut.mutate({ name, expires_in: expiresIn })}
              error={createMut.error?.message}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}

function CreateTokenForm({
  isPending,
  onSubmit,
  error,
}: {
  isPending: boolean
  onSubmit: (name: string, expiresIn: string) => void
  error?: string
}) {
  const [name, setName] = useState("")
  const [expiresIn, setExpiresIn] = useState(EXPIRY_OPTIONS[1].value)

  return (
    <>
      <DialogHeader>
        <DialogTitle style={{ color: "var(--glass-text)" }}>Generate Token</DialogTitle>
      </DialogHeader>
      <form
        onSubmit={(e) => { e.preventDefault(); onSubmit(name, expiresIn) }}
        className="space-y-4 mt-2"
      >
        <div>
          <label className="mb-1 block text-xs font-medium" style={{ color: "var(--glass-text-muted)" }}>
            Token Name
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            placeholder="e.g. My Desktop"
            className="h-9 w-full rounded-md border px-3 text-sm outline-none"
            style={{
              background: "var(--glass-hover)",
              borderColor: "var(--glass-border)",
              color: "var(--glass-text)",
            }}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium" style={{ color: "var(--glass-text-muted)" }}>
            Expires In
          </label>
          <div className="flex gap-2">
            {EXPIRY_OPTIONS.map((opt) => (
              <button
                key={opt.value}
                type="button"
                onClick={() => setExpiresIn(opt.value)}
                className="flex-1 rounded-lg py-1.5 text-xs font-medium transition-colors"
                style={{
                  background: expiresIn === opt.value ? "var(--glass-hover)" : "transparent",
                  border: `1px solid ${expiresIn === opt.value ? "var(--glass-border)" : "transparent"}`,
                  color: expiresIn === opt.value ? "var(--glass-text)" : "var(--glass-text-muted)",
                }}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </div>
        {error && <p className="text-xs" style={{ color: "#ef4444" }}>{error}</p>}
        <button
          type="submit"
          disabled={isPending}
          className="flex h-9 w-full items-center justify-center rounded-md text-sm font-medium transition-opacity disabled:opacity-50"
          style={{ background: "var(--color-primary)", color: "var(--color-primary-foreground)" }}
        >
          {isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : "Generate"}
        </button>
      </form>
    </>
  )
}

function TokenCreatedDialog({ rawToken, onClose }: { rawToken: string; onClose: () => void }) {
  const [copied, setCopied] = useState(false)

  const copyToken = async () => {
    await navigator.clipboard.writeText(rawToken)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle style={{ color: "var(--glass-text)" }}>Token Created</DialogTitle>
      </DialogHeader>
      <div className="mt-2 space-y-3">
        <p className="text-xs" style={{ color: "var(--glass-text-muted)" }}>
          Copy this token now. It won't be shown again.
        </p>
        <div
          className="flex items-center gap-2 rounded-lg p-3"
          style={{ background: "var(--glass-hover)", border: "1px solid var(--glass-border)" }}
        >
          <code className="flex-1 text-xs break-all font-mono" style={{ color: "var(--glass-text)" }}>
            {rawToken}
          </code>
          <button
            onClick={copyToken}
            className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md transition-colors"
            style={{ color: copied ? "var(--glass-success)" : "var(--glass-text-muted)" }}
          >
            {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
          </button>
        </div>
        <button
          onClick={onClose}
          className="flex h-9 w-full items-center justify-center rounded-md text-sm font-medium transition-opacity"
          style={{ background: "var(--glass-hover)", color: "var(--glass-text)", border: "1px solid var(--glass-border)" }}
        >
          Done
        </button>
      </div>
    </>
  )
}
