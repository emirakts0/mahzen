import { Toaster as Sonner } from "sonner";

export function Toaster() {
  return (
    <Sonner
      position="top-right"
      toastOptions={{
        style: {
          background: "var(--glass-bg)",
          border: "1px solid var(--glass-border)",
          color: "var(--glass-text)",
        },
      }}
    />
  );
}
