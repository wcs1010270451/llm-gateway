import { Button, Space, Table, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";

import { StatusTag } from "../../components/StatusTag";
import type { User } from "../../types";

interface UserTableProps {
  data: User[];
  loading: boolean;
  onEdit: (user: User) => void;
  onDelete: (user: User) => void;
}

export function UserTable({ data, loading, onEdit, onDelete }: UserTableProps) {
  const columns: ColumnsType<User> = [
    {
      title: "用户",
      dataIndex: "email",
      render: (value, record) => (
        <div>
          <Typography.Text strong>{value}</Typography.Text>
          <div className="table-subtitle">{record.display_name || "-"}</div>
        </div>
      ),
    },
    { title: "角色", dataIndex: "role", width: 120 },
    {
      title: "状态",
      dataIndex: "status",
      width: 100,
      render: (value) => <StatusTag value={value} />,
    },
    {
      title: "最近登录",
      dataIndex: "last_login_at",
      width: 180,
      render: (value) => value || "-",
    },
    {
      title: "操作",
      key: "actions",
      width: 128,
      render: (_, record) => (
        <Space size={4}>
          <Button size="small" onClick={() => onEdit(record)}>
            编辑
          </Button>
          <Button size="small" danger onClick={() => onDelete(record)}>
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return <Table className="admin-table" rowKey="id" size="middle" columns={columns} dataSource={data} loading={loading} pagination={false} scroll={{ x: "max-content" }} />;
}
