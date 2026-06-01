import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Alert, Button, Card, Descriptions, Empty, Modal, Segmented, Space, Statistic, Table, Tag, Typography, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import {
  createProviderModelRoute,
  deleteProviderModelRoute,
  fetchClaudeProxyStatuses,
  fetchModels,
  fetchProvider,
  fetchProviderModelRoutes,
  fetchProviderUsageStats,
  probeClaudeProxy,
  refreshClaudeProxyToken,
  setActiveProviderModel,
  updateProviderModelRoute,
} from "../api/admin";
import { PageHeader } from "../components/PageHeader";
import { StatusTag } from "../components/StatusTag";
import {
  ProviderModelRouteDrawer,
  type ProviderModelRouteSubmit,
} from "../features/providers/ProviderModelRouteDrawer";
import { ProviderModelRouteTable } from "../features/providers/ProviderModelRouteTable";
import type { ProviderModel } from "../types";
import type { ProviderUsageStats } from "../api/admin";

function formatDateTime(value?: string) {
  return value ? new Date(value).toLocaleString() : "-";
}

function formatTokenHours(value?: number) {
  if (value === undefined) {
    return "-";
  }
  if (value < 0) {
    return "已过期";
  }
  return `${value.toFixed(1)} 小时`;
}

function estimateExpiresAt(checkedAt?: string, tokenHours?: number) {
  if (!checkedAt || tokenHours === undefined) {
    return "-";
  }
  return new Date(new Date(checkedAt).getTime() + tokenHours * 60 * 60 * 1000).toLocaleString();
}

function formatUSD(value?: number) {
  return `$${(value ?? 0).toFixed(6)}`;
}

function formatInteger(value?: number) {
  return (value ?? 0).toLocaleString();
}

function formatPercent(successCount?: number, requestCount?: number) {
  if (!requestCount) {
    return "-";
  }
  return `${(((successCount ?? 0) / requestCount) * 100).toFixed(1)}%`;
}

function formatChartPeriod(value: string, granularity?: "hour" | "day") {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  if (granularity === "day") {
    return `${date.getMonth() + 1}/${date.getDate()}`;
  }
  return `${date.getMonth() + 1}/${date.getDate()} ${String(date.getHours()).padStart(2, "0")}:00`;
}

export function ProviderDetailPage() {
  const { id = "" } = useParams();
  const providerId = Number(id);
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [editing, setEditing] = useState<ProviderModel | undefined>();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [usageWindow, setUsageWindow] = useState<"24h" | "7d" | "30d">("24h");

  const providerQuery = useQuery({
    queryKey: ["providers", id],
    queryFn: () => fetchProvider(id),
    enabled: id !== "",
  });
  const routesQuery = useQuery({
    queryKey: ["providers", id, "provider-models"],
    queryFn: () => fetchProviderModelRoutes(id),
    enabled: id !== "",
  });
  const usageGranularity = usageWindow === "24h" ? "hour" : "day";
  const usageQuery = useQuery({
    queryKey: ["providers", id, "usage-stats", usageWindow, usageGranularity],
    queryFn: () => fetchProviderUsageStats(id, { window: usageWindow, granularity: usageGranularity }),
    enabled: id !== "",
  });
  const modelsQuery = useQuery({ queryKey: ["models"], queryFn: fetchModels });
  const isClaudeMaxProxy = providerQuery.data?.slug === "claude_max_proxy";
  const claudeStatusQuery = useQuery({
    queryKey: ["claude-proxies"],
    queryFn: fetchClaudeProxyStatuses,
    enabled: Boolean(isClaudeMaxProxy),
  });
  const claudeStatus = claudeStatusQuery.data?.items.find((item) => item.provider_id === providerId);

  const saveMutation = useMutation({
    mutationFn: (payload: ProviderModelRouteSubmit) =>
      editing
        ? updateProviderModelRoute(providerId, editing.id, payload.input)
        : createProviderModelRoute(providerId, payload.input),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["models"] }),
        queryClient.invalidateQueries({ queryKey: ["providers", id, "provider-models"] }),
      ]);
      message.success("上游模型已保存");
      setDrawerOpen(false);
      setEditing(undefined);
    },
    onError: (error) => message.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: (providerModel: ProviderModel) => deleteProviderModelRoute(providerId, providerModel.id),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["models"] }),
        queryClient.invalidateQueries({ queryKey: ["providers", id, "provider-models"] }),
      ]);
      message.success("上游模型已删除");
    },
    onError: (error) => message.error(error.message),
  });

  const activateMutation = useMutation({
    mutationFn: (providerModel: ProviderModel) => setActiveProviderModel(providerModel.model_id ?? 0, providerModel.id),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["models"] }),
        queryClient.invalidateQueries({ queryKey: ["providers", id, "provider-models"] }),
      ]);
      message.success("当前上游已切换");
    },
    onError: (error) => message.error(error.message),
  });

  const probeMutation = useMutation({
    mutationFn: () => probeClaudeProxy(providerId),
    onSuccess: async (result) => {
      queryClient.setQueryData(["claude-proxies"], {
        items: [result],
        total: 1,
      });
      await claudeStatusQuery.refetch();
      if (result.probe_ok) {
        message.success("Claude 代理探测成功");
      } else {
        message.error(result.error || "Claude 代理探测失败");
      }
    },
    onError: (error) => message.error(error.message),
  });

  const refreshTokenMutation = useMutation({
    mutationFn: () => refreshClaudeProxyToken(providerId),
    onSuccess: async (result) => {
      queryClient.setQueryData(["claude-proxies"], {
        items: [result],
        total: 1,
      });
      await claudeStatusQuery.refetch();
      if (result.refresh_ok) {
        message.success("Claude Token 已刷新");
      } else {
        message.error(result.error || "Claude Token 刷新失败");
      }
    },
    onError: (error) => message.error(error.message),
  });

  function confirmDelete(providerModel: ProviderModel) {
    Modal.confirm({
      title: "删除上游模型",
      content: `确定删除 ${providerModel.model?.name ?? "未关联平台模型"} -> ${providerModel.upstream_model} 的映射吗？`,
      okText: "删除",
      okButtonProps: { danger: true },
      cancelText: "取消",
      onOk: () => deleteMutation.mutateAsync(providerModel),
    });
  }

  const provider = providerQuery.data;
  const usageStats = usageQuery.data;
  const trendData = useMemo(
    () =>
      (usageStats?.trend ?? []).map((item) => ({
        ...item,
        label: formatChartPeriod(item.period, usageStats?.granularity),
      })),
    [usageStats],
  );
  const modelCostData = useMemo(
    () =>
      (usageStats?.models ?? []).slice(0, 8).map((item) => ({
        ...item,
        label: item.upstream_model || item.public_model_name || "unknown",
      })),
    [usageStats],
  );
  const modelUsageColumns: ColumnsType<ProviderUsageStats["models"][number]> = [
    {
      title: "上游模型",
      dataIndex: "upstream_model",
      render: (value, record) => (
        <div>
          <Typography.Text strong>{value || "-"}</Typography.Text>
          <div className="table-subtitle">{record.public_model_name || "-"}</div>
        </div>
      ),
    },
    { title: "请求", dataIndex: "request_count", width: 100, render: formatInteger },
    {
      title: "成功率",
      dataIndex: "success_count",
      width: 100,
      render: (_, record) => formatPercent(record.success_count, record.request_count),
    },
    { title: "输入", dataIndex: "prompt_tokens", width: 120, render: formatInteger },
    { title: "输出", dataIndex: "completion_tokens", width: 120, render: formatInteger },
    { title: "总 Token", dataIndex: "total_tokens", width: 130, render: formatInteger },
    { title: "消耗", dataIndex: "estimated_cost", width: 120, render: formatUSD },
    {
      title: "平均延迟",
      dataIndex: "average_latency_ms",
      width: 120,
      render: (value) => `${Math.round(value ?? 0)} ms`,
    },
    { title: "最近调用", dataIndex: "last_used_at", width: 170, render: formatDateTime },
  ];

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="UPSTREAM CONNECTION"
        title={provider ? provider.name : `#${id}`}
        description="查看该供应商提供的上游模型，并维护它们与平台模型的映射。"
        actions={
          <>
            <Button onClick={() => navigate("/providers")}>返回</Button>
            <Button
              icon={<ReloadOutlined />}
              aria-label="刷新供应商详情"
              onClick={() => {
                providerQuery.refetch();
                routesQuery.refetch();
                usageQuery.refetch();
                modelsQuery.refetch();
                if (isClaudeMaxProxy) {
                  claudeStatusQuery.refetch();
                }
              }}
            />
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setEditing(undefined);
                setDrawerOpen(true);
              }}
            >
              新增上游模型
            </Button>
          </>
        }
      />

      {provider ? (
        <Card className="admin-panel">
          <Descriptions className="admin-descriptions" size="small" column={3}>
            <Descriptions.Item label="名称">{provider.name}</Descriptions.Item>
            <Descriptions.Item label="标识">{provider.slug}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <StatusTag value={provider.status} />
            </Descriptions.Item>
            <Descriptions.Item label="厂商">{provider.vendor}</Descriptions.Item>
            <Descriptions.Item label="适配器">{provider.adapter_type}</Descriptions.Item>
            <Descriptions.Item label="鉴权">{provider.auth_type}</Descriptions.Item>
            <Descriptions.Item label="Base URL" span={3}>
              {provider.base_url || "-"}
            </Descriptions.Item>
            <Descriptions.Item label="备注" span={3}>
              {provider.description || "-"}
            </Descriptions.Item>
          </Descriptions>
        </Card>
      ) : null}

      <Card
        className="admin-panel admin-provider-usage-panel"
        title="用量统计"
        extra={
          <Segmented
            size="small"
            value={usageWindow}
            options={[
              { label: "近24小时", value: "24h" },
              { label: "近7天", value: "7d" },
              { label: "近30天", value: "30d" },
            ]}
            onChange={(value) => setUsageWindow(value as "24h" | "7d" | "30d")}
          />
        }
      >
        <div className="admin-provider-usage-summary">
          <Statistic title="总消耗" value={formatUSD(usageStats?.summary.estimated_cost)} />
          <Statistic title="请求数" value={formatInteger(usageStats?.summary.request_count)} />
          <Statistic title="成功率" value={formatPercent(usageStats?.summary.success_count, usageStats?.summary.request_count)} />
          <Statistic title="总 Token" value={formatInteger(usageStats?.summary.total_tokens)} />
          <Statistic title="活跃用户" value={formatInteger(usageStats?.summary.active_user_count)} />
          <Statistic title="活跃 Key" value={formatInteger(usageStats?.summary.active_key_count)} />
        </div>

        <div className="admin-provider-usage-grid">
          <section className="admin-provider-chart">
            <div className="admin-provider-chart-head">
              <Typography.Text strong>总消耗趋势</Typography.Text>
              <Typography.Text type="secondary">{usageGranularity === "hour" ? "按小时" : "按天"}</Typography.Text>
            </div>
            {trendData.length > 0 ? (
              <ResponsiveContainer width="100%" height={260}>
                <LineChart data={trendData} margin={{ top: 10, right: 18, bottom: 6, left: 0 }}>
                  <CartesianGrid stroke="#D9E6EC" strokeDasharray="4 4" vertical={false} />
                  <XAxis dataKey="label" tickLine={false} axisLine={false} tick={{ fill: "#60757D", fontSize: 12 }} />
                  <YAxis tickLine={false} axisLine={false} tick={{ fill: "#60757D", fontSize: 12 }} tickFormatter={(value) => `$${value}`} />
                  <Tooltip formatter={(value) => formatUSD(Number(value))} labelStyle={{ color: "#223842" }} />
                  <Line type="monotone" dataKey="estimated_cost" stroke="#237BB2" strokeWidth={2.5} dot={false} activeDot={{ r: 4 }} />
                </LineChart>
              </ResponsiveContainer>
            ) : (
              <Empty className="admin-chart-empty" description="当前时间范围暂无消耗数据" />
            )}
          </section>

          <section className="admin-provider-chart">
            <div className="admin-provider-chart-head">
              <Typography.Text strong>模型消耗排行</Typography.Text>
              <Typography.Text type="secondary">Top 8</Typography.Text>
            </div>
            {modelCostData.length > 0 ? (
              <ResponsiveContainer width="100%" height={260}>
                <BarChart data={modelCostData} margin={{ top: 10, right: 16, bottom: 6, left: 0 }}>
                  <CartesianGrid stroke="#D9E6EC" strokeDasharray="4 4" vertical={false} />
                  <XAxis dataKey="label" tickLine={false} axisLine={false} tick={{ fill: "#60757D", fontSize: 12 }} interval={0} />
                  <YAxis tickLine={false} axisLine={false} tick={{ fill: "#60757D", fontSize: 12 }} tickFormatter={(value) => `$${value}`} />
                  <Tooltip formatter={(value) => formatUSD(Number(value))} labelStyle={{ color: "#223842" }} />
                  <Bar dataKey="estimated_cost" fill="#6FAECD" radius={[8, 8, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <Empty className="admin-chart-empty" description="当前时间范围暂无模型消耗" />
            )}
          </section>
        </div>

        <Table
          className="admin-table admin-provider-usage-table"
          rowKey={(record) => `${record.provider_model_id ?? "none"}-${record.upstream_model}-${record.public_model_name}`}
          size="middle"
          loading={usageQuery.isLoading}
          columns={modelUsageColumns}
          dataSource={usageStats?.models ?? []}
          pagination={false}
          scroll={{ x: 1100 }}
        />
      </Card>

      {provider && isClaudeMaxProxy ? (
        <Card
          className="admin-panel"
          title="Claude 登录状态"
          extra={
            <Space>
              <Button size="small" onClick={() => claudeStatusQuery.refetch()} loading={claudeStatusQuery.isFetching}>
                刷新状态
              </Button>
              <Button size="small" onClick={() => refreshTokenMutation.mutate()} loading={refreshTokenMutation.isPending}>
                刷新 Token
              </Button>
              <Button size="small" type="primary" onClick={() => probeMutation.mutate()} loading={probeMutation.isPending}>
                探测登录
              </Button>
            </Space>
          }
        >
          {claudeStatus?.error ? <Alert className="admin-status-alert" type={claudeStatus.reachable ? "warning" : "error"} message={claudeStatus.error} showIcon /> : null}
          <div className="admin-claude-status">
            <Statistic title="Proxy 状态" value={claudeStatus?.proxy_status ?? "-"} />
            <Statistic title="Token 剩余" value={formatTokenHours(claudeStatus?.token_hours)} />
            <Statistic title="预计过期时间" value={estimateExpiresAt(claudeStatus?.checked_at, claudeStatus?.token_hours)} />
            <div className="admin-claude-status-meta">
              <Typography.Text type="secondary">订阅类型</Typography.Text>
              <Typography.Text>{claudeStatus?.subscription_type || "-"}</Typography.Text>
            </div>
            <div className="admin-claude-status-meta">
              <Typography.Text type="secondary">限流档位</Typography.Text>
              <Typography.Text>{claudeStatus?.rate_limit_tier || "-"}</Typography.Text>
            </div>
            <div className="admin-claude-status-meta">
              <Typography.Text type="secondary">可达性</Typography.Text>
              <Tag color={claudeStatus?.reachable ? "success" : "error"}>{claudeStatus?.reachable ? "可达" : "不可达"}</Tag>
            </div>
            <div className="admin-claude-status-meta">
              <Typography.Text type="secondary">Claude Code 版本</Typography.Text>
              <Typography.Text>{claudeStatus?.cc_version || "-"}</Typography.Text>
            </div>
            <div className="admin-claude-status-meta">
              <Typography.Text type="secondary">检查时间</Typography.Text>
              <Typography.Text>{formatDateTime(claudeStatus?.checked_at)}</Typography.Text>
            </div>
          </div>
        </Card>
      ) : null}

      <Card
        className="admin-panel admin-table-panel"
        title="上游模型列表"
        extra={
          <Space>
            <Button size="small" onClick={() => routesQuery.refetch()}>
              刷新
            </Button>
          </Space>
        }
      >
        <ProviderModelRouteTable
          data={routesQuery.data?.items ?? []}
          loading={routesQuery.isLoading}
          onEdit={(providerModel) => {
            setEditing(providerModel);
            setDrawerOpen(true);
          }}
          onDelete={confirmDelete}
          onSetActive={(providerModel) => activateMutation.mutate(providerModel)}
        />
      </Card>

      <ProviderModelRouteDrawer
        open={drawerOpen}
        providerId={providerId}
        models={modelsQuery.data?.items ?? []}
        providerModel={editing}
        submitting={saveMutation.isPending}
        onClose={() => {
          setDrawerOpen(false);
          setEditing(undefined);
        }}
        onSubmit={(payload) => saveMutation.mutateAsync(payload).then(() => undefined)}
      />
    </div>
  );
}
