import { Suspense, lazy } from "react"
import { createBrowserRouter, RouterProvider, Navigate } from "react-router"
import RootLayout from "@/components/layout/root-layout"

const SearchPage = lazy(() => import("@/pages/search-page"))
const EntriesPage = lazy(() => import("@/pages/entries-page"))

function PageLoader() {
  return (
    <div className="flex h-[calc(100vh-5rem)] items-center justify-center">
      <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
    </div>
  )
}

const router = createBrowserRouter([
  {
    path: "/",
    element: <RootLayout />,
    children: [
      {
        index: true,
        element: <Navigate to="/search" replace />,
      },
      {
        path: "search",
        element: (
          <Suspense fallback={<PageLoader />}>
            <SearchPage />
          </Suspense>
        ),
      },
      {
        path: "entries",
        element: (
          <Suspense fallback={<PageLoader />}>
            <EntriesPage />
          </Suspense>
        ),
      },
      {
        path: "login",
        element: <Navigate to="/search?auth=login" replace />,
      },
      {
        path: "signup",
        element: <Navigate to="/search?auth=signup" replace />,
      },
    ],
  },
])

export default function App() {
  return <RouterProvider router={router} />
}
