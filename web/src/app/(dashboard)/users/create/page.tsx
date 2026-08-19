"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Card, Form, Input, Select, Switch, Button, Typography } from "antd";
import { createUser } from "@/lib/api/users";
import { useConfirm } from "@/components/feedback/ConfirmModal";
import { useToast } from "@/lib/hooks/useToast";
import { ApiError } from "@/lib/api/client";

const { Title } = Typography;

interface FormValues {
  email: string;
  password: string;
  fullName: string;
  role: "admin" | "agent";
  isActive: boolean;
}

export default function CreateUserPage() {
  const router = useRouter();
  const confirmAction = useConfirm();
  const toast = useToast();
  const [submitting, setSubmitting] = useState(false);

  const handleFinish = (values: FormValues) => {
    confirmAction({
      title: "Create this user?",
      content: `${values.fullName} (${values.email}) as ${values.role}.`,
      onConfirm: async () => {
        setSubmitting(true);
        try {
          await createUser(values);
          toast.success("User created");
          router.push("/users");
        } catch (err) {
          toast.error(err instanceof ApiError ? err.message : "Failed to create user");
        } finally {
          setSubmitting(false);
        }
      },
    });
  };

  return (
    <Card className="max-w-xl">
      <Title level={3}>Create User</Title>
      <Form<FormValues> layout="vertical" onFinish={handleFinish} initialValues={{ role: "agent", isActive: true }}>
        <Form.Item label="Full Name" name="fullName" rules={[{ required: true }]}>
          <Input />
        </Form.Item>
        <Form.Item label="Email" name="email" rules={[{ required: true, type: "email" }]}>
          <Input />
        </Form.Item>
        <Form.Item label="Password" name="password" rules={[{ required: true, min: 8 }]}>
          <Input.Password />
        </Form.Item>
        <Form.Item label="Role" name="role" rules={[{ required: true }]}>
          <Select
            options={[
              { value: "admin", label: "Admin" },
              { value: "agent", label: "Agent" },
            ]}
          />
        </Form.Item>
        <Form.Item label="Active" name="isActive" valuePropName="checked">
          <Switch />
        </Form.Item>
        <Button type="primary" htmlType="submit" loading={submitting}>
          Create User
        </Button>
      </Form>
    </Card>
  );
}
