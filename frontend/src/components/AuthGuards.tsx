import { Navigate, Outlet, useLocation } from "react-router";

import { useAuthStore } from "../store/authStore";

export function RequireRole({ role }: { role: "admin" | "user" }) {
  const { token, user } = useAuthStore();
  const location = useLocation();

  if (!token || !user) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }
  if (user.role !== role) {
    return <Navigate to={user.role === "admin" ? "/dashboard" : "/portal/keys"} replace />;
  }
  return <Outlet />;
}
