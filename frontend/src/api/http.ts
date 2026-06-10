import axios from "axios";
import { useAuthStore } from "../store/authStore";

const apiBaseURL = import.meta.env.VITE_API_BASE_URL || (import.meta.env.PROD ? "" : "http://127.0.0.1:3212");

export const apiClient = axios.create({
  baseURL: apiBaseURL,
  timeout: 30_000,
});

apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && error.response.status === 401) {
      useAuthStore.getState().clearSession();
      // Public pages shouldn't be abruptly redirected when a background token check fails.
      // They just gracefully degrade to the logged-out state because Zustand updates localStorage and triggers re-renders.
      const publicPaths = ["/", "/login"];
      if (!publicPaths.includes(window.location.pathname)) {
        window.location.href = "/login";
      }
    }
    return Promise.reject(error);
  }
);
