"use client";

import { useState, type ReactNode } from "react";
import { Layout } from "antd";
import { Sidebar } from "./Sidebar";
import { Topbar } from "./Topbar";
import type { CurrentUser } from "@/lib/api/auth";

const { Content } = Layout;

export function AppShell({ children, user }: { children: ReactNode; user: CurrentUser | null }) {
  const [collapsed, setCollapsed] = useState(false);
  const userLabel = user ? `${user.fullName} (${user.role})` : undefined;
  return (
    <Layout className="min-h-screen">
      <Sidebar collapsed={collapsed} user={user} />
      <Layout>
        <Topbar collapsed={collapsed} onToggle={() => setCollapsed((c) => !c)} userLabel={userLabel} />
        <Content className="p-6">{children}</Content>
      </Layout>
    </Layout>
  );
}
