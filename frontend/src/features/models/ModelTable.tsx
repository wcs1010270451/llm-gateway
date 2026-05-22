import { Button, Space, Table, Tooltip, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useNavigate } from "react-router";

import { StatusTag } from "../../components/StatusTag";
import type { Model } from "../../types";
import { formatPricing } from "../../utils/pricing";

interface ModelTableProps {
  data: Model[];
  loading: boolean;
  togglingModelId?: number;
  onEdit: (model: Model) => void;
  onDelete: (model: Model) => void;
  onToggleStatus: (model: Model) => void;
}

export function ModelTable({
  data,
  loading,
  togglingModelId,
  onEdit,
  onDelete,
  onToggleStatus,
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
      title: "平台价格 / 1M",
      dataIndex: "pricing_json",
      width: 190,
      render: (value) => formatPricing(value, "CNY"),
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
      width: 240,
      render: (_, record) => {
        const isEnabled = record.status === "enabled";
        const canEnable = Boolean(record.active_provider_model_id);
        const statusButton = (
          <Button
            size="small"
            danger={isEnabled}
            disabled={!isEnabled && !canEnable}
            loading={togglingModelId === record.id}
            onClick={() => onToggleStatus(record)}
          >
            {isEnabled ? "禁用" : "启用"}
          </Button>
        );

        return (
          <Space size={4}>
            <Button size="small" onClick={() => navigate(`/models/${record.id}`)}>
              映射
            </Button>
            <Button size="small" onClick={() => onEdit(record)}>
              编辑
            </Button>
            {!isEnabled && !canEnable ? (
              <Tooltip title="启用前需要先设置当前上游">{statusButton}</Tooltip>
            ) : (
              statusButton
            )}
            <Button size="small" danger onClick={() => onDelete(record)}>
              删除
            </Button>
          </Space>
        );
      },
    },
  ];

  return <Table rowKey="id" size="middle" columns={columns} dataSource={data} loading={loading} pagination={false} />;
}
