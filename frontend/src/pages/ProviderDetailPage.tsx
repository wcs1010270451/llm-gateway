import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button, Card, Descriptions, Modal, Space, message } from "antd";
import { useState } from "react";
import { useNavigate, useParams } from "react-router";

import {
  createProviderModelRoute,
  deleteProviderModelRoute,
  fetchModels,
  fetchProvider,
  fetchProviderModelRoutes,
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
        title={provider ? `供应商：${provider.name}` : `供应商 #${id}`}
        description="查看该供应商提供的上游模型，并维护它们与平台模型的映射。"
        actions={
          <>
            <Button onClick={() => navigate("/providers")}>返回</Button>
            <Button
              icon={<ReloadOutlined />}
              onClick={() => {
                providerQuery.refetch();
                routesQuery.refetch();
                modelsQuery.refetch();
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
        <Card>
          <Descriptions size="small" column={3}>
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
