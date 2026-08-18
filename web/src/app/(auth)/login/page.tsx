"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Card, Form, Input, Button, Typography } from "antd";
import { login } from "@/lib/api/auth";
import { useToast } from "@/lib/hooks/useToast";
import { ApiError } from "@/lib/api/client";

const { Title } = Typography;

export default function LoginPage() {
  const router = useRouter();
  const toast = useToast();
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (values: { email: string; password: string }) => {
    setSubmitting(true);
    try {
      await login(values.email, values.password);
      router.replace("/");
    } catch (err) {
      if (err instanceof ApiError && err.code === "account_locked") {
        toast.error("Account is temporarily locked due to repeated failed attempts.");
      } else if (err instanceof ApiError && err.code === "account_disabled") {
        toast.error("This account has been disabled.");
      } else {
        toast.error("Invalid email or password");
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Card>
      <Title level={3} className="!text-center">
        Digital Memorial
      </Title>
      <Form layout="vertical" onFinish={handleSubmit}>
        <Form.Item label="Email" name="email" rules={[{ required: true, type: "email" }]}>
          <Input autoFocus />
        </Form.Item>
        <Form.Item label="Password" name="password" rules={[{ required: true }]}>
          <Input.Password />
        </Form.Item>
        <Button type="primary" htmlType="submit" loading={submitting} block>
          Log In
        </Button>
      </Form>
    </Card>
  );
}
