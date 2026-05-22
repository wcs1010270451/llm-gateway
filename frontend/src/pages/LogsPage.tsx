import { useQuery } from "@tanstack/react-query";
import { Button, Card, Descriptions, Drawer, Space, Table, Tag, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useState } from "react";

import { fetchLog, fetchLogs } from "../api/admin";
import { PageHeader } from "../components/PageHeader";
import type { RequestLog } from "../types";

function formatDateTime(value?: string) {
  return value ? new Date(value).toLocaleString("zh-CN") : "-";
}

function formatNumber(value?: number) {
  return new Intl.NumberFormat("zh-CN").format(value ?? 0);
}

function formatUSD(value?: number) {
  return new Intl.NumberFormat("zh-CN", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 6,
    maximumFractionDigits: 6,
  }).format(value ?? 0);
}

function formatJSON(value: unknown) {
  if (value === null || value === undefined || value === "") {
    return "{}";
  }
  if (typeof value === "string") {
    try {
      return JSON.stringify(JSON.parse(value), null, 2);
    } catch {
      return value;
    }
  }
  return JSON.stringify(value, null, 2);
}

export function LogsPage() {
  const [pageState, setPageState] = useState({ page: 1, pageSize: 20 });
  const [selectedLogID, setSelectedLogID] = useState<number | null>(null);
  const logsQuery = useQuery({
    queryKey: ["admin", "logs", pageState],
    queryFn: () => fetchLogs({ page: pageState.page, page_size: pageState.pageSize }),
  });
  const detailQuery = useQuery({
    queryKey: ["admin", "logs", selectedLogID],
    queryFn: () => fetchLog(selectedLogID ?? 0),
    enabled: selectedLogID !== null,
  });

  const columns: ColumnsType<RequestLog> = [
    { title: "时间", dataIndex: "created_at", width: 180, render: formatDateTime },
    { title: "用户 ID", dataIndex: "user_id", width: 90, render: (value) => value ?? "-" },
    { title: "Key ID", dataIndex: "api_key_id", width: 90, render: (value) => value ?? "-" },
    { title: "模型", dataIndex: "public_model_name", width: 180, render: (value) => value || "-" },
    { title: "供应商 ID", dataIndex: "provider_id", width: 100, render: (value) => value ?? "-" },
    { title: "上游模型", dataIndex: "upstream_model", width: 180, render: (value) => value || "-" },
    { title: "适配器", dataIndex: "adapter_type", width: 120, render: (value) => value || "-" },
    { title: "类型", dataIndex: "request_type", width: 120 },
    {
      title: "状态",
      dataIndex: "success",
      width: 90,
      render: (value) => <Tag color={value ? "success" : "error"}>{value ? "成功" : "失败"}</Tag>,
    },
    { title: "HTTP", dataIndex: "http_status", width: 90 },
    { title: "Token", dataIndex: "total_tokens", width: 110, render: formatNumber },
    { title: "延迟", dataIndex: "latency_ms", width: 110, render: (value) => `${value ?? 0} ms` },
    { title: "错误", dataIndex: "error_message", width: 240, ellipsis: true, render: (value) => value || "-" },
    {
      title: "操作",
      key: "actions",
      fixed: "right",
      width: 90,
      render: (_, record) => (
        <Button size="small" onClick={() => setSelectedLogID(record.id)}>
          详情
        </Button>
      ),
    },
  ];

  return (
    <div className="page-stack">
      <PageHeader title="请求日志" description="查看所有用户 Key 通过网关发起的请求明细、耗时、Token 和错误信息。" />
      <Card>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={logsQuery.data?.items ?? []}
          loading={logsQuery.isLoading}
          scroll={{ x: 1680 }}
          onRow={(record) => ({ onDoubleClick: () => setSelectedLogID(record.id) })}
          pagination={{
            current: logsQuery.data?.page ?? pageState.page,
            pageSize: logsQuery.data?.page_size ?? pageState.pageSize,
            total: logsQuery.data?.total ?? 0,
            showSizeChanger: true,
          }}
          onChange={(pagination) =>
            setPageState({
              page: pagination.current ?? 1,
              pageSize: pagination.pageSize ?? 20,
            })
          }
        />
      </Card>

      <Drawer
        open={selectedLogID !== null}
        title="日志详情"
        width={780}
        destroyOnHidden
        onClose={() => setSelectedLogID(null)}
      >
        {detailQuery.isLoading || !detailQuery.data ? (
          <Typography.Text type="secondary">正在加载...</Typography.Text>
        ) : (
          <Space direction="vertical" size={18} style={{ width: "100%" }}>
            <Descriptions bordered column={2} size="small">
              <Descriptions.Item label="Request ID" span={2}>
                {detailQuery.data.request_id || "-"}
              </Descriptions.Item>
              <Descriptions.Item label="用户 ID">{detailQuery.data.user_id ?? "-"}</Descriptions.Item>
              <Descriptions.Item label="Key ID">{detailQuery.data.api_key_id ?? "-"}</Descriptions.Item>
              <Descriptions.Item label="模型">{detailQuery.data.public_model_name || "-"}</Descriptions.Item>
              <Descriptions.Item label="上游模型">{detailQuery.data.upstream_model || "-"}</Descriptions.Item>
              <Descriptions.Item label="供应商 ID">{detailQuery.data.provider_id ?? "-"}</Descriptions.Item>
              <Descriptions.Item label="Provider Model ID">{detailQuery.data.provider_model_id ?? "-"}</Descriptions.Item>
              <Descriptions.Item label="适配器">{detailQuery.data.adapter_type || "-"}</Descriptions.Item>
              <Descriptions.Item label="请求类型">{detailQuery.data.request_type || "-"}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={detailQuery.data.success ? "success" : "error"}>{detailQuery.data.success ? "成功" : "失败"}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="HTTP">{detailQuery.data.http_status}</Descriptions.Item>
              <Descriptions.Item label="延迟">{detailQuery.data.latency_ms} ms</Descriptions.Item>
              <Descriptions.Item label="Token">{formatNumber(detailQuery.data.total_tokens)}</Descriptions.Item>
              <Descriptions.Item label="输入 Token">{formatNumber(detailQuery.data.prompt_tokens)}</Descriptions.Item>
              <Descriptions.Item label="输出 Token">{formatNumber(detailQuery.data.completion_tokens)}</Descriptions.Item>
              <Descriptions.Item label="预估成本">{formatUSD(detailQuery.data.estimated_cost)}</Descriptions.Item>
              <Descriptions.Item label="客户端 IP">{detailQuery.data.client_ip || "-"}</Descriptions.Item>
              <Descriptions.Item label="请求路径">{detailQuery.data.request_path || "-"}</Descriptions.Item>
              <Descriptions.Item label="错误类型">{detailQuery.data.error_type || "-"}</Descriptions.Item>
              <Descriptions.Item label="错误信息" span={2}>
                {detailQuery.data.error_message || "-"}
              </Descriptions.Item>
            </Descriptions>
            <div>
              <Typography.Title level={5}>请求预览</Typography.Title>
              <pre className="json-preview">{formatJSON(detailQuery.data.request_preview)}</pre>
            </div>
            <div>
              <Typography.Title level={5}>响应预览</Typography.Title>
              <pre className="json-preview">{formatJSON(detailQuery.data.response_preview)}</pre>
            </div>
            <div>
              <Typography.Title level={5}>元数据</Typography.Title>
              <pre className="json-preview">{formatJSON(detailQuery.data.metadata)}</pre>
            </div>
          </Space>
        )}
      </Drawer>
    </div>
  );
}
