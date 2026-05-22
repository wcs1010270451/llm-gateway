import { Button, Space, Table, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";

import { StatusTag } from "../../components/StatusTag";
import type { APIKey } from "../../types";

interface APIKeyTableProps {
  data: APIKey[];
  loading: boolean;
  onDetail: (apiKey: APIKey) => void;
  onEdit: (apiKey: APIKey) => void;
  onDelete: (apiKey: APIKey) => void;
}

export function APIKeyTable({ data, loading, onDetail, onEdit, onDelete }: APIKeyTableProps) {
  const columns: ColumnsType<APIKey> = [
    {
      title: "名称",
      dataIndex: "name",
      render: (value, record) => (
        <div>
          <Typography.Text strong>{value}</Typography.Text>
          <div className="table-subtitle">{record.masked_key}</div>
        </div>
      ),
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 100,
      render: (value) => <StatusTag value={value} />,
    },
    { title: "RPM", dataIndex: "rpm_limit", width: 90, render: (value) => value || "不限" },
    { title: "日请求", dataIndex: "daily_request_limit", width: 100, render: (value) => value || "不限" },
    { title: "日 Token", dataIndex: "daily_token_limit", width: 100, render: (value) => value || "不限" },
    { title: "最近使用", dataIndex: "last_used_at", width: 180, render: (value) => value || "-" },
    {
      title: "操作",
      key: "actions",
      width: 176,
      render: (_, record) => (
        <Space size={4}>
          <Button size="small" onClick={() => onDetail(record)}>
            详情
          </Button>
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

  return <Table rowKey="id" size="middle" columns={columns} dataSource={data} loading={loading} pagination={false} />;
}
