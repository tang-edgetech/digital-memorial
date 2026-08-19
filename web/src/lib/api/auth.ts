import { apiFetch } from "./client";

export type UserRole = "super_admin" | "admin" | "agent";

export interface CurrentUser {
  id: number;
  email: string;
  fullName: string;
  role: UserRole;
  isOwner: boolean;
  isActive: boolean;
  themePreference: "light" | "dark";
  permissions: Record<string, Record<string, boolean>>;
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
