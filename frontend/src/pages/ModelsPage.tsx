import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button, Card, Modal, message } from "antd";
import { useState } from "react";

import { createModel, deleteModel, fetchActiveModelFamilies, fetchModels, updateModel } from "../api/admin";
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
  const query = useQuery({ queryKey: ["models"], queryFn: fetchModels });
  const familiesQuery = useQuery({ queryKey: ["model-families", "active"], queryFn: fetchActiveModelFamilies });
  const [editing, setEditing] = useState<Model | undefined>();
  const [drawerOpen, setDrawerOpen] = useState(false);

  const saveMutation = useMutation({
    mutationFn: (input: ModelInput) => (editing ? updateModel(editing.id, input) : createModel(input)),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["models"] });
      message.success("模型已保存");
      setDrawerOpen(false);
      setEditing(undefined);
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
        title="模型"
        description="维护平台对外提供的模型能力，并手动选择当前上游供应商模型。"
        actions={
          <>
            <Button icon={<ReloadOutlined />} onClick={() => query.refetch()} />
            <Button
              type="primary"
              icon={<PlusOutlined />}
              disabled={(familiesQuery.data?.items ?? []).length === 0}
              onClick={() => {
                setEditing(undefined);
                setDrawerOpen(true);
              }}
            >
              新建模型
            </Button>
          </>
        }
      />
      <Card>
        <ModelTable
          data={query.data?.items ?? []}
          loading={query.isLoading}
          togglingModelId={toggleStatusMutation.isPending ? toggleStatusMutation.variables?.id : undefined}
          onEdit={(model) => {
            setEditing(model);
            setDrawerOpen(true);
          }}
          onDelete={confirmDelete}
          onToggleStatus={toggleStatus}
        />
      </Card>
      <ModelEditorDrawer
        open={drawerOpen}
        model={editing}
        families={familiesQuery.data?.items ?? []}
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
