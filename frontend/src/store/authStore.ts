import { create } from "zustand";

import type { User } from "../types";

interface AuthState {
  token: string;
  user?: User;
  setSession: (token: string, user: User) => void;
  clearSession: () => void;
}

const storedToken = localStorage.getItem("llm_gateway_token") ?? "";
const storedUser = localStorage.getItem("llm_gateway_user");

export const useAuthStore = create<AuthState>((set) => ({
  token: storedToken,
  user: storedUser ? (JSON.parse(storedUser) as User) : undefined,
  setSession: (token, user) => {
    localStorage.setItem("llm_gateway_token", token);
    localStorage.setItem("llm_gateway_user", JSON.stringify(user));
    set({ token, user });
  },
  clearSession: () => {
    localStorage.removeItem("llm_gateway_token");
    localStorage.removeItem("llm_gateway_user");
    set({ token: "", user: undefined });
  },
}));
