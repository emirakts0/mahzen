import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router";
import { motion } from "framer-motion";
import {
  Plus,
  BookOpen,
  RefreshCw,
  FolderTree as FolderTreeIcon,
  Filter,
} from "lucide-react";
import { listEntries } from "@/api/entries";
import { listTags } from "@/api/tags";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useAuth } from "@/hooks/use-auth";
import { useTheme } from "@/hooks/use-theme";
import { FolderTree, CreateEntryDialog } from "@/components/entries";
import { EntryPreviewModal } from "@/components/search/entry-preview-modal";

export default function EntriesPage() {
  const { isAuthenticated } = useAuth();
  const { theme } = useTheme();
  const [createOpen, setCreateOpen] = useState(false);
  const [previewEntryId, setPreviewEntryId] = useState<string | null>(null);
  const [ownOnly, setOwnOnly] = useState(false);
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [tagSearch, setTagSearch] = useState("");
  const queryClient = useQueryClient();

  // Use same query key as FolderTree to share cache
  const { data, isLoading, isFetching } = useQuery({
    queryKey: ["entries", "/", ownOnly],
    queryFn: () => listEntries({ path: "/", limit: 0, own: ownOnly }),
    enabled: isAuthenticated,
  });

  const { data: tagsData } = useQuery({
    queryKey: ["tags"],
    queryFn: () => listTags({ limit: 100 }),
    enabled: isAuthenticated,
  });

  const total = data?.total ?? 0;
  const isRefetching = isFetching && !isLoading;

  const filteredTags = tagsData?.tags?.filter((t) =>
    t.name.toLowerCase().includes(tagSearch.toLowerCase())
  ) ?? [];

  const toggleTag = (id: string) => {
    setSelectedTags((prev) =>
      prev.includes(id) ? prev.filter((t) => t !== id) : [...prev, id]
    );
  };

  const handleRefresh = () => {
    queryClient.invalidateQueries({ queryKey: ["entries"] });
  };

  if (!isAuthenticated) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-4 text-center">
        <div
          className="backdrop-blur-xl rounded-2xl p-8"
          style={{
            background: "var(--glass-bg)",
            border: "1px solid var(--glass-border)",
          }}
        >
          <BookOpen className="h-12 w-12 mx-auto mb-4 text-primary" />
          <h2
            className="text-xl font-semibold mb-2"
            style={{ color: "var(--glass-text)" }}
          >
            Sign in to view entries
          </h2>
          <p
            className="text-sm mb-6"
            style={{ color: "var(--glass-text-muted)" }}
          >
            Access your knowledge base from anywhere
          </p>
          <div className="flex gap-3 justify-center">
            <Button asChild>
              <Link to="?auth=login" replace>
                Sign in
              </Link>
            </Button>
            <Button variant="outline" asChild>
              <Link to="?auth=signup" replace>
                Create account
              </Link>
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full min-h-0">
      <div className="mx-auto w-full max-w-5xl px-4 py-6">
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-4 rounded-xl px-4 py-2"
          style={{
            background: "var(--glass-bg)",
            border: "1px solid var(--glass-border)",
          }}
        >
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <h1
                className="text-2xl font-bold flex items-center gap-2"
                style={{ color: "var(--glass-text)" }}
              >
                <FolderTreeIcon className="h-6 w-6" style={{ color: theme === "dark" ? "var(--glass-text)" : "black" }} />
                Entries
              </h1>
              <span
                className="text-sm mt-1"
                style={{ color: "var(--glass-text-muted)" }}
              >
                {total}
              </span>
            </div>
            <div className="flex items-center gap-4">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    className="backdrop-blur-md gap-2"
                    style={{
                      background: "var(--glass-bg)",
                      border: "1px solid var(--glass-border)",
                      color: "var(--glass-text)",
                    }}
                  >
                    <Filter className="h-4 w-4" />
                    Filter
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  align="end"
                  className="w-64 rounded-xl border-0 shadow-xl backdrop-blur-xl p-2"
                  style={{
                    background: "var(--glass-bg)",
                    border: "1px solid var(--glass-border)",
                  }}
                >
                  <div className="space-y-2 px-2 py-1.5">
                    <label className="flex items-center justify-between cursor-pointer">
                      <span className="text-sm" style={{ color: "var(--glass-text)" }}>
                        Only My Entries
                      </span>
                      <Switch
                        checked={ownOnly}
                        onCheckedChange={setOwnOnly}
                        className="data-[state=checked]:bg-[var(--glass-border)] data-[state=unchecked]:bg-white/10"
                      />
                    </label>
                  </div>

                  <div className="h-px my-2" style={{ background: "var(--glass-divider)" }} />

                  <div className="space-y-2 px-2">
                    <Input
                      value={tagSearch}
                      onChange={(e) => setTagSearch(e.target.value)}
                      placeholder="Search tags..."
                      className="h-8 text-sm backdrop-blur-md placeholder:text-[var(--glass-text-muted)]"
                      style={{
                        background: "var(--glass-bg)",
                        borderColor: "var(--glass-border)",
                        color: "var(--glass-text)",
                      }}
                    />
                  </div>

                  {filteredTags.length > 0 && (
                    <div className="mt-2 max-h-40 overflow-y-auto px-2 pb-2">
                      <div className="flex flex-wrap gap-1.5">
                        {filteredTags.map((t) => (
                          <button
                            key={t.id}
                            type="button"
                            onClick={() => toggleTag(t.id)}
                            className="rounded-full border px-2 py-0.5 text-xs font-medium transition-colors backdrop-blur-sm"
                            style={{
                              background: selectedTags.includes(t.id)
                                ? "var(--glass-hover)"
                                : "transparent",
                              borderColor: selectedTags.includes(t.id)
                                ? "var(--glass-border)"
                                : "var(--glass-border)",
                              color: selectedTags.includes(t.id)
                                ? "var(--glass-text)"
                                : "var(--glass-text-muted)",
                            }}
                          >
                            {t.name}
                          </button>
                        ))}
                      </div>
                    </div>
                  )}

                  {filteredTags.length === 0 && tagSearch && (
                    <div className="px-2 py-2 text-xs" style={{ color: "var(--glass-text-muted)" }}>
                      No tags found
                    </div>
                  )}
                </DropdownMenuContent>
              </DropdownMenu>
              <Button
                variant="ghost"
                size="icon"
                onClick={handleRefresh}
                disabled={isRefetching}
                className="backdrop-blur-md"
                style={{
                  background: "var(--glass-bg)",
                  border: "1px solid var(--glass-border)",
                  color: "var(--glass-text)",
                }}
              >
                <RefreshCw
                  className={`h-4 w-4 ${isRefetching ? "animate-spin" : ""}`}
                />
              </Button>
              <Button
                onClick={() => setCreateOpen(true)}
                className="backdrop-blur-md"
                style={{
                  background: "var(--glass-bg)",
                  border: "1px solid var(--glass-border)",
                  color: "var(--glass-text)",
                }}
              >
                <Plus className="mr-2 h-4 w-4" />
                New Entry
              </Button>
            </div>
          </div>
        </motion.div>

        {isLoading ? (
          <div className="space-y-2">
            {[1, 2, 3, 4, 5].map((i) => (
              <motion.div
                key={i}
                className="h-10 rounded-md animate-pulse"
                style={{
                  background: "var(--glass-bg)",
                  width: `${50 + Math.random() * 50}%`,
                  marginLeft: `${Math.random() * 20}px`,
                }}
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: i * 0.05 }}
              />
            ))}
          </div>
        ) : total === 0 ? (
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="flex flex-col items-center justify-center py-20"
          >
            <div
              className="backdrop-blur-xl rounded-2xl p-8 text-center"
              style={{
                background: "var(--glass-bg)",
                border: "1px solid var(--glass-border)",
              }}
            >
              <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl mx-auto bg-primary/10">
                <BookOpen className="h-8 w-8 text-primary" />
              </div>
              <h2
                className="text-xl font-semibold mb-2"
                style={{ color: "var(--glass-text)" }}
              >
                No entries yet
              </h2>
              <p
                className="text-sm mb-6"
                style={{ color: "var(--glass-text-muted)" }}
              >
                Create your first knowledge base entry
              </p>
              <Button
                onClick={() => setCreateOpen(true)}
                className="backdrop-blur-md"
                style={{
                  background: "var(--glass-bg)",
                  border: "1px solid var(--glass-border)",
                  color: "var(--glass-text)",
                }}
              >
                <Plus className="mr-2 h-4 w-4" />
                New Entry
              </Button>
            </div>
          </motion.div>
        ) : (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.1 }}
            className="backdrop-blur-xl rounded-xl overflow-hidden"
            style={{
              background: "var(--glass-bg)",
              border: "1px solid var(--glass-border)",
            }}
          >
            <FolderTree onEntrySelect={setPreviewEntryId} ownOnly={ownOnly} />
          </motion.div>
        )}
      </div>

      <CreateEntryDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        defaultPath="/"
      />

      <EntryPreviewModal
        entryId={previewEntryId}
        onClose={() => setPreviewEntryId(null)}
      />
    </div>
  );
}
