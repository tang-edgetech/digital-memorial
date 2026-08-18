import { apiFetch } from "./client";

export type SiteSettings = Record<string, string>;

export function getSettings() {
  return apiFetch<{ settings: SiteSettings }>("/api/settings");
}

export function updateSettings(partial: Record<string, string>) {
  return apiFetch<{ settings: SiteSettings }>("/api/settings", {
    method: "PUT",
    body: JSON.stringify({ settings: partial }),
  });
}
