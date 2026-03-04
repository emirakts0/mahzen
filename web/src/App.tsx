import { createBrowserRouter, RouterProvider, Navigate } from "react-router"
import RootLayout from "@/components/layout/root-layout"
import SearchPage from "@/pages/search-page"
import EntriesPage from "@/pages/entries-page"
import EntryDetailPage from "@/pages/entry-detail-page"
import TagsPage from "@/pages/tags-page"
import LoginPage from "@/pages/login-page"
import SignupPage from "@/pages/signup-page"

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
        path: "entries/:entryId",
        element: <EntryDetailPage />,
      },
      {
        path: "tags",
        element: <TagsPage />,
      },
    ],
  },
  {
    path: "/login",
    element: <LoginPage />,
  },
  {
    path: "/signup",
    element: <SignupPage />,
  },
])

export default function App() {
  return <RouterProvider router={router} />
}
