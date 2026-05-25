import { CopyOutlined } from "@ant-design/icons";
import { Button, Empty, message, Space, Table, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useState } from "react";

import { revealMyAPIKey } from "../../api/apiKeys";
import { StatusTag } from "../../components/StatusTag";
import type { APIKey } from "../../types";

interface APIKeyTableProps {
  data: APIKey[];
  loading: boolean;
  onOpen: (apiKey: APIKey) => void;
  onEdit: (apiKey: APIKey) => void;
  onDelete: (apiKey: APIKey) => void;
}

export function APIKeyTable({ data, loading, onOpen, onEdit, onDelete }: APIKeyTableProps) {
  const [copyingID, setCopyingID] = useState<number>();

  function formatDateTime(value?: string) {
    return value ? new Date(value).toLocaleString("zh-CN") : "-";
  }

  async function copyKey(apiKey: APIKey) {
    setCopyingID(apiKey.id);
    try {
      const result = await revealMyAPIKey(apiKey.id);
      await navigator.clipboard.writeText(result.plain_key);
      message.success("完整 Key 已复制");
    } catch (error) {
      const errorMessage =
        typeof error === "object" &&
        error !== null &&
        "response" in error &&
        typeof error.response === "object" &&
        error.response !== null &&
        "data" in error.response
          ? (error.response.data as { error?: { message?: string } })?.error?.message
          : undefined;
      message.error(
        errorMessage === "full key is unavailable; create a new key to enable copying"
          ? "该 Key 创建时未保存可恢复密文，请新建 Key 后复制"
          : errorMessage ?? "复制失败，请稍后重试",
      );
    } finally {
      setCopyingID(undefined);
    }
  }

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
    { title: "最近使用", dataIndex: "last_used_at", width: 180, render: formatDateTime },
    {
      title: "操作",
      key: "actions",
      width: 216,
      render: (_, record) => (
        <Space size={4}>
          <Button
            size="small"
            icon={<CopyOutlined />}
            loading={copyingID === record.id}
            onClick={(event) => {
              event.stopPropagation();
              void copyKey(record);
            }}
            aria-label={`复制 ${record.name} 的完整 Key`}
          >
            复制
          </Button>
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
            aria-label={`删除 ${record.name}`}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Table
      className="portal-table"
      rowKey="id"
      size="middle"
      columns={columns}
      dataSource={data}
      loading={loading}
      pagination={false}
      scroll={{ x: 820 }}
      onRow={(record) => ({
        className: "portal-table-row-action",
        tabIndex: 0,
        "aria-label": `查看 ${record.name} 详情`,
        onClick: () => onOpen(record),
        onKeyDown: (event) => {
          if (event.target !== event.currentTarget) {
            return;
          }

          if (event.key === "Enter" || event.key === " ") {
            event.preventDefault();
            onOpen(record);
          }
        },
      })}
      locale={{
        emptyText: (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="还没有凭据。创建第一把 Key 后即可开始调用模型。"
          />
        ),
      }}
    />
  );
}
