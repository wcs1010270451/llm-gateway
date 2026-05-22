import { Button, Space, Table, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";

import { StatusTag } from "../../components/StatusTag";
import type { ProviderModel } from "../../types";
import { formatPricing } from "../../utils/pricing";

interface ProviderModelRouteTableProps {
  data: ProviderModel[];
  loading: boolean;
  onEdit: (providerModel: ProviderModel) => void;
  onDelete: (providerModel: ProviderModel) => void;
  onSetActive: (providerModel: ProviderModel) => void;
}

export function ProviderModelRouteTable({
  data,
  loading,
  onEdit,
  onDelete,
  onSetActive,
}: ProviderModelRouteTableProps) {
  const columns: ColumnsType<ProviderModel> = [
    {
      title: "平台模型",
      dataIndex: "model",
      render: (_, record) => (
        <div>
          <Typography.Text strong>{record.model?.name ?? "未关联"}</Typography.Text>
          <div className="table-subtitle">{record.model?.display_name || record.model?.family || "-"}</div>
        </div>
      ),
    },
    { title: "上游模型名", dataIndex: "upstream_model" },
    {
      title: "供应商成本 / 1M",
      key: "cost",
      width: 220,
      render: (_, record) =>
        Object.keys(record.pricing_json ?? {}).length > 0
          ? formatPricing(record.pricing_json, "USD")
          : `USD 输入 ${record.input_cost_per_1m} / 输出 ${record.output_cost_per_1m} / 缓存 0`,
    },
    { title: "超时", dataIndex: "timeout_seconds", width: 90, render: (value) => `${value}s` },
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
      render: (_, record) => (record.model?.active_provider_model_id === record.id ? <StatusTag value="enabled" /> : "-"),
    },
    {
      title: "操作",
      key: "actions",
      width: 220,
      render: (_, record) => (
        <Space size={4}>
          <Button
            size="small"
            disabled={!record.model_id || record.status !== "enabled" || record.model?.active_provider_model_id === record.id}
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

  return <Table rowKey="id" size="middle" columns={columns} dataSource={data} loading={loading} pagination={false} />;
}
