import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button, Card, Modal, message } from "antd";
import { useState } from "react";

import { createProvider, deleteProvider, fetchProviders, updateProvider } from "../api/admin";
import { PageHeader } from "../components/PageHeader";
import { ProviderEditorDrawer } from "../features/providers/ProviderEditorDrawer";
import { ProviderTable } from "../features/providers/ProviderTable";
import type { Provider, ProviderInput } from "../types";

export function ProvidersPage() {
  const queryClient = useQueryClient();
  const query = useQuery({ queryKey: ["providers"], queryFn: fetchProviders });
  const [editing, setEditing] = useState<Provider | undefined>();
  const [drawerOpen, setDrawerOpen] = useState(false);

  const saveMutation = useMutation({
    mutationFn: (input: ProviderInput) => (editing ? updateProvider(editing.id, input) : createProvider(input)),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["providers"] });
      message.success("供应商已保存");
      setDrawerOpen(false);
      setEditing(undefined);
    },
    onError: (error) => message.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: deleteProvider,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["providers"] });
      message.success("供应商已删除");
    },
    onError: (error) => message.error(error.message),
  });

  function openCreate() {
    setEditing(undefined);
    setDrawerOpen(true);
  }

  function confirmDelete(provider: Provider) {
    Modal.confirm({
      title: "删除供应商",
      content: `确定删除 ${provider.name} 吗？如果该供应商下还有上游模型，系统会拒绝删除。`,
      okText: "删除",
      okButtonProps: { danger: true },
      cancelText: "取消",
      onOk: () => deleteMutation.mutateAsync(provider.id),
    });
  }

  return (
    <div className="page-stack">
      <PageHeader
        title="供应商"
        description="管理上游通道，例如 Anthropic 官方、Claude Code、Google Vertex AI。"
        actions={
          <>
            <Button icon={<ReloadOutlined />} onClick={() => query.refetch()} />
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              新建供应商
            </Button>
          </>
        }
      />
      <Card>
        <ProviderTable
          data={query.data?.items ?? []}
          loading={query.isLoading}
          onEdit={(provider) => {
            setEditing(provider);
            setDrawerOpen(true);
          }}
          onDelete={confirmDelete}
        />
      </Card>
      <ProviderEditorDrawer
        open={drawerOpen}
        provider={editing}
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
