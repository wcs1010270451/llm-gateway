import { Button, Space, Table, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";

import { StatusTag } from "../../components/StatusTag";
import type { ProviderModel } from "../../types";
import { formatPricingAmount, pricingJSONToFields, type PricingFields } from "../../utils/pricing";

interface ProviderModelRouteTableProps {
  data: ProviderModel[];
  loading: boolean;
  onEdit: (providerModel: ProviderModel) => void;
  onDelete: (providerModel: ProviderModel) => void;
  onSetActive: (providerModel: ProviderModel) => void;
}

function getPricing(record: ProviderModel): PricingFields {
  if (Object.keys(record.pricing_json ?? {}).length > 0) {
    return pricingJSONToFields(record.pricing_json, "USD");
  }

  return {
    pricing_currency: "USD",
    pricing_input: record.input_cost_per_1m,
    pricing_output: record.output_cost_per_1m,
    pricing_cache: 0,
  };
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
      title: "输入成本 / 1M",
      key: "input_cost",
      width: 130,
      render: (_, record) => {
        const pricing = getPricing(record);
        return formatPricingAmount(pricing.pricing_input, pricing.pricing_currency);
      },
    },
    {
      title: "输出成本 / 1M",
      key: "output_cost",
      width: 130,
      render: (_, record) => {
        const pricing = getPricing(record);
        return formatPricingAmount(pricing.pricing_output, pricing.pricing_currency);
      },
    },
    {
      title: "缓存成本 / 1M",
      key: "cache_cost",
      width: 130,
      render: (_, record) => {
        const pricing = getPricing(record);
        return formatPricingAmount(pricing.pricing_cache, pricing.pricing_currency);
      },
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

  return <Table className="admin-table" rowKey="id" size="middle" columns={columns} dataSource={data} loading={loading} pagination={false} scroll={{ x: "max-content" }} />;
}
