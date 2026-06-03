import { Navigate, RouterProvider, createBrowserRouter } from "react-router";

import { RequireRole } from "./components/AuthGuards";
import { AdminLayout } from "./layouts/AdminLayout";
import { PortalLayout } from "./layouts/PortalLayout";
import { DashboardPage } from "./pages/DashboardPage";
import { LandingPage } from "./pages/LandingPage";
import { LogDetailPage } from "./pages/LogDetailPage";
import { LoginPage } from "./pages/LoginPage";
import { ModelDetailPage } from "./pages/ModelDetailPage";
import { ModelFamiliesPage } from "./pages/ModelFamiliesPage";
import { ProviderDetailPage } from "./pages/ProviderDetailPage";
import { ProvidersPage } from "./pages/ProvidersPage";
import { LogsPage } from "./pages/LogsPage";
import { PortalKeyDetailPage } from "./pages/PortalKeyDetailPage";
import { PortalKeysPage } from "./pages/PortalKeysPage";
import { PortalLogDetailPage } from "./pages/PortalLogDetailPage";
import { UsersPage } from "./pages/UsersPage";

const router = createBrowserRouter([
  { path: "/", element: <LandingPage /> },
  { path: "/login", element: <LoginPage /> },
  {
    element: <RequireRole role="admin" />,
    children: [
      {
        element: <AdminLayout />,
        children: [
          { path: "dashboard", element: <DashboardPage /> },
          { path: "providers", element: <ProvidersPage /> },
          { path: "providers/:id", element: <ProviderDetailPage /> },
          { path: "models", element: <ModelFamiliesPage /> },
          { path: "model-families", element: <Navigate to="/models" replace /> },
          { path: "models/:id", element: <ModelDetailPage /> },
          { path: "users", element: <UsersPage /> },
          { path: "logs", element: <LogsPage /> },
          { path: "logs/:id", element: <LogDetailPage /> },
        ],
      },
    ],
  },
  {
    path: "/portal",
    element: <RequireRole role="user" />,
    children: [
      {
        element: <PortalLayout />,
        children: [
          { index: true, element: <Navigate to="/portal/keys" replace /> },
          { path: "keys", element: <PortalKeysPage /> },
          { path: "keys/:id", element: <PortalKeyDetailPage /> },
          { path: "keys/:keyID/logs/:logID", element: <PortalLogDetailPage /> },
        ],
      },
    ],
  },
]);

export function AppRouter() {
  return <RouterProvider router={router} />;
}
