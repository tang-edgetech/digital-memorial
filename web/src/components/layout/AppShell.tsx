"use client";

import { useState, type ReactNode } from "react";
import { Layout } from "antd";
import { Sidebar } from "./Sidebar";
import { Topbar } from "./Topbar";

const { Content } = Layout;

export function AppShell({ children, userLabel }: { children: ReactNode; userLabel?: string }) {
  const [collapsed, setCollapsed] = useState(false);
  return (
    <Layout className="min-h-screen">
      <Sidebar collapsed={collapsed} />
      <Layout>
        <Topbar collapsed={collapsed} onToggle={() => setCollapsed((c) => !c)} userLabel={userLabel} />
        <Content className="p-6">{children}</Content>
      </Layout>
    </Layout>
  );
}
