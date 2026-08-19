"use client";

import { Layout, Menu } from "antd";
import { DashboardOutlined, SettingOutlined, TeamOutlined, SafetyCertificateOutlined } from "@ant-design/icons";
import Link from "next/link";
import { usePathname } from "next/navigation";
import type { CurrentUser } from "@/lib/api/auth";
import { hasPermission } from "@/lib/hooks/usePermission";

const { Sider } = Layout;

function buildItems(user: CurrentUser | null) {
  const items = [{ key: "/", icon: <DashboardOutlined />, label: <Link href="/">Dashboard</Link> }];

  if (hasPermission(user, "users", "view")) {
    items.push({ key: "/users", icon: <TeamOutlined />, label: <Link href="/users">Users</Link> });
  }
  if (user && (user.isOwner || user.role === "super_admin")) {
    items.push({
      key: "/permissions",
      icon: <SafetyCertificateOutlined />,
      label: <Link href="/permissions">Roles &amp; Permissions</Link>,
    });
  }
  if (hasPermission(user, "settings", "edit")) {
    items.push({ key: "/settings", icon: <SettingOutlined />, label: <Link href="/settings">Site Settings</Link> });
  }

  return items;
}

export function Sidebar({ collapsed, user }: { collapsed: boolean; user: CurrentUser | null }) {
  const pathname = usePathname();
  return (
    <Sider collapsible collapsed={collapsed} trigger={null} theme="light">
      <div className="h-16 flex items-center justify-center font-semibold text-lg overflow-hidden whitespace-nowrap">
        {collapsed ? "DM" : "Digital Memorial"}
      </div>
      <Menu mode="inline" selectedKeys={[pathname]} items={buildItems(user)} />
    </Sider>
  );
}
