"use client";

import type { CurrentUser } from "@/lib/api/auth";

/** UI-convenience check only — the backend remains the real gate on every
 * request. Used to conditionally show/hide buttons and nav entries. */
export function hasPermission(user: CurrentUser | null, module: string, action: string): boolean {
  if (!user) return false;
  if (user.role === "super_admin") return true;
  return Boolean(user.permissions?.[module]?.[action]);
}
