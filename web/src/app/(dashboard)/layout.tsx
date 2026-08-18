"use client";

import { useEffect, useState, type ReactNode } from "react";
import { useRouter } from "next/navigation";
import { Spin } from "antd";
import { ThemeProvider } from "@/components/theme/ThemeProvider";
import { AppShell } from "@/components/layout/AppShell";
import { SessionExpiredModal } from "@/components/feedback/SessionExpiredModal";
import { useIdleTimer } from "@/lib/hooks/useIdleTimer";
import { useCurrentUser } from "@/lib/hooks/useCurrentUser";
import { getSettings } from "@/lib/api/settings";
import { triggerSessionExpired } from "@/lib/api/client";

const DEFAULT_SESSION_TIMEOUT_MINUTES = 120;

export default function DashboardLayout({ children }: { children: ReactNode }) {
  const router = useRouter();
  const { user, loading } = useCurrentUser();
  const [sessionTimeoutMinutes, setSessionTimeoutMinutes] = useState(DEFAULT_SESSION_TIMEOUT_MINUTES);

  useEffect(() => {
    if (!loading && !user) {
      router.replace("/login");
    }
  }, [loading, user, router]);

  useEffect(() => {
    getSettings()
      .then(({ settings }) => {
        const parsed = parseInt(settings.session_timeout_minutes ?? "", 10);
        if (!Number.isNaN(parsed) && parsed > 0) setSessionTimeoutMinutes(parsed);
      })
      .catch(() => {
        // keep the default if settings can't be read (e.g. session already gone)
      });
  }, []);

  useIdleTimer(sessionTimeoutMinutes, () => triggerSessionExpired());

  if (loading || !user) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Spin />
      </div>
    );
  }

  return (
    <ThemeProvider initialMode={user.themePreference}>
      <AppShell userLabel={`${user.fullName} (${user.role})`}>{children}</AppShell>
      <SessionExpiredModal />
    </ThemeProvider>
  );
}
