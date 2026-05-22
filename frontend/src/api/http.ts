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
