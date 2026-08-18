"use client";

import { Button, Tooltip, type ButtonProps } from "antd";

interface IconButtonWithTooltipProps extends ButtonProps {
  label: string;
}

/** Icon-only button that shows its action label on hover — the standard
 * pattern for quick-action buttons on listing pages. */
export function IconButtonWithTooltip({ label, ...buttonProps }: IconButtonWithTooltipProps) {
  return (
    <Tooltip title={label}>
      <Button aria-label={label} {...buttonProps} />
    </Tooltip>
  );
}
