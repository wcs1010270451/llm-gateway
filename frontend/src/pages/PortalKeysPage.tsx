import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button, Card, Modal, message } from "antd";
import { useState } from "react";
import { useNavigate } from "react-router";

import { createMyAPIKey, deleteMyAPIKey, fetchMyAPIKeys, updateMyAPIKey } from "../api/apiKeys";
import { PageHeader } from "../components/PageHeader";
import { APIKeyEditorDrawer } from "../features/apiKeys/APIKeyEditorDrawer";
import { APIKeyTable } from "../features/apiKeys/APIKeyTable";
import type { APIKey, APIKeyInput, CreatedAPIKey } from "../types";

export function PortalKeysPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const query = useQuery({ queryKey: ["me", "api-keys"], queryFn: fetchMyAPIKeys });
  const [editing, setEditing] = useState<APIKey | undefined>();
  const [drawerOpen, setDrawerOpen] = useState(false);

  const saveMutation = useMutation<CreatedAPIKey | APIKey, Error, APIKeyInput>({
    mutationFn: (input: APIKeyInput) => (editing ? updateMyAPIKey(editing.id, input) : createMyAPIKey(input)),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["me", "api-keys"] });
      setDrawerOpen(false);
      setEditing(undefined);
      message.success("Key 已保存");
    },
    onError: (error) => message.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: deleteMyAPIKey,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["me", "api-keys"] });
      message.success("Key 已删除");
    },
    onError: (error) => message.error(error.message),
  });

  function confirmDelete(apiKey: APIKey) {
    Modal.confirm({
      title: "删除 Key",
      content: `确定删除 ${apiKey.name} 吗？删除后无法恢复。`,
      okText: "删除",
      okButtonProps: { danger: true },
      cancelText: "取消",
      onOk: () => deleteMutation.mutateAsync(apiKey.id),
    });
  }

  return (
    <div className="page-stack">
      <PageHeader
        title="我的 Key"
        description="创建用于调用网关 API 的下游 Key，并查看每把 Key 的使用情况。"
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
              新建 Key
            </Button>
          </>
        }
      />
      <Card>
        <APIKeyTable
          data={query.data?.items ?? []}
          loading={query.isLoading}
          onDetail={(apiKey) => navigate(`/portal/keys/${apiKey.id}`, { state: { apiKey } })}
          onEdit={(apiKey) => {
            setEditing(apiKey);
            setDrawerOpen(true);
          }}
          onDelete={confirmDelete}
        />
      </Card>
      <APIKeyEditorDrawer
        open={drawerOpen}
        apiKey={editing}
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
