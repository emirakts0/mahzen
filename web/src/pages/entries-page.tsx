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
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useAuth } from "@/hooks/use-auth";
import { useTheme } from "@/hooks/use-theme";
import { FolderTree, CreateEntryDialog } from "@/components/entries";
import { EntryPreviewModal } from "@/components/search/entry-preview-modal";
import {
  EntryFiltersPanel,
  countActiveFilters,
  type EntryFilters,
} from "@/components/shared/entry-filters";

export default function EntriesPage() {
  const { isAuthenticated } = useAuth();
  const { theme } = useTheme();
  const [createOpen, setCreateOpen] = useState(false);
  const [previewEntryId, setPreviewEntryId] = useState<string | null>(null);
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [filters, setFilters] = useState<EntryFilters>({});
  const queryClient = useQueryClient();

  const ownOnly = filters.only_mine ?? false;

  const { data, isLoading, isFetching } = useQuery({
    queryKey: ["entries", "/", ownOnly, filters],
    queryFn: () =>
      listEntries({
        path: "/",
        limit: 0,
        own: ownOnly,
        visibility: filters.visibility,
        tags: filters.tags,
        from_date: filters.from_date,
        to_date: filters.to_date,
      }),
    enabled: isAuthenticated,
  });

  const total = data?.total ?? 0;
  const isRefetching = isFetching && !isLoading;

  const handleRefresh = () => {
    queryClient.invalidateQueries({ queryKey: ["entries"] });
  };

  const activeFilterCount = countActiveFilters(filters);

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
            <div className="flex items-center gap-2">
              <DropdownMenu open={filtersOpen} onOpenChange={setFiltersOpen}>
                <DropdownMenuTrigger asChild>
                  <button
                    className="flex items-center justify-center h-9 w-9 rounded-md transition-all hover:opacity-80 cursor-pointer backdrop-blur-md"
                    style={{
                      background: "var(--glass-bg)",
                      border: "1px solid var(--glass-border)",
                      color: "var(--glass-text)",
                    }}
                  >
                    <Filter className="h-4 w-4" />
                    {activeFilterCount > 0 && (
                      <span className="absolute -top-1 -right-1 flex h-4 min-w-4 items-center justify-center rounded-full text-[10px] font-semibold" style={{ background: "var(--color-primary)", color: "var(--color-primary-foreground)" }}>
                        {activeFilterCount}
                      </span>
                    )}
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  align="end"
                  className="p-0 border-0 shadow-none bg-transparent"
                >
                  <EntryFiltersPanel
                    filters={filters}
                    onFiltersChange={setFilters}
                    isOpen={filtersOpen}
                    onClose={() => setFiltersOpen(false)}
                  />
                </DropdownMenuContent>
              </DropdownMenu>
              <button
                onClick={handleRefresh}
                disabled={isRefetching}
                className="flex items-center justify-center h-9 w-9 rounded-md transition-all hover:opacity-80 cursor-pointer backdrop-blur-md disabled:opacity-50"
                style={{
                  background: "var(--glass-bg)",
                  border: "1px solid var(--glass-border)",
                  color: "var(--glass-text)",
                }}
              >
                <RefreshCw
                  className={`h-4 w-4 ${isRefetching ? "animate-spin" : ""}`}
                />
              </button>
              <button
                onClick={() => setCreateOpen(true)}
                className="flex items-center gap-1.5 h-9 px-3 rounded-md text-sm font-medium transition-all hover:opacity-80 cursor-pointer backdrop-blur-md"
                style={{
                  background: "var(--glass-bg)",
                  border: "1px solid var(--glass-border)",
                  color: "var(--glass-text)",
                }}
              >
                <Plus className="h-4 w-4" />
                New
              </button>
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
            <FolderTree
              onEntrySelect={setPreviewEntryId}
              filters={{
                own: filters.only_mine,
                visibility: filters.visibility,
                tags: filters.tags,
                from_date: filters.from_date,
                to_date: filters.to_date,
              }}
            />
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
