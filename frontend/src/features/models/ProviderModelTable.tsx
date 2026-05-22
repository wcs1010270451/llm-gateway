import { Button, Space, Table, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";

import { StatusTag } from "../../components/StatusTag";
import type { ProviderModel } from "../../types";
import { formatPricing } from "../../utils/pricing";

interface ProviderModelTableProps {
  data: ProviderModel[];
  loading: boolean;
  activeProviderModelId?: number;
  onEdit: (providerModel: ProviderModel) => void;
  onDelete: (providerModel: ProviderModel) => void;
  onSetActive: (providerModel: ProviderModel) => void;
}

export function ProviderModelTable({
  data,
  loading,
  activeProviderModelId,
  onEdit,
  onDelete,
  onSetActive,
}: ProviderModelTableProps) {
  const columns: ColumnsType<ProviderModel> = [
    {
      title: "供应商",
      dataIndex: "provider",
      render: (_, record) => (
        <div>
          <Typography.Text strong>{record.provider?.name ?? "-"}</Typography.Text>
          <div className="table-subtitle">{record.provider?.adapter_type ?? "-"}</div>
        </div>
      ),
    },
    { title: "上游模型名", dataIndex: "upstream_model" },
    {
      title: "供应商成本 / 1M",
      key: "cost",
      width: 190,
      render: (_, record) =>
        Object.keys(record.pricing_json ?? {}).length > 0
          ? formatPricing(record.pricing_json, "USD")
          : `USD ${record.input_cost_per_1m} / ${record.output_cost_per_1m} / 0`,
    },
    { title: "超时", dataIndex: "timeout_seconds", width: 100, render: (value) => `${value}s` },
    {
      title: "状态",
      dataIndex: "status",
      width: 100,
      render: (value) => <StatusTag value={value} />,
    },
    {
      title: "当前上游",
      key: "active",
      width: 110,
      render: (_, record) => (record.id === activeProviderModelId ? <StatusTag value="enabled" /> : "-"),
    },
    {
      title: "操作",
      key: "actions",
      width: 220,
      render: (_, record) => (
        <Space size={4}>
          <Button
            size="small"
            disabled={record.status !== "enabled" || record.id === activeProviderModelId}
            onClick={() => onSetActive(record)}
          >
            设为当前
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

  return (
    <Table
      rowKey="id"
      size="middle"
      columns={columns}
      dataSource={data}
      loading={loading}
      pagination={false}
    />
  );
}
