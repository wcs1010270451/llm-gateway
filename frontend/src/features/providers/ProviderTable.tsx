import { Button, Space, Table, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useNavigate } from "react-router";

import { StatusTag } from "../../components/StatusTag";
import type { Provider } from "../../types";

interface ProviderTableProps {
  data: Provider[];
  loading: boolean;
  onEdit: (provider: Provider) => void;
  onDelete: (provider: Provider) => void;
}

export function ProviderTable({ data, loading, onEdit, onDelete }: ProviderTableProps) {
  const navigate = useNavigate();

  const columns: ColumnsType<Provider> = [
    {
      title: "名称",
      dataIndex: "name",
      render: (value, record) => (
        <div>
          <Typography.Text strong>{value}</Typography.Text>
          <div className="table-subtitle">{record.slug}</div>
        </div>
      ),
    },
    { title: "厂商", dataIndex: "vendor", width: 120 },
    { title: "适配器", dataIndex: "adapter_type", width: 160 },
    { title: "鉴权", dataIndex: "auth_type", width: 120 },
    {
      title: "Base URL",
      dataIndex: "base_url",
      ellipsis: true,
      render: (value) => value || "-",
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 100,
      render: (value) => <StatusTag value={value} />,
    },
    {
      title: "操作",
      key: "actions",
      width: 128,
      render: (_, record) => (
        <Space size={4}>
          <Button
            size="small"
            onClick={(event) => {
              event.stopPropagation();
              onEdit(record);
            }}
          >
            编辑
          </Button>
          <Button
            size="small"
            danger
            onClick={(event) => {
              event.stopPropagation();
              onDelete(record);
            }}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Table
      className="admin-table"
      rowKey="id"
      size="middle"
      columns={columns}
      dataSource={data}
      loading={loading}
      pagination={false}
      scroll={{ x: "max-content" }}
      onRow={(record) => ({
        className: "admin-table-row-action",
        tabIndex: 0,
        onClick: () => navigate(`/providers/${record.id}`),
        onKeyDown: (event) => {
          if (event.key === "Enter" || event.key === " ") {
            event.preventDefault();
            navigate(`/providers/${record.id}`);
          }
        },
      })}
    />
  );
}
