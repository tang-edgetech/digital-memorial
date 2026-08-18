"use client";

import { Layout, Menu } from "antd";
import { DashboardOutlined, SettingOutlined } from "@ant-design/icons";
import Link from "next/link";
import { usePathname } from "next/navigation";

const { Sider } = Layout;

const items = [
  { key: "/", icon: <DashboardOutlined />, label: <Link href="/">Dashboard</Link> },
  { key: "/settings", icon: <SettingOutlined />, label: <Link href="/settings">Site Settings</Link> },
];

export function Sidebar({ collapsed }: { collapsed: boolean }) {
  const pathname = usePathname();
  return (
    <Sider collapsible collapsed={collapsed} trigger={null} theme="light">
      <div className="h-16 flex items-center justify-center font-semibold text-lg overflow-hidden whitespace-nowrap">
        {collapsed ? "DM" : "Digital Memorial"}
      </div>
      <Menu mode="inline" selectedKeys={[pathname]} items={items} />
    </Sider>
  );
}
