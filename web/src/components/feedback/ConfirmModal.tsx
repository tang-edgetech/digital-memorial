"use client";

import { App } from "antd";
import { ExclamationCircleFilled } from "@ant-design/icons";
import type { ReactNode } from "react";
import { createElement } from "react";

interface ConfirmActionArgs {
  title: string;
  content?: ReactNode;
  danger?: boolean;
  okText?: string;
  cancelText?: string;
  onConfirm: () => void | Promise<void>;
}

/** Every state-changing action (create/update/delete/enable-disable/etc.)
 * should route through this before executing, per the "confirm before
 * execute" interaction rule. */
export function useConfirm() {
  const { modal } = App.useApp();

  return function confirmAction({
    title,
    content,
    danger,
    okText = "Confirm",
    cancelText = "Cancel",
    onConfirm,
  }: ConfirmActionArgs) {
    modal.confirm({
      title,
      icon: createElement(ExclamationCircleFilled),
      content,
      okText,
      cancelText,
      okButtonProps: { danger },
      onOk: onConfirm,
    });
  };
}
