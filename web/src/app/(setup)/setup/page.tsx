"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Card, Steps, Form, Input, Button, InputNumber, Typography, Spin } from "antd";
import {
  getSetupStatus,
  submitDbConfig,
  createSuperAdmin,
  submitInitialSettings,
} from "@/lib/api/setup";
import { useToast } from "@/lib/hooks/useToast";
import { ApiError } from "@/lib/api/client";

const { Title, Paragraph } = Typography;

export default function SetupPage() {
  const router = useRouter();
  const toast = useToast();
  const [checkingStatus, setCheckingStatus] = useState(true);
  const [step, setStep] = useState(0);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    getSetupStatus()
      .then(({ setupCompleted }) => {
        if (setupCompleted) {
          router.replace("/login");
        }
      })
      .finally(() => setCheckingStatus(false));
  }, [router]);

  if (checkingStatus) {
    return (
      <Card>
        <div className="flex justify-center py-8">
          <Spin />
        </div>
      </Card>
    );
  }

  const handleDbSubmit = async (values: {
    host: string;
    port: number;
    user: string;
    password: string;
    dbName: string;
  }) => {
    setSubmitting(true);
    try {
      await submitDbConfig({
        host: values.host,
        port: String(values.port),
        user: values.user,
        password: values.password,
        dbName: values.dbName,
      });
      toast.success("Database connected and schema initialized");
      setStep(1);
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to connect to database");
    } finally {
      setSubmitting(false);
    }
  };

  const handleAdminSubmit = async (values: { email: string; password: string; fullName: string }) => {
    setSubmitting(true);
    try {
      await createSuperAdmin(values);
      toast.success("Super Admin account created");
      setStep(2);
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to create Super Admin");
    } finally {
      setSubmitting(false);
    }
  };

  const handleSettingsSubmit = async (values: { siteTitle: string; logoPath?: string }) => {
    setSubmitting(true);
    try {
      await submitInitialSettings(values);
      toast.success("Setup complete — you can now log in");
      router.replace("/login");
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to save site settings");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Card>
      <Title level={3}>Digital Memorial — First-Run Setup</Title>
      <Paragraph type="secondary">
        Complete these steps once to connect your database, create the Super Admin account, and configure basic
        site settings.
      </Paragraph>

      <Steps
        current={step}
        className="!mb-6"
        items={[{ title: "Database" }, { title: "Super Admin" }, { title: "Site Settings" }]}
      />

      {step === 0 && (
        <Form layout="vertical" onFinish={handleDbSubmit} initialValues={{ port: 3306 }}>
          <Form.Item label="DB Host" name="host" rules={[{ required: true }]}>
            <Input placeholder="localhost" />
          </Form.Item>
          <Form.Item label="DB Port" name="port" rules={[{ required: true }]}>
            <InputNumber className="w-full" />
          </Form.Item>
          <Form.Item label="DB User" name="user" rules={[{ required: true }]}>
            <Input placeholder="root" />
          </Form.Item>
          <Form.Item label="DB Password" name="password">
            <Input.Password />
          </Form.Item>
          <Form.Item label="Database Name" name="dbName" rules={[{ required: true }]}>
            <Input placeholder="digital_memorial" />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={submitting} block>
            Connect &amp; Initialize
          </Button>
        </Form>
      )}

      {step === 1 && (
        <Form layout="vertical" onFinish={handleAdminSubmit}>
          <Form.Item label="Full Name" name="fullName" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item label="Email" name="email" rules={[{ required: true, type: "email" }]}>
            <Input />
          </Form.Item>
          <Form.Item label="Password" name="password" rules={[{ required: true, min: 8 }]}>
            <Input.Password />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={submitting} block>
            Create Super Admin
          </Button>
        </Form>
      )}

      {step === 2 && (
        <Form layout="vertical" onFinish={handleSettingsSubmit} initialValues={{ siteTitle: "Digital Memorial" }}>
          <Form.Item label="Site Title" name="siteTitle" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item
            label="Logo Path / URL"
            name="logoPath"
            extra="Optional for now — file upload can be added later; point this at an already-hosted image if you have one."
          >
            <Input placeholder="/logo.png" />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={submitting} block>
            Finish Setup
          </Button>
        </Form>
      )}
    </Card>
  );
}
