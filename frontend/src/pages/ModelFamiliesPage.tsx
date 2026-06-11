import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button, Card, Drawer, Empty, Form, Input, Modal, Select, Space, Table, Typography, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import type { Key } from "react";
import { useEffect, useMemo, useState } from "react";

import {
  createModel,
  createModelFamily,
  deleteModel,
  deleteModelFamily,
  fetchActiveModelFamilies,
  fetchModelFamilies,
  fetchModels,
  updateModel,
  updateModelFamily,
} from "../api/admin";
import { PageHeader } from "../components/PageHeader";
import { StatusTag } from "../components/StatusTag";
import { ModelEditorDrawer } from "../features/models/ModelEditorDrawer";
import { ModelTable } from "../features/models/ModelTable";
import type { Model, ModelFamily, ModelFamilyInput, ModelInput } from "../types";

interface ModelFamilyFormValues {
  name: string;
  display_name: string;
  status: ModelFamily["status"];
  description: string;
}

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

export function ModelFamiliesPage() {
  const queryClient = useQueryClient();
  const familiesQuery = useQuery({ queryKey: ["model-families"], queryFn: fetchModelFamilies });
  const activeFamiliesQuery = useQuery({ queryKey: ["model-families", "active"], queryFn: fetchActiveModelFamilies });
  const modelsQuery = useQuery({ queryKey: ["models"], queryFn: fetchModels });
  const [editing, setEditing] = useState<ModelFamily | undefined>();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [modelDrawerOpen, setModelDrawerOpen] = useState(false);
  const [initialModelFamily, setInitialModelFamily] = useState<string | undefined>();
  const [expandedFamilyKeys, setExpandedFamilyKeys] = useState<Key[]>([]);

  const modelsByFamily = useMemo(() => {
    const grouped = new Map<string, Model[]>();
    for (const model of modelsQuery.data?.items ?? []) {
      const list = grouped.get(model.family) ?? [];
      list.push(model);
      grouped.set(model.family, list);
    }
    return grouped;
  }, [modelsQuery.data?.items]);

  const saveMutation = useMutation({
    mutationFn: (input: ModelFamilyInput) => (editing ? updateModelFamily(editing.id, input) : createModelFamily(input)),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["model-families"] });
      await queryClient.invalidateQueries({ queryKey: ["model-families", "active"] });
      message.success("模型系列已保存");
      setDrawerOpen(false);
      setEditing(undefined);
    },
    onError: (error) => message.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: deleteModelFamily,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["model-families"] });
      await queryClient.invalidateQueries({ queryKey: ["model-families", "active"] });
      message.success("模型系列已删除");
    },
    onError: (error) => message.error(error.message),
  });

  const modelSaveMutation = useMutation({
    mutationFn: createModel,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["models"] });
      message.success("模型已保存");
      setModelDrawerOpen(false);
      setInitialModelFamily(undefined);
    },
    onError: (error) => message.error(error.message),
  });

  const modelDeleteMutation = useMutation({
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

  function confirmDelete(item: ModelFamily) {
    Modal.confirm({
      title: "删除模型系列",
      content: `确定删除 ${item.name} 吗？已使用该系列的模型不会被删除，但后续编辑需要重新选择系列。`,
      okText: "删除",
      okButtonProps: { danger: true },
      cancelText: "取消",
      onOk: () => deleteMutation.mutateAsync(item.id),
    });
  }

  function confirmDeleteModel(model: Model) {
    Modal.confirm({
      title: "删除模型",
      content: `确定删除 ${model.name} 吗？如果该模型已经关联供应商上游模型，系统会拒绝删除。`,
      okText: "删除",
      okButtonProps: { danger: true },
      cancelText: "取消",
      onOk: () => modelDeleteMutation.mutateAsync(model.id),
    });
  }

  function toggleModelStatus(model: Model) {
    if (model.status === "disabled" && !model.active_provider_model_id) {
      message.warning("启用前需要先设置当前上游");
      return;
    }
    toggleStatusMutation.mutate(model);
  }

  function openModelDrawer(family?: string) {
    setInitialModelFamily(family);
    setModelDrawerOpen(true);
  }

  function toggleExpanded(id: number) {
    setExpandedFamilyKeys((keys) => (keys.includes(id) ? keys.filter((key) => key !== id) : [...keys, id]));
  }

  const columns: ColumnsType<ModelFamily> = [
    {
      title: "系列",
      dataIndex: "name",
      fixed: "left",
      width: 150,
      render: (value, record) => (
        <div>
          <Typography.Text strong>{record.display_name || value}</Typography.Text>
          <div className="table-subtitle">{value}</div>
        </div>
      ),
    },
    {
      title: "模型数",
      key: "model_count",
      width: 110,
      render: (_, record) => modelsByFamily.get(record.name)?.length ?? 0,
    },
    { title: "状态", dataIndex: "status", width: 120, render: (value) => <StatusTag value={value} /> },
    { title: "备注", dataIndex: "description", render: (value) => value || "-" },
    {
      title: "操作",
      key: "actions",
      width: 210,
      fixed: "right",
      render: (_, record) => (
        <Space size={4}>
          <Button
            size="small"
            onClick={(event) => {
              event.stopPropagation();
              openModelDrawer(record.name);
            }}
          >
            新增模型
          </Button>
          <Button
            size="small"
            onClick={(event) => {
              event.stopPropagation();
              setEditing(record);
              setDrawerOpen(true);
            }}
          >
            编辑
          </Button>
          <Button
            size="small"
            danger
            onClick={(event) => {
              event.stopPropagation();
              confirmDelete(record);
            }}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="ROUTING CATALOG"
        title="模型"
        description="按模型系列组织平台模型，点击系列行展开或收起该系列下的模型。"
        actions={
          <>
            <Button
              icon={<ReloadOutlined />}
              onClick={() => {
                familiesQuery.refetch();
                modelsQuery.refetch();
              }}
              aria-label="刷新模型列表"
            />
            <Button icon={<PlusOutlined />} disabled={(activeFamiliesQuery.data?.items ?? []).length === 0} onClick={() => openModelDrawer()}>
              新建模型
            </Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setEditing(undefined);
                setDrawerOpen(true);
              }}
            >
              新建系列
            </Button>
          </>
        }
      />
      <Card className="admin-panel admin-table-panel">
        <Table
          className="admin-table admin-family-model-table"
          rowKey="id"
          columns={columns}
          dataSource={familiesQuery.data?.items ?? []}
          loading={familiesQuery.isLoading}
          pagination={false}
          scroll={{ x: "max-content" }}
          expandable={{
            expandedRowKeys: expandedFamilyKeys,
            onExpandedRowsChange: (keys) => setExpandedFamilyKeys([...keys]),
            expandedRowRender: (family) => {
              const models = modelsByFamily.get(family.name) ?? [];
              return (
                <div className="admin-family-models">
                  <div className="admin-family-models-head">
                    <Typography.Text className="admin-panel-note">
                      {family.display_name || family.name} 下共有 {models.length} 个模型
                    </Typography.Text>
                    <Button size="small" type="primary" onClick={() => openModelDrawer(family.name)}>
                      新增模型
                    </Button>
                  </div>
                  {models.length > 0 || modelsQuery.isLoading ? (
                    <ModelTable
                      data={models}
                      loading={modelsQuery.isLoading}
                      togglingModelId={toggleStatusMutation.isPending ? toggleStatusMutation.variables?.id : undefined}
                      onDelete={confirmDeleteModel}
                      onToggleStatus={toggleModelStatus}
                    />
                  ) : (
                    <Empty className="admin-empty" description="该系列下暂无模型" />
                  )}
                </div>
              );
            },
          }}
          onRow={(record) => ({
            className: "admin-table-row-action",
            tabIndex: 0,
            onClick: () => toggleExpanded(record.id),
            onKeyDown: (event) => {
              if (event.key === "Enter" || event.key === " ") {
                event.preventDefault();
                toggleExpanded(record.id);
              }
            },
          })}
        />
      </Card>
      <ModelFamilyDrawer
        open={drawerOpen}
        item={editing}
        submitting={saveMutation.isPending}
        onClose={() => {
          setDrawerOpen(false);
          setEditing(undefined);
        }}
        onSubmit={(input) => saveMutation.mutateAsync(input).then(() => undefined)}
      />
      <ModelEditorDrawer
        open={modelDrawerOpen}
        initialFamily={initialModelFamily}
        families={activeFamiliesQuery.data?.items ?? []}
        submitting={modelSaveMutation.isPending}
        onClose={() => {
          setModelDrawerOpen(false);
          setInitialModelFamily(undefined);
        }}
        onSubmit={(input) => modelSaveMutation.mutateAsync(input).then(() => undefined)}
      />
    </div>
  );
}

function ModelFamilyDrawer({
  open,
  item,
  submitting,
  onClose,
  onSubmit,
}: {
  open: boolean;
  item?: ModelFamily;
  submitting: boolean;
  onClose: () => void;
  onSubmit: (input: ModelFamilyInput) => Promise<void>;
}) {
  const [form] = Form.useForm<ModelFamilyFormValues>();

  useEffect(() => {
    if (!open) {
      return;
    }
    form.setFieldsValue({
      name: item?.name ?? "",
      display_name: item?.display_name ?? "",
      status: item?.status ?? "active",
      description: item?.description ?? "",
    });
  }, [form, item, open]);

  async function handleFinish(values: ModelFamilyFormValues) {
    await onSubmit({
      name: values.name,
      display_name: values.display_name ?? "",
      status: values.status,
      description: values.description ?? "",
    });
  }

  return (
    <Modal
      title={item ? "编辑族类" : "新建族类"}
      width={480}
      open={open}
      onCancel={onClose}
      destroyOnClose
      footer={
        <Space>
          <Button onClick={onClose}>取消</Button>
          <Button type="primary" loading={submitting} onClick={() => form.submit()}>
            保存
          </Button>
        </Space>
      }
    >
      <Form layout="vertical" form={form} onFinish={handleFinish} style={{ marginTop: 24 }}>
        <Form.Item name="name" label="系列标识" rules={[{ required: true, message: "请输入系列标识" }]}>
          <Input placeholder="claude / gpt / gemini" />
        </Form.Item>
        <Form.Item name="display_name" label="显示名称">
          <Input placeholder="Claude / GPT / Gemini" />
        </Form.Item>
        <Form.Item name="status" label="状态" rules={[{ required: true }]}>
          <Select
            options={[
              { value: "active", label: "启用" },
              { value: "disabled", label: "禁用" },
            ]}
          />
        </Form.Item>
        <Form.Item name="description" label="备注">
          <Input.TextArea rows={3} />
        </Form.Item>
      </Form>
    </Modal>
  );
}
