"use client";

import { useEffect, type ReactNode } from "react";
import { ConfigProvider, theme as antdTheme } from "antd";

/** Login always renders light, regardless of the logged-in user's saved
 * theme preference — it never mounts the app's ThemeProvider. If a previous
 * session left the `dark` class on <html>, strip it here so a same-tab
 * navigation to /login doesn't carry it over. */
export default function LoginLayout({ children }: { children: ReactNode }) {
  useEffect(() => {
    document.documentElement.classList.remove("dark");
  }, []);

  return (
    <ConfigProvider theme={{ algorithm: antdTheme.defaultAlgorithm }}>
      <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
        <div className="w-full max-w-sm">{children}</div>
      </div>
    </ConfigProvider>
  );
}
