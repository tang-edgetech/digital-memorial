"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Card, Form, Input, Select, Button, Typography, Spin, Tag } from "antd";
import { CrownOutlined } from "@ant-design/icons";
import { getUser, updateUser, transferOwnership, type UserRecord } from "@/lib/api/users";
import { useCurrentUser } from "@/lib/hooks/useCurrentUser";
import { useConfirm } from "@/components/feedback/ConfirmModal";
import { useToast } from "@/lib/hooks/useToast";
import { ApiError } from "@/lib/api/client";

const { Title } = Typography;

interface FormValues {
  fullName: string;
  email: string;
  role: "admin" | "agent";
  password?: string;
}

export default function EditUserPage() {
  const params = useParams<{ id: string }>();
  const userId = Number(params.id);
  const router = useRouter();
  const { user: viewer } = useCurrentUser();
  const confirmAction = useConfirm();
  const toast = useToast();

  const [form] = Form.useForm<FormValues>();
  const [target, setTarget] = useState<UserRecord | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    getUser(userId)
      .then(({ user }) => {
        setTarget(user);
        form.setFieldsValue({ fullName: user.fullName, email: user.email, role: user.role as "admin" | "agent" });
      })
      .catch((err) => toast.error(err instanceof ApiError ? err.message : "Failed to load user"))
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [userId]);

  const handleFinish = (values: FormValues) => {
    const payload = { ...values, password: values.password || undefined };
    confirmAction({
      title: "Save changes to this user?",
      onConfirm: async () => {
        setSubmitting(true);
        try {
          await updateUser(userId, payload);
          toast.success("User updated");
          router.push("/users");
        } catch (err) {
          toast.error(err instanceof ApiError ? err.message : "Failed to update user");
        } finally {
          setSubmitting(false);
        }
      },
    });
  };

  const handleMakeOwner = () => {
    confirmAction({
      title: "Transfer ownership?",
      content: `${target?.fullName} will become the Owner.`,
      onConfirm: async () => {
        try {
          await transferOwnership(userId);
          toast.success("Ownership transferred");
          router.push("/users");
        } catch (err) {
          toast.error(err instanceof ApiError ? err.message : "Transfer failed");
        }
      },
    });
  };

  if (!loading && !target) {
    return (
      <Card>
        <Typography.Text type="danger">User not found.</Typography.Text>
      </Card>
    );
  }

  const canTransfer =
    target?.role === "admin" && !target?.isOwner && (viewer?.role === "super_admin" || viewer?.isOwner);

  return (
    <Card className="max-w-xl">
      <Title level={3}>
        Edit User {target?.isOwner && <Tag color="gold" icon={<CrownOutlined />}>Owner</Tag>}
      </Title>
      <Spin spinning={loading}>
        <Form<FormValues> form={form} layout="vertical" onFinish={handleFinish}>
          <Form.Item label="Full Name" name="fullName" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item label="Email" name="email" rules={[{ required: true, type: "email" }]}>
            <Input />
          </Form.Item>
          <Form.Item label="Role" name="role" rules={[{ required: true }]}>
            <Select
              options={[
                { value: "admin", label: "Admin" },
                { value: "agent", label: "Agent" },
              ]}
            />
          </Form.Item>
          <Form.Item label="New Password" name="password" extra="Leave blank to keep the current password.">
            <Input.Password />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={submitting}>
            Save Changes
          </Button>
        </Form>
      </Spin>

      {canTransfer && (
        <div className="mt-6 pt-6 border-t border-black/10">
          <Button icon={<CrownOutlined />} onClick={handleMakeOwner}>
            Make Owner
          </Button>
        </div>
      )}
    </Card>
  );
}
