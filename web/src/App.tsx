import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { Toaster } from "@/components/ui/sonner"
import { MainLayout } from "@/components/layout/main-layout"
import { LoginPage } from "@/pages/login"
import { RegisterPage } from "@/pages/register"
import { EntriesPage } from "@/pages/entries"
import { NewEntryPage } from "@/pages/entry-new"
import { EntryDetailPage } from "@/pages/entry-detail"
import { EntryEditPage } from "@/pages/entry-edit"
import { TagsPage } from "@/pages/tags"
import { SearchPage } from "@/pages/search"

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 30, // 30 seconds
      retry: 1,
    },
  },
})

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          {/* Auth routes — no layout */}
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />

          {/* App routes — with layout */}
          <Route element={<MainLayout />}>
            <Route path="/" element={<Navigate to="/entries" replace />} />
            <Route path="/entries" element={<EntriesPage />} />
            <Route path="/entries/new" element={<NewEntryPage />} />
            <Route path="/entries/:id" element={<EntryDetailPage />} />
            <Route path="/entries/:id/edit" element={<EntryEditPage />} />
            <Route path="/tags" element={<TagsPage />} />
            <Route path="/search" element={<SearchPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
      <Toaster />
    </QueryClientProvider>
  )
}

export default App
