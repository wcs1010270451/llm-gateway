import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button, Card, Drawer, Form, Input, Modal, Select, Space, Table, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useEffect, useState } from "react";

import { createModelFamily, deleteModelFamily, fetchModelFamilies, updateModelFamily } from "../api/admin";
import { PageHeader } from "../components/PageHeader";
import { StatusTag } from "../components/StatusTag";
import type { ModelFamily, ModelFamilyInput } from "../types";

interface ModelFamilyFormValues {
  name: string;
  display_name: string;
  status: ModelFamily["status"];
  description: string;
}

export function ModelFamiliesPage() {
  const queryClient = useQueryClient();
  const query = useQuery({ queryKey: ["model-families"], queryFn: fetchModelFamilies });
  const [editing, setEditing] = useState<ModelFamily | undefined>();
  const [drawerOpen, setDrawerOpen] = useState(false);

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

  const columns: ColumnsType<ModelFamily> = [
    { title: "系列标识", dataIndex: "name" },
    { title: "显示名称", dataIndex: "display_name", render: (value) => value || "-" },
    { title: "状态", dataIndex: "status", width: 120, render: (value) => <StatusTag value={value} /> },
    { title: "备注", dataIndex: "description", render: (value) => value || "-" },
    {
      title: "操作",
      key: "actions",
      width: 140,
      render: (_, record) => (
        <Space size={4}>
          <Button
            size="small"
            onClick={() => {
              setEditing(record);
              setDrawerOpen(true);
            }}
          >
            编辑
          </Button>
          <Button size="small" danger onClick={() => confirmDelete(record)}>
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div className="page-stack">
      <PageHeader
        title="模型系列"
        description="维护模型系列字典，新增模型时从这里选择系列。"
        actions={
          <>
            <Button icon={<ReloadOutlined />} onClick={() => query.refetch()} />
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
      <Card>
        <Table rowKey="id" columns={columns} dataSource={query.data?.items ?? []} loading={query.isLoading} pagination={false} />
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

  return (
    <Drawer
      title={item ? "编辑模型系列" : "新建模型系列"}
      width={480}
      open={open}
      onClose={onClose}
      destroyOnHidden
      extra={
        <Space>
          <Button onClick={onClose}>取消</Button>
          <Button type="primary" loading={submitting} onClick={() => form.submit()}>
            保存
          </Button>
        </Space>
      }
    >
      <Form layout="vertical" form={form} onFinish={onSubmit}>
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
    </Drawer>
  );
}
