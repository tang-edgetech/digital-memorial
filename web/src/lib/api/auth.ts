import { apiFetch } from "./client";

export interface CurrentUser {
  id: number;
  email: string;
  fullName: string;
  role: "super_admin" | "admin" | "agent";
  themePreference: "light" | "dark";
}

export function login(email: string, password: string) {
  return apiFetch<{ user: CurrentUser }>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
    skipSessionExpiredHandling: true,
  });
}

export function logout() {
  return apiFetch<{ success: boolean }>("/api/auth/logout", {
    method: "POST",
    skipSessionExpiredHandling: true,
  });
}

export function getCurrentUser() {
  return apiFetch<{ user: CurrentUser }>("/api/auth/me", {
    skipSessionExpiredHandling: true,
  });
}
