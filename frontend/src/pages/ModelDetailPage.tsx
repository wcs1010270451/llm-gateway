import { EditOutlined, PlusOutlined } from "@ant-design/icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button, Card, Descriptions, Modal, Typography, message } from "antd";
import { useState } from "react";
import { useNavigate, useParams } from "react-router";

import {
  createProviderModel,
  deleteProviderModel,
  fetchActiveModelFamilies,
  fetchModel,
  fetchProviderModels,
  fetchProviders,
  setActiveProviderModel,
  updateModel,
  updateProviderModel,
} from "../api/admin";
import { PageHeader } from "../components/PageHeader";
import { StatusTag } from "../components/StatusTag";
import { ModelEditorDrawer } from "../features/models/ModelEditorDrawer";
import { ProviderModelEditorDrawer } from "../features/models/ProviderModelEditorDrawer";
import { ProviderModelTable } from "../features/models/ProviderModelTable";
import type { ModelInput, ProviderModel, ProviderModelInput } from "../types";
import { formatPricingAmount, pricingJSONToFields } from "../utils/pricing";

export function ModelDetailPage() {
  const { id = "" } = useParams();
  const modelId = Number(id);
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [editing, setEditing] = useState<ProviderModel | undefined>();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [modelDrawerOpen, setModelDrawerOpen] = useState(false);

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
  const familiesQuery = useQuery({ queryKey: ["model-families", "active"], queryFn: fetchActiveModelFamilies });

  const updateModelMutation = useMutation({
    mutationFn: (input: ModelInput) => updateModel(modelId, input),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["models"] }),
        queryClient.invalidateQueries({ queryKey: ["models", id] }),
      ]);
      message.success("模型已保存");
      setModelDrawerOpen(false);
    },
    onError: (error) => message.error(error.message),
  });

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
  const pricing = pricingJSONToFields(model?.pricing_json, "CNY");

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="MODEL ROUTING"
        title={model ? `模型映射：${model.name}` : `模型映射 #${id}`}
        description="维护供应商支持哪些平台模型，以及该模型在上游的真实名称。"
        actions={
          <>
            <Button onClick={() => navigate("/models")}>返回</Button>
            <Button icon={<EditOutlined />} disabled={!model} onClick={() => setModelDrawerOpen(true)}>
              编辑
            </Button>
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
        <Card className="admin-panel">
          <Descriptions className="admin-descriptions" size="small" column={3}>
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
          </Descriptions>
          <div className="admin-pricing-band" aria-label="平台价格，每百万 Token">
            <Typography.Text className="admin-pricing-title">平台价格 / 1M Token</Typography.Text>
            <div className="admin-pricing-items">
              <div className="admin-pricing-item">
                <Typography.Text className="admin-pricing-label">输入</Typography.Text>
                <Typography.Text className="admin-pricing-value">{formatPricingAmount(pricing.pricing_input, pricing.pricing_currency)}</Typography.Text>
              </div>
              <div className="admin-pricing-item">
                <Typography.Text className="admin-pricing-label">输出</Typography.Text>
                <Typography.Text className="admin-pricing-value">{formatPricingAmount(pricing.pricing_output, pricing.pricing_currency)}</Typography.Text>
              </div>
              <div className="admin-pricing-item">
                <Typography.Text className="admin-pricing-label">缓存</Typography.Text>
                <Typography.Text className="admin-pricing-value">{formatPricingAmount(pricing.pricing_cache, pricing.pricing_currency)}</Typography.Text>
              </div>
            </div>
          </div>
        </Card>
      ) : null}

      <Card
        className="admin-panel admin-table-panel"
        title="供应商模型映射"
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
      <ModelEditorDrawer
        open={modelDrawerOpen}
        model={model}
        families={familiesQuery.data?.items ?? []}
        submitting={updateModelMutation.isPending}
        onClose={() => setModelDrawerOpen(false)}
        onSubmit={(input) => updateModelMutation.mutateAsync(input).then(() => undefined)}
      />
    </div>
  );
}
