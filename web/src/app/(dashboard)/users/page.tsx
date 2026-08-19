"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { Card, Table, Input, Select, Button, Tag, Space, Typography } from "antd";
import type { TableProps } from "antd";
import type { ColumnsType } from "antd/es/table";
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  StopOutlined,
  CheckCircleOutlined,
  CrownOutlined,
} from "@ant-design/icons";
import {
  listUsers,
  setUserStatus,
  deleteUser,
  bulkSetStatus,
  bulkDeleteUsers,
  transferOwnership,
  type UserRecord,
} from "@/lib/api/users";
import { useCurrentUser } from "@/lib/hooks/useCurrentUser";
import { hasPermission } from "@/lib/hooks/usePermission";
import { useConfirm } from "@/components/feedback/ConfirmModal";
import { useToast } from "@/lib/hooks/useToast";
import { IconButtonWithTooltip } from "@/components/feedback/IconButtonWithTooltip";
import { ApiError } from "@/lib/api/client";

const { Title } = Typography;

export default function UsersPage() {
  const { user: viewer } = useCurrentUser();
  const confirmAction = useConfirm();
  const toast = useToast();

  const [data, setData] = useState<UserRecord[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [selectedIds, setSelectedIds] = useState<number[]>([]);

  const [role, setRole] = useState<string | undefined>();
  const [status, setStatus] = useState<string | undefined>();
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [sortBy, setSortBy] = useState<string | undefined>();
  const [sortDir, setSortDir] = useState<string | undefined>();

  const load = () => {
    setLoading(true);
    listUsers({ role, status, search, page, pageSize, sortBy, sortDir })
      .then(({ users, total }) => {
        setData(users);
        setTotal(total);
      })
      .catch(() => toast.error("Failed to load users"))
      .finally(() => setLoading(false));
  };

  // load() also runs from event handlers (bulk actions, delete) where a sync
  // setState is fine; here it re-shows the spinner on filter/page change.
  // eslint-disable-next-line react-hooks/exhaustive-deps, react-hooks/set-state-in-effect
  useEffect(load, [role, status, search, page, pageSize, sortBy, sortDir]);

  const canCreate = hasPermission(viewer, "users", "create");
  const canEdit = hasPermission(viewer, "users", "edit");
  const canDelete = hasPermission(viewer, "users", "delete");
  const canToggle = hasPermission(viewer, "users", "enable_disable");

  const canActOn = (target: UserRecord) => {
    if (!viewer || viewer.id === target.id) return false;
    if (target.role === "admin" && viewer.role !== "super_admin" && !viewer.isOwner) return false;
    return true;
  };

  const handleToggleStatus = (target: UserRecord) => {
    confirmAction({
      title: target.isActive ? "Disable this account?" : "Enable this account?",
      content: `${target.fullName} (${target.email})`,
      danger: target.isActive,
      onConfirm: async () => {
        try {
          await setUserStatus(target.id, !target.isActive);
          toast.success(target.isActive ? "Account disabled" : "Account enabled");
          load();
        } catch (err) {
          toast.error(err instanceof ApiError ? err.message : "Action failed");
        }
      },
    });
  };

  const handleDelete = (target: UserRecord) => {
    confirmAction({
      title: "Delete this account?",
      content: `${target.fullName} (${target.email}) — this cannot be undone.`,
      danger: true,
      onConfirm: async () => {
        try {
          await deleteUser(target.id);
          toast.success("Account deleted");
          load();
        } catch (err) {
          toast.error(err instanceof ApiError ? err.message : "Delete failed");
        }
      },
    });
  };

  const handleMakeOwner = (target: UserRecord) => {
    confirmAction({
      title: "Transfer ownership?",
      content: `${target.fullName} will become the Owner. This can be reversed later by the new Owner or a Super Admin.`,
      onConfirm: async () => {
        try {
          await transferOwnership(target.id);
          toast.success("Ownership transferred");
          load();
        } catch (err) {
          toast.error(err instanceof ApiError ? err.message : "Transfer failed");
        }
      },
    });
  };

  const handleBulk = (action: "enable" | "disable" | "delete") => {
    confirmAction({
      title: `${action.charAt(0).toUpperCase()}${action.slice(1)} ${selectedIds.length} selected account(s)?`,
      danger: action !== "enable",
      onConfirm: async () => {
        try {
          const result =
            action === "delete" ? await bulkDeleteUsers(selectedIds) : await bulkSetStatus(selectedIds, action === "enable");
          if (result.failed.length === 0) {
            toast.success(`${result.succeeded.length} updated`);
          } else {
            toast.warning(`${result.succeeded.length} updated, ${result.failed.length} skipped`);
          }
          setSelectedIds([]);
          load();
        } catch (err) {
          toast.error(err instanceof ApiError ? err.message : "Bulk action failed");
        }
      },
    });
  };

  const columns = useMemo<ColumnsType<UserRecord>>(
    () => [
      {
        title: "Name",
        dataIndex: "fullName",
        sorter: true,
        render: (_: string, record: UserRecord) => (
          <span>
            {record.fullName}{" "}
            {record.isOwner && (
              <Tag color="gold" icon={<CrownOutlined />}>
                Owner
              </Tag>
            )}
          </span>
        ),
      },
      { title: "Email", dataIndex: "email", sorter: true },
      {
        title: "Role",
        dataIndex: "role",
        sorter: true,
        render: (role: string) => <Tag>{role.replace("_", " ")}</Tag>,
      },
      {
        title: "Status",
        dataIndex: "isActive",
        render: (isActive: boolean) => <Tag color={isActive ? "green" : "red"}>{isActive ? "Active" : "Disabled"}</Tag>,
      },
      {
        title: "Actions",
        key: "actions",
        render: (_: unknown, record: UserRecord) => (
          <Space>
            {canEdit && canActOn(record) && (
              <Link href={`/users/${record.id}/edit`}>
                <IconButtonWithTooltip label="Edit" icon={<EditOutlined />} />
              </Link>
            )}
            {canToggle && canActOn(record) && (
              <IconButtonWithTooltip
                label={record.isActive ? "Disable" : "Enable"}
                icon={record.isActive ? <StopOutlined /> : <CheckCircleOutlined />}
                onClick={() => handleToggleStatus(record)}
              />
            )}
            {record.role === "admin" && !record.isOwner && (viewer?.role === "super_admin" || viewer?.isOwner) && (
              <IconButtonWithTooltip label="Make Owner" icon={<CrownOutlined />} onClick={() => handleMakeOwner(record)} />
            )}
            {canDelete && canActOn(record) && (
              <IconButtonWithTooltip label="Delete" danger icon={<DeleteOutlined />} onClick={() => handleDelete(record)} />
            )}
          </Space>
        ),
      },
    ],
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [viewer, canEdit, canToggle, canDelete],
  );

  const handleTableChange: NonNullable<TableProps<UserRecord>["onChange"]> = (pagination, _filters, sorter) => {
    setPage(pagination.current ?? 1);
    setPageSize(pagination.pageSize ?? 20);
    const s = Array.isArray(sorter) ? sorter[0] : sorter;
    setSortBy(typeof s?.field === "string" ? s.field : undefined);
    setSortDir(s?.order === "descend" ? "desc" : s?.order === "ascend" ? "asc" : undefined);
  };

  return (
    <Card>
      <div className="flex items-center justify-between mb-4">
        <Title level={3} className="!mb-0">
          Users
        </Title>
        {canCreate && (
          <Link href="/users/create">
            <Button type="primary" icon={<PlusOutlined />}>
              Create User
            </Button>
          </Link>
        )}
      </div>

      <Space className="mb-4" wrap>
        <Input.Search
          placeholder="Search name or email"
          onSearch={(v) => {
            setPage(1);
            setSearch(v);
          }}
          allowClear
          className="w-64"
        />
        <Select
          placeholder="Role"
          allowClear
          className="w-32"
          options={[
            { value: "admin", label: "Admin" },
            { value: "agent", label: "Agent" },
          ]}
          onChange={(v) => {
            setPage(1);
            setRole(v);
          }}
        />
        <Select
          placeholder="Status"
          allowClear
          className="w-32"
          options={[
            { value: "active", label: "Active" },
            { value: "disabled", label: "Disabled" },
          ]}
          onChange={(v) => {
            setPage(1);
            setStatus(v);
          }}
        />
      </Space>

      {selectedIds.length > 0 && (
        <Space className="mb-4">
          <span>{selectedIds.length} selected</span>
          {canToggle && <Button onClick={() => handleBulk("enable")}>Enable</Button>}
          {canToggle && (
            <Button danger onClick={() => handleBulk("disable")}>
              Disable
            </Button>
          )}
          {canDelete && (
            <Button danger onClick={() => handleBulk("delete")}>
              Delete
            </Button>
          )}
        </Space>
      )}

      <Table<UserRecord>
        rowKey="id"
        loading={loading}
        dataSource={data}
        columns={columns}
        pagination={{ current: page, pageSize, total }}
        onChange={handleTableChange}
        rowSelection={{
          selectedRowKeys: selectedIds,
          onChange: (keys) => setSelectedIds(keys as number[]),
        }}
      />
    </Card>
  );
}
