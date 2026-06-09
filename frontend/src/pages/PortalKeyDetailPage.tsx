import { ArrowLeftOutlined, SendOutlined } from "@ant-design/icons";
import { useQuery } from "@tanstack/react-query";
import { Button, Card, Checkbox, Descriptions, Input, Modal, Select, Space, Table, Tag, Typography, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useMemo, useRef, useState } from "react";
import { useLocation, useNavigate, useParams } from "react-router";

import {
  fetchMyAPIKey,
  fetchMyAPIKeyLogs,
  fetchMyAPIKeyModelStats,
  fetchMyModels,
  sendMyDebugChatCompletions,
  sendMyDebugGeminiGenerateContent,
  sendMyDebugMessages,
  type DebugMessagesResult,
} from "../api/apiKeys";
import { PageHeader } from "../components/PageHeader";
import { StatusTag } from "../components/StatusTag";
import type { APIKey, KeyModelUsageStat, Model, RequestLog } from "../types";

type DebugProtocol = "anthropic" | "openai_compatible" | "gemini";

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

function adapterTypeOf(model: Model) {
  return model.active_provider_model?.provider?.adapter_type;
}

function debugProtocolMatches(model: Model, protocol: DebugProtocol) {
  const adapterType = adapterTypeOf(model);
  if (protocol === "gemini") {
    return adapterType === "gemini" || adapterType === "vertexai";
  }
  return adapterType === protocol;
}

export function PortalKeyDetailPage() {
  const params = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const keyID = Number(params.id);
  const initialAPIKey = (location.state as { apiKey?: APIKey } | null)?.apiKey;
  const [logPage, setLogPage] = useState({ page: 1, pageSize: 20 });
  const [debugOpen, setDebugOpen] = useState(false);
  const [debugProtocol, setDebugProtocol] = useState<DebugProtocol>("anthropic");
  const [debugModel, setDebugModel] = useState("");
  const [debugMessage, setDebugMessage] = useState("");
  const [debugStream, setDebugStream] = useState(false);
  const [debugLoading, setDebugLoading] = useState(false);
  const [debugResult, setDebugResult] = useState<DebugMessagesResult | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const keyQuery = useQuery({
    queryKey: ["me", "api-keys", keyID],
    queryFn: () => fetchMyAPIKey(keyID),
    initialData: initialAPIKey,
    enabled: Number.isFinite(keyID) && keyID > 0,
  });
  const modelsQuery = useQuery({
    queryKey: ["me", "models"],
    queryFn: fetchMyModels,
    enabled: Number.isFinite(keyID) && keyID > 0,
  });
  const statsQuery = useQuery({
    queryKey: ["me", "api-keys", keyID, "model-stats"],
    queryFn: () => fetchMyAPIKeyModelStats(keyID),
    enabled: Number.isFinite(keyID) && keyID > 0,
  });
  const logsQuery = useQuery({
    queryKey: ["me", "api-keys", keyID, "logs", logPage],
    queryFn: () => fetchMyAPIKeyLogs(keyID, { page: logPage.page, page_size: logPage.pageSize }),
    enabled: Number.isFinite(keyID) && keyID > 0,
  });

  const modelOptions = useMemo(() => {
    const items = modelsQuery.data?.items ?? [];
    return items
      .filter((model) => debugProtocolMatches(model, debugProtocol))
      .map((model) => ({
        value: model.name,
        label: model.display_name ? `${model.display_name} (${model.name})` : model.name,
      }));
  }, [debugProtocol, modelsQuery.data?.items]);

  async function sendDebug() {
    const model = debugModel || modelOptions[0]?.value;
    if (!model || !debugMessage.trim()) {
      return;
    }
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    setDebugLoading(true);
    setDebugResult(null);
    try {
      const payload = {
        model,
        max_tokens: 1024,
        stream: debugStream,
        messages: [{ role: "user" as const, content: debugMessage.trim() }],
      };
      const result =
        debugProtocol === "gemini"
          ? await sendMyDebugGeminiGenerateContent(
              keyID,
              {
                model,
                contents: [{ role: "user", parts: [{ text: debugMessage.trim() }] }],
                generationConfig: { maxOutputTokens: 1024 },
              },
              debugStream,
              {
                signal: controller.signal,
                onUpdate: (next) => setDebugResult({ ...next }),
              },
            )
          : debugProtocol === "openai_compatible"
          ? await sendMyDebugChatCompletions(keyID, payload, {
              signal: controller.signal,
              onUpdate: (next) => setDebugResult({ ...next }),
            })
          : await sendMyDebugMessages(keyID, payload, {
              signal: controller.signal,
              onUpdate: (next) => setDebugResult({ ...next }),
            });
      setDebugResult(result);
    } catch (error) {
      if ((error as Error).name === "AbortError") {
        return;
      }
      const text = (error as Error).message || "调试请求失败";
      message.error(text);
      setDebugResult({ assistantText: "", rawText: "", error: text });
    } finally {
      setDebugLoading(false);
      abortRef.current = null;
    }
  }

  const statColumns: ColumnsType<KeyModelUsageStat> = [
    { title: "Cache Create", dataIndex: "cache_creation_input_tokens", width: 140, render: formatNumber },
    { title: "Cache Hit", dataIndex: "cache_read_input_tokens", width: 130, render: formatNumber },
    { title: "Reasoning", dataIndex: "reasoning_tokens", width: 130, render: formatNumber },
    { title: "Tool", dataIndex: "tool_tokens", width: 110, render: formatNumber },
    { title: "模型", dataIndex: "public_model_name", key: "public_model_name" },
    { title: "请求数", dataIndex: "request_count", width: 120, render: formatNumber },
    { title: "输入 Token", dataIndex: "prompt_tokens", width: 140, render: formatNumber },
    { title: "输出 Token", dataIndex: "completion_tokens", width: 140, render: formatNumber },
    { title: "总 Token", dataIndex: "total_tokens", width: 130, render: formatNumber },
    { title: "预估成本", dataIndex: "estimated_cost", width: 140, render: formatUSD },
  ];

  const logColumns: ColumnsType<RequestLog> = [
    { title: "时间", dataIndex: "created_at", width: 180, render: formatDateTime },
    { title: "模型", dataIndex: "public_model_name", width: 180, ellipsis: true, render: (value) => value || "-" },
    { title: "上游模型", dataIndex: "upstream_model", width: 180, ellipsis: true, render: (value) => value || "-" },
    { title: "类型", dataIndex: "request_type", width: 140 },
    {
      title: "状态",
      dataIndex: "success",
      width: 90,
      render: (value) => <Tag color={value ? "success" : "error"}>{value ? "成功" : "失败"}</Tag>,
    },
    { title: "HTTP", dataIndex: "http_status", width: 90 },
    { title: "Token", dataIndex: "total_tokens", width: 110, render: formatNumber },
    { title: "延迟", dataIndex: "latency_ms", width: 110, render: (value) => `${value ?? 0} ms` },
    { title: "错误", dataIndex: "error_message", width: 220, ellipsis: true, render: (value) => value || "-" },
  ];

  return (
    <div className="page-stack portal-key-detail">
      <PageHeader
        eyebrow="CREDENTIAL DETAIL"
        title={keyQuery.data?.name ?? "Key 详情"}
        description="查看凭据配置、模型用量和调用日志，需要时可以直接调试 API 响应。"
        actions={
          <>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate("/portal/keys")}>
              返回
            </Button>
            <Button type="primary" icon={<SendOutlined />} onClick={() => setDebugOpen(true)}>
              API 调试
            </Button>
          </>
        }
      />

      <Card className="portal-panel" title="凭据配置" loading={keyQuery.isLoading}>
        <Descriptions className="portal-descriptions" bordered column={{ xs: 1, sm: 2, lg: 3 }} size="small">
          <Descriptions.Item label="名称">{keyQuery.data?.name ?? "-"}</Descriptions.Item>
          <Descriptions.Item label="状态">{keyQuery.data ? <StatusTag value={keyQuery.data.status} /> : "-"}</Descriptions.Item>
          <Descriptions.Item label="脱敏 Key">{keyQuery.data?.masked_key ?? "-"}</Descriptions.Item>
          <Descriptions.Item label="RPM 限额">{keyQuery.data?.rpm_limit || "不限"}</Descriptions.Item>
          <Descriptions.Item label="日请求限额">{keyQuery.data?.daily_request_limit || "不限"}</Descriptions.Item>
          <Descriptions.Item label="日 Token 限额">{keyQuery.data?.daily_token_limit || "不限"}</Descriptions.Item>
          <Descriptions.Item label="最近使用">{formatDateTime(keyQuery.data?.last_used_at)}</Descriptions.Item>
          <Descriptions.Item label="过期时间">{keyQuery.data?.expires_at ? formatDateTime(keyQuery.data.expires_at) : "永不过期"}</Descriptions.Item>
          <Descriptions.Item label="创建时间">{formatDateTime(keyQuery.data?.created_at)}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card className="portal-panel" title="模型用量">
        <Table className="portal-table" rowKey="public_model_name" columns={statColumns} dataSource={statsQuery.data?.items ?? []} loading={statsQuery.isLoading} pagination={false} scroll={{ x: "max-content" }} />
      </Card>

      <Card className="portal-panel" title="调用日志" extra={<Typography.Text type="secondary">选择一行打开详情页</Typography.Text>}>
        <Table
          className="portal-table portal-log-table"
          rowKey="id"
          columns={logColumns}
          dataSource={logsQuery.data?.items ?? []}
          loading={logsQuery.isLoading}
          scroll={{ x: 1260 }}
          onRow={(record) => ({ onClick: () => navigate(`/portal/keys/${keyID}/logs/${record.id}`), style: { cursor: "pointer" } })}
          pagination={{
            current: logsQuery.data?.page ?? logPage.page,
            pageSize: logsQuery.data?.page_size ?? logPage.pageSize,
            total: logsQuery.data?.total ?? 0,
            showSizeChanger: true,
          }}
          onChange={(pagination) =>
            setLogPage({
              page: pagination.current ?? 1,
              pageSize: pagination.pageSize ?? 20,
            })
          }
        />
      </Card>

      <Modal
        className="portal-modal"
        open={debugOpen}
        title="API 调试"
        width={720}
        footer={null}
        destroyOnHidden
        onCancel={() => {
          abortRef.current?.abort();
          setDebugOpen(false);
          setDebugResult(null);
        }}
      >
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Space wrap>
            <Select
              value={debugProtocol}
              onChange={(value) => {
                setDebugProtocol(value);
                setDebugModel("");
                setDebugResult(null);
              }}
              style={{ width: 180 }}
              options={[
                { value: "anthropic", label: "Anthropic Messages" },
                { value: "openai_compatible", label: "OpenAI Chat" },
                { value: "gemini", label: "Gemini GenerateContent" },
              ]}
            />
            <Select value={debugModel || modelOptions[0]?.value} onChange={setDebugModel} style={{ width: 300 }} placeholder="选择模型" options={modelOptions} />
            <Checkbox checked={debugStream} onChange={(event) => setDebugStream(event.target.checked)}>
              流式输出
            </Checkbox>
          </Space>
          <Input.TextArea
            rows={5}
            value={debugMessage}
            onChange={(event) => setDebugMessage(event.target.value)}
            placeholder="输入测试消息。Enter 发送，Shift+Enter 换行。"
            onPressEnter={(event) => {
              if (!event.shiftKey) {
                event.preventDefault();
                void sendDebug();
              }
            }}
          />
          <Space>
            <Button type="primary" icon={<SendOutlined />} loading={debugLoading} disabled={!debugMessage.trim() || modelOptions.length === 0} onClick={() => void sendDebug()}>
              发送
            </Button>
            {debugLoading ? <Button onClick={() => abortRef.current?.abort()}>停止</Button> : null}
          </Space>
          {debugResult ? (
            <div>
              {debugResult.error ? (
                <Typography.Text type="danger">{debugResult.error}</Typography.Text>
              ) : (
                <>
                  <pre className="json-preview">{debugResult.assistantText || "等待响应..."}</pre>
                  {debugResult.usage ? (
                    <Space size={16} wrap>
                      <Typography.Text type="secondary">Cache Create: {formatNumber(debugResult.usage.cache_creation_input_tokens)}</Typography.Text>
                      <Typography.Text type="secondary">Cache Hit: {formatNumber(debugResult.usage.cache_read_input_tokens)}</Typography.Text>
                      <Typography.Text type="secondary">Reasoning: {formatNumber(debugResult.usage.reasoning_tokens)}</Typography.Text>
                      <Typography.Text type="secondary">Tool: {formatNumber(debugResult.usage.tool_tokens)}</Typography.Text>
                      <Typography.Text type="secondary">输入 Token: {formatNumber(debugResult.usage.prompt_tokens)}</Typography.Text>
                      <Typography.Text type="secondary">输出 Token: {formatNumber(debugResult.usage.completion_tokens)}</Typography.Text>
                      <Typography.Text type="secondary">总 Token: {formatNumber(debugResult.usage.total_tokens)}</Typography.Text>
                    </Space>
                  ) : null}
                </>
              )}
            </div>
          ) : null}
        </Space>
      </Modal>

    </div>
  );
}
