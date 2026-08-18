"use client";

import { Layout } from "antd";
import { MenuFoldOutlined, MenuUnfoldOutlined, LogoutOutlined } from "@ant-design/icons";
import { useRouter } from "next/navigation";
import { IconButtonWithTooltip } from "@/components/feedback/IconButtonWithTooltip";
import { useConfirm } from "@/components/feedback/ConfirmModal";
import { useToast } from "@/lib/hooks/useToast";
import { ThemeToggle } from "@/components/theme/ThemeToggle";
import { logout } from "@/lib/api/auth";

const { Header } = Layout;

interface TopbarProps {
  collapsed: boolean;
  onToggle: () => void;
  userLabel?: string;
}

export function Topbar({ collapsed, onToggle, userLabel }: TopbarProps) {
  const confirmAction = useConfirm();
  const toast = useToast();
  const router = useRouter();

  const handleLogout = () => {
    confirmAction({
      title: "Log out?",
      content: "You will need to log in again to continue.",
      onConfirm: async () => {
        try {
          await logout();
          toast.success("Logged out");
        } catch {
          // cookie may already be gone; proceed to redirect regardless
        }
        router.replace("/login");
      },
    });
  };

  return (
    <Header className="flex items-center justify-between px-4 bg-[var(--background)] border-b border-black/10">
      <IconButtonWithTooltip
        label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
        type="text"
        icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
        onClick={onToggle}
      />
      <div className="flex items-center gap-3">
        {userLabel && <span className="text-sm opacity-70">{userLabel}</span>}
        <ThemeToggle />
        <IconButtonWithTooltip label="Log out" type="text" icon={<LogoutOutlined />} onClick={handleLogout} />
      </div>
    </Header>
  );
}
