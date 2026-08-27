"use client";

import { useEffect, useState } from "react";
import { Card, Form, Input, InputNumber, Switch, Button, Tabs, Typography, Spin, Select } from "antd";
import { getSettings, updateSettings } from "@/lib/api/settings";
import { useConfirm } from "@/components/feedback/ConfirmModal";
import { useToast } from "@/lib/hooks/useToast";
import { ApiError } from "@/lib/api/client";

const { Title } = Typography;

type FormValues = Record<string, string | number | boolean>;

function toInitialValues(settings: Record<string, string>): FormValues {
  const boolKeys = new Set(["maintenance_mode"]);
  const intKeys = new Set([
    "session_timeout_minutes",
    "password_min_length",
    "password_expiry_days",
    "lockout_threshold",
    "lockout_duration_minutes",
    "upload_max_size_mb",
    "pagination_default",
    "smtp_port",
    "audit_log_retention_days",
  ]);

  const result: FormValues = {};
  for (const [key, value] of Object.entries(settings)) {
    if (key === "setup_completed") continue;
    if (boolKeys.has(key)) {
      result[key] = value === "true";
    } else if (intKeys.has(key)) {
      const parsed = parseInt(value, 10);
      result[key] = Number.isNaN(parsed) ? 0 : parsed;
    } else {
      result[key] = value;
    }
  }
  return result;
}

function toApiPayload(values: FormValues): Record<string, string> {
  const payload: Record<string, string> = {};
  for (const [key, value] of Object.entries(values)) {
    if (value === undefined || value === null) continue;
    payload[key] = typeof value === "boolean" ? String(value) : String(value);
  }
  return payload;
}

export default function SiteSettingsPage() {
  const [form] = Form.useForm<FormValues>();
  const confirmAction = useConfirm();
  const toast = useToast();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    getSettings()
      .then(({ settings }) => form.setFieldsValue(toInitialValues(settings)))
      .catch(() => toast.error("Failed to load site settings"))
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleFinish = (values: FormValues) => {
    confirmAction({
      title: "Save site settings?",
      content: "These changes apply immediately across the system.",
      onConfirm: async () => {
        setSaving(true);
        try {
          await updateSettings(toApiPayload(values));
          toast.success("Site settings saved");
        } catch (err) {
          toast.error(err instanceof ApiError ? err.message : "Failed to save settings");
        } finally {
          setSaving(false);
        }
      },
    });
  };

  return (
    <Card>
      <Title level={3}>Site Settings</Title>
      <Spin spinning={loading}>
        <Form form={form} layout="vertical" onFinish={handleFinish} className="max-w-2xl">
        <Tabs
          items={[
            {
              key: "general",
              label: "General",
              children: (
                <>
                  <Form.Item label="Site Title" name="site_title" rules={[{ required: true }]}>
                    <Input />
                  </Form.Item>
                  <Form.Item label="Logo Path / URL" name="logo_path">
                    <Input placeholder="/logo.png" />
                  </Form.Item>
                  <Form.Item label="Favicon Path / URL" name="favicon_path">
                    <Input placeholder="/favicon.ico" />
                  </Form.Item>
                </>
              ),
            },
            {
              key: "org",
              label: "Organization",
              children: (
                <>
                  <Form.Item label="Organization Name" name="org_name">
                    <Input />
                  </Form.Item>
                  <Form.Item label="Address" name="org_address">
                    <Input.TextArea rows={2} />
                  </Form.Item>
                  <Form.Item label="Phone" name="org_phone">
                    <Input />
                  </Form.Item>
                  <Form.Item label="Email" name="org_email">
                    <Input />
                  </Form.Item>
                </>
              ),
            },
            {
              key: "security",
              label: "Auth & Security",
              children: (
                <>
                  <Form.Item
                    label="Login Route"
                    name="login_route"
                    extra="Displayed/editable now; /login always keeps working as a fallback in Phase 1 — full route replacement is a later enhancement."
                  >
                    <Input />
                  </Form.Item>
                  <Form.Item
                    label="Session Timeout (minutes)"
                    name="session_timeout_minutes"
                    extra="Users are logged out after this many minutes of inactivity. Default: 120 (2 hours)."
                  >
                    <InputNumber min={1} className="w-full" />
                  </Form.Item>
                  <Form.Item label="Password Minimum Length" name="password_min_length">
                    <InputNumber min={4} className="w-full" />
                  </Form.Item>
                  <Form.Item label="Password Complexity" name="password_complexity">
                    <Select
                      options={[
                        { value: "low", label: "Low" },
                        { value: "medium", label: "Medium" },
                        { value: "high", label: "High" },
                      ]}
                    />
                  </Form.Item>
                  <Form.Item label="Password Expiry (days, 0 = never)" name="password_expiry_days">
                    <InputNumber min={0} className="w-full" />
                  </Form.Item>
                  <Form.Item label="Lockout Threshold (failed attempts)" name="lockout_threshold">
                    <InputNumber min={1} className="w-full" />
                  </Form.Item>
                  <Form.Item label="Lockout Duration (minutes)" name="lockout_duration_minutes">
                    <InputNumber min={1} className="w-full" />
                  </Form.Item>
                </>
              ),
            },
            {
              key: "uploads",
              label: "Uploads",
              children: (
                <>
                  <Form.Item label="Max Upload Size (MB)" name="upload_max_size_mb">
                    <InputNumber min={1} className="w-full" />
                  </Form.Item>
                  <Form.Item label="Allowed File Types" name="upload_allowed_types">
                    <Input placeholder="image/png,image/jpeg" />
                  </Form.Item>
                </>
              ),
            },
            {
              key: "localization",
              label: "Localization",
              children: (
                <>
                  <Form.Item label="Default Timezone" name="default_timezone">
                    <Input />
                  </Form.Item>
                  <Form.Item label="Default Date Format" name="default_date_format">
                    <Input />
                  </Form.Item>
                  <Form.Item label="Default Theme (new users)" name="default_theme">
                    <Select
                      options={[
                        { value: "light", label: "Light" },
                        { value: "dark", label: "Dark" },
                      ]}
                    />
                  </Form.Item>
                  <Form.Item label="Pagination Default (rows per page)" name="pagination_default">
                    <InputNumber min={5} className="w-full" />
                  </Form.Item>
                </>
              ),
            },
            {
              key: "smtp",
              label: "Email (SMTP)",
              children: (
                <>
                  <Form.Item label="SMTP Host" name="smtp_host">
                    <Input />
                  </Form.Item>
                  <Form.Item label="SMTP Port" name="smtp_port">
                    <InputNumber min={1} className="w-full" />
                  </Form.Item>
                  <Form.Item label="SMTP User" name="smtp_user">
                    <Input />
                  </Form.Item>
                  <Form.Item label="SMTP Password" name="smtp_pass">
                    <Input.Password />
                  </Form.Item>
                  <Form.Item label="SMTP From Address" name="smtp_from">
                    <Input />
                  </Form.Item>
                </>
              ),
            },
            {
              key: "system",
              label: "System",
              children: (
                <>
                  <Form.Item label="Maintenance Mode" name="maintenance_mode" valuePropName="checked">
                    <Switch />
                  </Form.Item>
                  <Form.Item label="Audit Log Retention (days)" name="audit_log_retention_days">
                    <InputNumber min={1} className="w-full" />
                  </Form.Item>
                </>
              ),
            },
          ]}
        />
        <Button type="primary" htmlType="submit" loading={saving}>
          Save Settings
        </Button>
        </Form>
      </Spin>
    </Card>
  );
}
