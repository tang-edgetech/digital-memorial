"use client";

import { useEffect, useState } from "react";
import { Card, Table, Checkbox, Button, Typography, Spin } from "antd";
import { getPermissions, updatePermissions, type PermissionUpdate } from "@/lib/api/permissions";
import { useCurrentUser } from "@/lib/hooks/useCurrentUser";
import { useConfirm } from "@/components/feedback/ConfirmModal";
import { useToast } from "@/lib/hooks/useToast";
import { ApiError } from "@/lib/api/client";

const { Title, Paragraph } = Typography;

type MatrixKey = `${string}:${string}:${string}`; // role:module:action

interface MatrixRow {
  module: string;
  action: string;
}

export default function PermissionsPage() {
  const { user: viewer } = useCurrentUser();
  const confirmAction = useConfirm();
  const toast = useToast();

  const [rows, setRows] = useState<MatrixRow[]>([]);
  const [values, setValues] = useState<Record<MatrixKey, boolean>>({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    getPermissions()
      .then(({ permissions, registry }) => {
        const nextRows: MatrixRow[] = [];
        for (const [module, actions] of Object.entries(registry)) {
          for (const action of actions) nextRows.push({ module, action });
        }
        setRows(nextRows);

        const nextValues: Record<MatrixKey, boolean> = {};
        for (const p of permissions) {
          nextValues[`${p.role}:${p.module}:${p.action}`] = p.allowed;
        }
        setValues(nextValues);
      })
      .catch((err) => toast.error(err instanceof ApiError ? err.message : "Failed to load permissions"))
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const toggle = (role: "admin" | "agent", module: string, action: string) => {
    const key: MatrixKey = `${role}:${module}:${action}`;
    setValues((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  const handleSave = () => {
    const selfWarning =
      viewer?.role === "admin"
        ? " You are an Admin — changes to the Admin row apply to your own account too, and could lock you out of parts of the system."
        : "";
    confirmAction({
      title: "Save permission changes?",
      content: `These take effect immediately for every logged-in user of the affected role, no re-login required.${selfWarning}`,
      danger: viewer?.role === "admin",
      onConfirm: async () => {
        setSaving(true);
        try {
          const updates: PermissionUpdate[] = [];
          for (const role of ["admin", "agent"] as const) {
            for (const { module, action } of rows) {
              updates.push({ role, module, action, allowed: Boolean(values[`${role}:${module}:${action}`]) });
            }
          }
          await updatePermissions(updates);
          toast.success("Permissions saved");
        } catch (err) {
          toast.error(err instanceof ApiError ? err.message : "Failed to save permissions");
        } finally {
          setSaving(false);
        }
      },
    });
  };

  if (loading) {
    return (
      <div className="flex justify-center py-12">
        <Spin />
      </div>
    );
  }

  const columns = [
    {
      title: "Module.Action",
      key: "moduleAction",
      render: (_: unknown, record: MatrixRow) => (
        <span>
          {record.module}.{record.action}
        </span>
      ),
    },
    {
      title: "Admin",
      key: "admin",
      render: (_: unknown, record: MatrixRow) => (
        <Checkbox
          checked={Boolean(values[`admin:${record.module}:${record.action}`])}
          onChange={() => toggle("admin", record.module, record.action)}
        />
      ),
    },
    {
      title: "Agent",
      key: "agent",
      render: (_: unknown, record: MatrixRow) => (
        <Checkbox
          checked={Boolean(values[`agent:${record.module}:${record.action}`])}
          onChange={() => toggle("agent", record.module, record.action)}
        />
      ),
    },
  ];

  return (
    <Card>
      <Title level={3}>Roles &amp; Permissions</Title>
      <Paragraph type="secondary">
        Super Admin always has full access and isn&apos;t shown here. Changes apply immediately to every logged-in
        user of that role.
      </Paragraph>
      <Table rowKey={(r) => `${r.module}.${r.action}`} dataSource={rows} columns={columns} pagination={false} className="mb-4" />
      <Button type="primary" onClick={handleSave} loading={saving}>
        Save Permissions
      </Button>
    </Card>
  );
}
