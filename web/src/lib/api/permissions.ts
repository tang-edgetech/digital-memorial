import { apiFetch } from "./client";

export interface RolePermissionRow {
  id: number;
  role: "admin" | "agent";
  module: string;
  action: string;
  allowed: boolean;
}

export function getPermissions() {
  return apiFetch<{ permissions: RolePermissionRow[]; registry: Record<string, string[]> }>("/api/permissions");
}

export interface PermissionUpdate {
  role: "admin" | "agent";
  module: string;
  action: string;
  allowed: boolean;
}

export function updatePermissions(updates: PermissionUpdate[]) {
  return apiFetch<{ permissions: RolePermissionRow[] }>("/api/permissions", {
    method: "PUT",
    body: JSON.stringify({ updates }),
  });
}
