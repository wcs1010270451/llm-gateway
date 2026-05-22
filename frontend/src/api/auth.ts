import { apiClient } from "./http";
import type { User } from "../types";

interface LoginResponse {
  token: string;
  user: User;
}

export async function login(email: string, password: string) {
  const response = await apiClient.post<LoginResponse>("/api/auth/login", { email, password });
  return response.data;
}

export async function fetchMe() {
  const response = await apiClient.get<User>("/api/auth/me");
  return response.data;
}
