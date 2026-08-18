"use client";

import { App as AntApp } from "antd";
import type { ReactNode } from "react";

/** Wraps AntD's <App> so message/notification/modal are cssinjs-aware and
 * theme-token-aware everywhere in the tree below it. */
export function ToastProvider({ children }: { children: ReactNode }) {
  return <AntApp>{children}</AntApp>;
}
