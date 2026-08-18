import { apiFetch } from "./client";

export interface SetupStatus {
  setupCompleted: boolean;
}

export function getSetupStatus() {
  return apiFetch<SetupStatus>("/api/setup/status", { skipSessionExpiredHandling: true });
}

export interface DbConfigPayload {
  host: string;
  port: string;
  user: string;
  password: string;
  dbName: string;
}

export function submitDbConfig(payload: DbConfigPayload) {
  return apiFetch<{ success: boolean }>("/api/setup/db", {
    method: "POST",
    body: JSON.stringify(payload),
    skipSessionExpiredHandling: true,
  });
}

export interface SuperAdminPayload {
  email: string;
  password: string;
  fullName: string;
}

export function createSuperAdmin(payload: SuperAdminPayload) {
  return apiFetch<{ success: boolean }>("/api/setup/admin", {
    method: "POST",
    body: JSON.stringify(payload),
    skipSessionExpiredHandling: true,
  });
}

export interface InitialSettingsPayload {
  siteTitle: string;
  logoPath?: string;
}

export function submitInitialSettings(payload: InitialSettingsPayload) {
  return apiFetch<{ success: boolean }>("/api/setup/settings", {
    method: "POST",
    body: JSON.stringify(payload),
    skipSessionExpiredHandling: true,
  });
}
