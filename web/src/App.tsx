import { createBrowserRouter, RouterProvider, Navigate } from "react-router"
import RootLayout from "@/components/layout/root-layout"
import SearchPage from "@/pages/search-page"
import EntriesPage from "@/pages/entries-page"

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
        element: <SearchPage />,
      },
      {
        path: "entries",
        element: <EntriesPage />,
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
