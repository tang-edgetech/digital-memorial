"use client";

import { useEffect, useState } from "react";
import { Modal } from "antd";
import { useRouter } from "next/navigation";
import { registerSessionExpiredHandler } from "@/lib/api/client";
import { logout } from "@/lib/api/auth";

/** Mounted once inside the dashboard layout. Triggered either by the client
 * idle timer or by apiFetch's global 401 handler (client.ts) — whichever
 * fires first — so the user is never left looking at a dead page after their
 * session lapses. */
export function SessionExpiredModal() {
  const [open, setOpen] = useState(false);
  const router = useRouter();

  useEffect(() => {
    registerSessionExpiredHandler(() => setOpen(true));
  }, []);

  const handleAcknowledge = async () => {
    setOpen(false);
    try {
      await logout();
    } catch {
      // session was already gone server-side; nothing left to clean up
    }
    router.replace("/login");
  };

  return (
    <Modal
      title="Session expired"
      open={open}
      onOk={handleAcknowledge}
      onCancel={handleAcknowledge}
      okText="Back to login"
      cancelButtonProps={{ style: { display: "none" } }}
      closable={false}
      maskClosable={false}
    >
      <p>You&apos;ve been logged out due to inactivity. Please log in again to continue.</p>
    </Modal>
  );
}
