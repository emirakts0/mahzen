import { useState, useEffect } from "react";
import { invoke } from "@tauri-apps/api/core";
import { X, Loader2 } from "lucide-react";
import { clearBaseUrlCache } from "@/api/client";

interface AppConfig {
  backend_url: string;
}

interface SettingsPanelProps {
  isOpen: boolean;
  onClose: () => void;
  onSave?: (url: string) => void;
}

export function SettingsPanel({ isOpen, onClose, onSave }: SettingsPanelProps) {
  const [url, setUrl] = useState("");
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (isOpen) {
      loadConfig();
    }
  }, [isOpen]);

  const loadConfig = async () => {
    try {
      const config = await invoke<AppConfig>("get_config");
      setUrl(config.backend_url);
      setError("");
    } catch (e) {
      setError("Failed to load config");
    }
  };

  const saveConfig = async () => {
    setIsSaving(true);
    try {
      await invoke("set_config", { config: { backend_url: url } });
      clearBaseUrlCache();
      onSave?.(url);
    } catch (e) {
      setError("Failed to save config");
    } finally {
      setIsSaving(false);
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
            Settings
          </h2>
          <button
            onClick={onClose}
            className="flex h-6 w-6 items-center justify-center rounded-md hover:opacity-80"
            style={{ color: "var(--glass-icon)" }}
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="space-y-3">
          <div>
            <label
              className="mb-1 block text-xs"
              style={{ color: "var(--glass-text-muted)" }}
            >
              Backend URL
            </label>
            <input
              type="text"
              value={url}
              onChange={(e) => {
                setUrl(e.target.value);
                setError("");
              }}
              className="h-9 w-full rounded-md border px-3 text-sm outline-none"
              style={{
                background: "var(--glass-hover)",
                borderColor: "var(--glass-border)",
                color: "var(--glass-text)",
              }}
              placeholder="https://localhost:8080"
            />
          </div>

          {error && (
             <p className="text-xs" style={{ color: "var(--color-destructive)" }}>
               {error}
             </p>
           )}

           <div className="flex gap-2 pt-2">
             <button
               onClick={saveConfig}
               disabled={isSaving || !url}
               className="flex h-8 w-full items-center justify-center gap-1 rounded-md text-xs font-medium transition-opacity disabled:opacity-50"
               style={{
                 background: "var(--color-primary)",
                 color: "var(--color-primary-foreground)",
               }}
             >
               {isSaving ? (
                 <Loader2 className="h-3 w-3 animate-spin" />
               ) : (
                 "Save"
               )}
             </button>
           </div>
        </div>
      </div>
    </div>
  );
}
