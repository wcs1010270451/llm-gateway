import { Button, Space, Table, Tooltip, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useNavigate } from "react-router";

import { StatusTag } from "../../components/StatusTag";
import type { Model } from "../../types";
import { formatPricingAmount, pricingJSONToFields } from "../../utils/pricing";

interface ModelTableProps {
  data: Model[];
  loading: boolean;
  page: number;
  pageSize: number;
  total: number;
  togglingModelId?: number;
  onDelete: (model: Model) => void;
  onToggleStatus: (model: Model) => void;
  onPageChange: (page: number) => void;
}

export function ModelTable({
  data,
  loading,
  page,
  pageSize,
  total,
  togglingModelId,
  onDelete,
  onToggleStatus,
  onPageChange,
}: ModelTableProps) {
  const navigate = useNavigate();

  const columns: ColumnsType<Model> = [
    {
      title: "模型",
      dataIndex: "name",
      render: (value, record) => (
        <div>
          <Typography.Text strong>{value}</Typography.Text>
          <div className="table-subtitle">{record.display_name || record.family || "-"}</div>
        </div>
      ),
    },
    { title: "系列", dataIndex: "family", width: 120, render: (value) => value || "-" },
    { title: "能力", dataIndex: "modality", width: 120 },
    {
      title: "输入价格 / 1M",
      dataIndex: "pricing_json",
      key: "pricing_input",
      width: 130,
      render: (value) => {
        const pricing = pricingJSONToFields(value, "CNY");
        return formatPricingAmount(pricing.pricing_input, pricing.pricing_currency);
      },
    },
    {
      title: "输出价格 / 1M",
      dataIndex: "pricing_json",
      key: "pricing_output",
      width: 130,
      render: (value) => {
        const pricing = pricingJSONToFields(value, "CNY");
        return formatPricingAmount(pricing.pricing_output, pricing.pricing_currency);
      },
    },
    {
      title: "缓存价格 / 1M",
      dataIndex: "pricing_json",
      key: "pricing_cache",
      width: 130,
      render: (value) => {
        const pricing = pricingJSONToFields(value, "CNY");
        return formatPricingAmount(pricing.pricing_cache, pricing.pricing_currency);
      },
    },
    {
      title: "当前供应商",
      dataIndex: "active_provider_model",
      render: (_, record) => record.active_provider_model?.provider?.name ?? "-",
    },
    {
      title: "上游模型",
      dataIndex: "active_provider_model",
      render: (_, record) => record.active_provider_model?.upstream_model ?? "-",
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
      width: 134,
      render: (_, record) => {
        const isEnabled = record.status === "enabled";
        const canEnable = Boolean(record.active_provider_model_id);
        const statusButton = (
          <Button
            size="small"
            danger={isEnabled}
            disabled={!isEnabled && !canEnable}
            loading={togglingModelId === record.id}
            onClick={(event) => {
              event.stopPropagation();
              onToggleStatus(record);
            }}
          >
            {isEnabled ? "禁用" : "启用"}
          </Button>
        );

        return (
          <Space size={4}>
            {!isEnabled && !canEnable ? (
              <Tooltip title="启用前需要先设置当前上游">{statusButton}</Tooltip>
            ) : (
              statusButton
            )}
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
        );
      },
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
      pagination={{
        current: page,
        pageSize,
        total,
        showSizeChanger: false,
      }}
      onChange={(pagination) => onPageChange(pagination.current ?? 1)}
      scroll={{ x: "max-content" }}
      onRow={(record) => ({
        className: "admin-table-row-action",
        tabIndex: 0,
        onClick: () => navigate(`/models/${record.id}`),
        onKeyDown: (event) => {
          if (event.key === "Enter" || event.key === " ") {
            event.preventDefault();
            navigate(`/models/${record.id}`);
          }
        },
      })}
    />
  );
}
