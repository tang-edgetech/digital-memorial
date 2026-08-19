import { apiFetch } from "./client";
import type { UserRole } from "./auth";

export interface UserRecord {
  id: number;
  email: string;
  fullName: string;
  role: UserRole;
  isOwner: boolean;
  isActive: boolean;
  themePreference: "light" | "dark";
  permissions: Record<string, Record<string, boolean>>;
}

export interface ListUsersParams {
  role?: string;
  status?: string;
  search?: string;
  sortBy?: string;
  sortDir?: string;
  page?: number;
  pageSize?: number;
}

export interface BulkFailure {
  id: number;
  reason: string;
}

function buildQuery(params: ListUsersParams): string {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== "") query.set(key, String(value));
  }
  const qs = query.toString();
  return qs ? `?${qs}` : "";
}

export function listUsers(params: ListUsersParams = {}) {
  return apiFetch<{ users: UserRecord[]; total: number; page: number; pageSize: number }>(
    `/api/users${buildQuery(params)}`,
  );
}

export function getUser(id: number) {
  return apiFetch<{ user: UserRecord }>(`/api/users/${id}`);
}

export interface CreateUserPayload {
  email: string;
  password: string;
  fullName: string;
  role: "admin" | "agent";
  isActive?: boolean;
}

export function createUser(payload: CreateUserPayload) {
  return apiFetch<{ user: UserRecord }>("/api/users", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export interface UpdateUserPayload {
  email?: string;
  fullName?: string;
  password?: string;
  role?: "admin" | "agent";
}

export function updateUser(id: number, payload: UpdateUserPayload) {
  return apiFetch<{ user: UserRecord }>(`/api/users/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export function setUserStatus(id: number, isActive: boolean) {
  return apiFetch<{ user: UserRecord }>(`/api/users/${id}/status`, {
    method: "PATCH",
    body: JSON.stringify({ isActive }),
  });
}

export function deleteUser(id: number) {
  return apiFetch<{ success: boolean }>(`/api/users/${id}`, { method: "DELETE" });
}

export function transferOwnership(newOwnerId: number) {
  return apiFetch<{ success: boolean }>(`/api/users/${newOwnerId}/owner`, { method: "PUT" });
}

export function bulkSetStatus(ids: number[], isActive: boolean) {
  return apiFetch<{ succeeded: number[]; failed: BulkFailure[] }>("/api/users/bulk-status", {
    method: "POST",
    body: JSON.stringify({ ids, isActive }),
  });
}

export function bulkDeleteUsers(ids: number[]) {
  return apiFetch<{ succeeded: number[]; failed: BulkFailure[] }>("/api/users/bulk-delete", {
    method: "POST",
    body: JSON.stringify({ ids }),
  });
}
