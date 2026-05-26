import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button, Card, Modal, Select, Typography, message } from "antd";
import { useState } from "react";

import { createModel, deleteModel, fetchActiveModelFamilies, fetchModelFamilies, fetchModelPage, updateModel } from "../api/admin";
import { PageHeader } from "../components/PageHeader";
import { ModelEditorDrawer } from "../features/models/ModelEditorDrawer";
import { ModelTable } from "../features/models/ModelTable";
import type { Model, ModelInput } from "../types";

function buildModelInput(model: Model, status: Model["status"]): ModelInput {
  return {
    name: model.name,
    display_name: model.display_name ?? "",
    family: model.family,
    modality: model.modality,
    status,
    description: model.description ?? "",
    pricing_json: model.pricing_json ?? {},
    config_json: model.config_json ?? {},
  };
}

export function ModelsPage() {
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [selectedFamily, setSelectedFamily] = useState<string>();
  const query = useQuery({
    queryKey: ["models", "page", page, selectedFamily],
    queryFn: () => fetchModelPage({ page, page_size: 10, family: selectedFamily }),
  });
  const familiesQuery = useQuery({ queryKey: ["model-families", "active"], queryFn: fetchActiveModelFamilies });
  const filterFamiliesQuery = useQuery({ queryKey: ["model-families"], queryFn: fetchModelFamilies });
  const [drawerOpen, setDrawerOpen] = useState(false);

  const models = query.data?.items ?? [];
  const familyOptions = (filterFamiliesQuery.data?.items ?? []).map((family) => ({
    label: family.display_name ? `${family.display_name} (${family.name})` : family.name,
    value: family.name,
  }));

  const saveMutation = useMutation({
    mutationFn: createModel,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["models"] });
      message.success("模型已保存");
      setDrawerOpen(false);
    },
    onError: (error) => message.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: deleteModel,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["models"] });
      message.success("模型已删除");
    },
    onError: (error) => message.error(error.message),
  });

  const toggleStatusMutation = useMutation({
    mutationFn: (model: Model) => {
      const nextStatus = model.status === "enabled" ? "disabled" : "enabled";
      return updateModel(model.id, buildModelInput(model, nextStatus));
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["models"] });
      message.success("模型状态已更新");
    },
    onError: (error) => message.error(error.message),
  });

  function toggleStatus(model: Model) {
    if (model.status === "disabled" && !model.active_provider_model_id) {
      message.warning("启用前需要先设置当前上游");
      return;
    }
    toggleStatusMutation.mutate(model);
  }

  function confirmDelete(model: Model) {
    Modal.confirm({
      title: "删除模型",
      content: `确定删除 ${model.name} 吗？如果该模型已经关联供应商上游模型，系统会拒绝删除。`,
      okText: "删除",
      okButtonProps: { danger: true },
      cancelText: "取消",
      onOk: () => deleteMutation.mutateAsync(model.id),
    });
  }

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="ROUTING CATALOG"
        title="模型"
        description="维护平台对外提供的模型能力，并手动选择当前上游供应商模型。"
        actions={
          <>
            <Button icon={<ReloadOutlined />} onClick={() => query.refetch()} aria-label="刷新模型列表" />
            <Button
              type="primary"
              icon={<PlusOutlined />}
              disabled={(familiesQuery.data?.items ?? []).length === 0}
              onClick={() => {
                setDrawerOpen(true);
              }}
            >
              新建模型
            </Button>
          </>
        }
      />
      <Card className="admin-panel admin-table-panel">
        <div className="admin-filter-bar" aria-label="模型筛选">
          <div>
            <Typography.Text className="admin-filter-label">模型系列</Typography.Text>
            <Select
              className="admin-filter-select"
              value={selectedFamily}
              options={familyOptions}
              allowClear
              showSearch
              optionFilterProp="label"
              placeholder="全部系列"
              onChange={(value) => {
                setSelectedFamily(value);
                setPage(1);
              }}
              aria-label="按模型系列筛选"
            />
          </div>
          <Typography.Text className="admin-filter-count">
            共 {query.data?.total ?? 0} 个模型
          </Typography.Text>
        </div>
        <ModelTable
          data={models}
          loading={query.isLoading}
          page={query.data?.page ?? page}
          pageSize={query.data?.page_size ?? 10}
          total={query.data?.total ?? 0}
          togglingModelId={toggleStatusMutation.isPending ? toggleStatusMutation.variables?.id : undefined}
          onDelete={confirmDelete}
          onToggleStatus={toggleStatus}
          onPageChange={setPage}
        />
      </Card>
      <ModelEditorDrawer
        open={drawerOpen}
        families={familiesQuery.data?.items ?? []}
        submitting={saveMutation.isPending}
        onClose={() => {
          setDrawerOpen(false);
        }}
        onSubmit={(input) => saveMutation.mutateAsync(input).then(() => undefined)}
      />
    </div>
  );
}
