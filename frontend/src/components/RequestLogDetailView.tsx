import { Descriptions, Space, Tag, Typography } from "antd";

import type { RequestLog } from "../types";

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

interface RequestLogDetailViewProps {
  item: RequestLog;
  showUser?: boolean;
}

export function RequestLogDetailView({ item, showUser = false }: RequestLogDetailViewProps) {
  return (
    <Space direction="vertical" size={18} style={{ width: "100%" }}>
      <Descriptions className="admin-descriptions" bordered column={2} size="small">
        <Descriptions.Item label="Request ID" span={2}>
          {item.request_id || "-"}
        </Descriptions.Item>
        {showUser ? <Descriptions.Item label="User ID">{item.user_id ?? "-"}</Descriptions.Item> : null}
        {showUser ? <Descriptions.Item label="Key ID">{item.api_key_id ?? "-"}</Descriptions.Item> : null}
        <Descriptions.Item label="Model">{item.public_model_name || "-"}</Descriptions.Item>
        <Descriptions.Item label="Upstream Model">{item.upstream_model || "-"}</Descriptions.Item>
        <Descriptions.Item label="Provider ID">{item.provider_id ?? "-"}</Descriptions.Item>
        <Descriptions.Item label="Provider Model ID">{item.provider_model_id ?? "-"}</Descriptions.Item>
        <Descriptions.Item label="Adapter">{item.adapter_type || "-"}</Descriptions.Item>
        <Descriptions.Item label="Request Type">
          <span className="admin-log-nowrap">{item.request_type || "-"}</span>
        </Descriptions.Item>
        <Descriptions.Item label="Status">
          <Tag color={item.success ? "success" : "error"}>{item.success ? "Success" : "Failed"}</Tag>
        </Descriptions.Item>
        <Descriptions.Item label="HTTP">{item.http_status}</Descriptions.Item>
        <Descriptions.Item label="Latency">{item.latency_ms} ms</Descriptions.Item>
        <Descriptions.Item label="Total Token">{formatNumber(item.total_tokens)}</Descriptions.Item>
        <Descriptions.Item label="Input Token">{formatNumber(item.prompt_tokens)}</Descriptions.Item>
        <Descriptions.Item label="Output Token">{formatNumber(item.completion_tokens)}</Descriptions.Item>
        <Descriptions.Item label="Cache Create Token">{formatNumber(item.cache_creation_input_tokens)}</Descriptions.Item>
        <Descriptions.Item label="Cache Hit Token">{formatNumber(item.cache_read_input_tokens)}</Descriptions.Item>
        <Descriptions.Item label="Reasoning Token">{formatNumber(item.reasoning_tokens)}</Descriptions.Item>
        <Descriptions.Item label="Tool Token">{formatNumber(item.tool_tokens)}</Descriptions.Item>
        <Descriptions.Item label="Estimated Cost">{formatUSD(item.estimated_cost)}</Descriptions.Item>
        <Descriptions.Item label="Client IP">{item.client_ip || "-"}</Descriptions.Item>
        <Descriptions.Item label="Request Path" span={2}>
          <span className="admin-log-nowrap">{item.request_path || "-"}</span>
        </Descriptions.Item>
        <Descriptions.Item label="Error Type">{item.error_type || "-"}</Descriptions.Item>
        <Descriptions.Item label="Error Message" span={2}>
          {item.error_message || "-"}
        </Descriptions.Item>
      </Descriptions>
      <div>
        <Typography.Title level={5}>Request</Typography.Title>
        <pre className="json-preview request-log-body">{formatJSON(item.request_preview)}</pre>
      </div>
      <div>
        <Typography.Title level={5}>Response</Typography.Title>
        <pre className="json-preview request-log-body">{formatJSON(item.response_preview)}</pre>
      </div>
      <div>
        <Typography.Title level={5}>Metadata</Typography.Title>
        <pre className="json-preview request-log-body">{formatJSON(item.metadata)}</pre>
      </div>
    </Space>
  );
}
