"use client";

import { App } from "antd";

/** Thin wrapper over AntD's App-scoped message API — the single call site
 * every mutating action's success/error feedback should go through. */
export function useToast() {
  const { message, notification } = App.useApp();
  return {
    success: (content: string) => message.success(content),
    error: (content: string) => message.error(content),
    info: (content: string) => message.info(content),
    warning: (content: string) => message.warning(content),
    notify: notification,
  };
}
