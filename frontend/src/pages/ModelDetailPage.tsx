import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button, Card, Descriptions, Modal, Space, message } from "antd";
import { useState } from "react";
import { useParams } from "react-router";

import {
  createProviderModel,
  deleteProviderModel,
  fetchModel,
  fetchProviderModels,
  fetchProviders,
  setActiveProviderModel,
  updateProviderModel,
} from "../api/admin";
import { PageHeader } from "../components/PageHeader";
import { StatusTag } from "../components/StatusTag";
import { ProviderModelEditorDrawer } from "../features/models/ProviderModelEditorDrawer";
import { ProviderModelTable } from "../features/models/ProviderModelTable";
import type { ProviderModel, ProviderModelInput } from "../types";
import { formatPricing } from "../utils/pricing";

export function ModelDetailPage() {
  const { id = "" } = useParams();
  const modelId = Number(id);
  const queryClient = useQueryClient();
  const [editing, setEditing] = useState<ProviderModel | undefined>();
  const [drawerOpen, setDrawerOpen] = useState(false);

  const modelQuery = useQuery({
    queryKey: ["models", id],
    queryFn: () => fetchModel(id),
    enabled: id !== "",
  });
  const providerModelsQuery = useQuery({
    queryKey: ["models", id, "provider-models"],
    queryFn: () => fetchProviderModels(id),
    enabled: id !== "",
  });
  const providersQuery = useQuery({ queryKey: ["providers"], queryFn: fetchProviders });

  const saveMutation = useMutation({
    mutationFn: (input: ProviderModelInput) =>
      editing ? updateProviderModel(modelId, editing.id, input) : createProviderModel(modelId, input),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["models"] }),
        queryClient.invalidateQueries({ queryKey: ["models", id] }),
        queryClient.invalidateQueries({ queryKey: ["models", id, "provider-models"] }),
      ]);
      message.success("供应商模型已保存");
      setDrawerOpen(false);
      setEditing(undefined);
    },
    onError: (error) => message.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: (providerModelID: number) => deleteProviderModel(modelId, providerModelID),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["models"] }),
        queryClient.invalidateQueries({ queryKey: ["models", id] }),
        queryClient.invalidateQueries({ queryKey: ["models", id, "provider-models"] }),
      ]);
      message.success("供应商模型已删除");
    },
    onError: (error) => message.error(error.message),
  });

  const activateMutation = useMutation({
    mutationFn: (providerModelID: number) => setActiveProviderModel(modelId, providerModelID),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["models"] }),
        queryClient.invalidateQueries({ queryKey: ["models", id] }),
        queryClient.invalidateQueries({ queryKey: ["models", id, "provider-models"] }),
      ]);
      message.success("当前上游已切换");
    },
    onError: (error) => message.error(error.message),
  });

  function confirmDelete(providerModel: ProviderModel) {
    Modal.confirm({
      title: "删除供应商模型",
      content: `确定删除 ${providerModel.provider?.name ?? "该供应商"} 的 ${providerModel.upstream_model} 映射吗？`,
      okText: "删除",
      okButtonProps: { danger: true },
      cancelText: "取消",
      onOk: () => deleteMutation.mutateAsync(providerModel.id),
    });
  }

  const model = modelQuery.data;

  return (
    <div className="page-stack">
      <PageHeader
        title={model ? `模型映射：${model.name}` : `模型映射 #${id}`}
        description="维护供应商支持哪些平台模型，以及该模型在上游的真实名称。"
        actions={
          <>
            <Button
              icon={<ReloadOutlined />}
              onClick={() => {
                modelQuery.refetch();
                providerModelsQuery.refetch();
                providersQuery.refetch();
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
              新增供应商模型
            </Button>
          </>
        }
      />

      {model ? (
        <Card>
          <Descriptions size="small" column={3}>
            <Descriptions.Item label="对外模型名">{model.name}</Descriptions.Item>
            <Descriptions.Item label="显示名称">{model.display_name || "-"}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <StatusTag value={model.status} />
            </Descriptions.Item>
            <Descriptions.Item label="当前供应商">
              {model.active_provider_model?.provider?.name ?? "-"}
            </Descriptions.Item>
            <Descriptions.Item label="上游模型">
              {model.active_provider_model?.upstream_model ?? "-"}
            </Descriptions.Item>
            <Descriptions.Item label="系列">{model.family || "-"}</Descriptions.Item>
            <Descriptions.Item label="平台价格 / 1M">{formatPricing(model.pricing_json, "CNY")}</Descriptions.Item>
          </Descriptions>
        </Card>
      ) : null}

      <Card
        title="供应商模型映射"
        extra={
          <Space>
            <Button size="small" onClick={() => providerModelsQuery.refetch()}>
              刷新
            </Button>
          </Space>
        }
      >
        <ProviderModelTable
          data={providerModelsQuery.data?.items ?? []}
          loading={providerModelsQuery.isLoading}
          activeProviderModelId={model?.active_provider_model_id}
          onEdit={(providerModel) => {
            setEditing(providerModel);
            setDrawerOpen(true);
          }}
          onDelete={confirmDelete}
          onSetActive={(providerModel) => activateMutation.mutate(providerModel.id)}
        />
      </Card>

      <ProviderModelEditorDrawer
        open={drawerOpen}
        providers={providersQuery.data?.items ?? []}
        providerModel={editing}
        submitting={saveMutation.isPending}
        onClose={() => {
          setDrawerOpen(false);
          setEditing(undefined);
        }}
        onSubmit={(input) => saveMutation.mutateAsync(input).then(() => undefined)}
      />
    </div>
  );
}
