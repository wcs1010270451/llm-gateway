import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Alert, Button, Card, Descriptions, Modal, Space, Statistic, Tag, Typography, message } from "antd";
import { useState } from "react";
import { useNavigate, useParams } from "react-router";

import {
  createProviderModelRoute,
  deleteProviderModelRoute,
  fetchClaudeProxyStatuses,
  fetchModels,
  fetchProvider,
  fetchProviderModelRoutes,
  probeClaudeProxy,
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

export function ProviderDetailPage() {
  const { id = "" } = useParams();
  const providerId = Number(id);
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [editing, setEditing] = useState<ProviderModel | undefined>();
  const [drawerOpen, setDrawerOpen] = useState(false);

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
      await queryClient.invalidateQueries({ queryKey: ["claude-proxies"] });
      if (result.probe_ok) {
        message.success("Claude 代理探测成功");
      } else {
        message.error(result.error || "Claude 代理探测失败");
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

      {provider && isClaudeMaxProxy ? (
        <Card
          className="admin-panel"
          title="Claude 登录状态"
          extra={
            <Space>
              <Button size="small" onClick={() => claudeStatusQuery.refetch()} loading={claudeStatusQuery.isFetching}>
                刷新状态
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
              <Typography.Text type="secondary">可达性</Typography.Text>
              <Tag color={claudeStatus?.reachable ? "success" : "error"}>{claudeStatus?.reachable ? "可达" : "不可达"}</Tag>
            </div>
            <div className="admin-claude-status-meta">
              <Typography.Text type="secondary">Claude Code</Typography.Text>
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
