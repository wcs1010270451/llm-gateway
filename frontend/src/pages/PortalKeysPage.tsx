import { PlusOutlined, ReloadOutlined, SafetyCertificateOutlined } from "@ant-design/icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button, Card, Modal, Typography, message } from "antd";
import { useState } from "react";
import { useNavigate } from "react-router";

import { createMyAPIKey, deleteMyAPIKey, fetchMyAPIKeys, updateMyAPIKey } from "../api/apiKeys";
import { PageHeader } from "../components/PageHeader";
import { APIKeyCreatedModal } from "../features/apiKeys/APIKeyCreatedModal";
import { APIKeyEditorDrawer } from "../features/apiKeys/APIKeyEditorDrawer";
import { APIKeyTable } from "../features/apiKeys/APIKeyTable";
import type { APIKey, APIKeyInput, CreatedAPIKey } from "../types";

export function PortalKeysPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const query = useQuery({ queryKey: ["me", "api-keys"], queryFn: fetchMyAPIKeys });
  const [editing, setEditing] = useState<APIKey | undefined>();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [created, setCreated] = useState<CreatedAPIKey | undefined>();

  const saveMutation = useMutation<CreatedAPIKey | APIKey, Error, APIKeyInput>({
    mutationFn: (input: APIKeyInput) => (editing ? updateMyAPIKey(editing.id, input) : createMyAPIKey(input)),
    onSuccess: async (result) => {
      await queryClient.invalidateQueries({ queryKey: ["me", "api-keys"] });
      setDrawerOpen(false);
      setEditing(undefined);
      if ("plain_key" in result) {
        setCreated(result);
        return;
      }
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
        eyebrow="ACCESS WORKSPACE"
        title="访问凭据"
        description="创建用于调用网关 API 的 Key，管理调用额度，并追踪每把凭据的使用情况。"
        actions={
          <>
            <Button icon={<ReloadOutlined />} onClick={() => query.refetch()} aria-label="刷新凭据列表" />
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
      <section className="portal-guidance" aria-label="凭据安全提示">
        <SafetyCertificateOutlined className="portal-guidance-icon" />
        <div>
          <Typography.Text strong>妥善保管完整 Key</Typography.Text>
          <Typography.Paragraph>
            完整 Key 会加密保存，仅在你点击复制时读取。仍建议将生产凭据保存在安全的密钥管理工具中。
          </Typography.Paragraph>
        </div>
      </section>
      <Card className="portal-panel portal-key-list" title="已创建的凭据" extra={<Typography.Text type="secondary">{query.data?.items.length ?? 0} 项</Typography.Text>}>
        <APIKeyTable
          data={query.data?.items ?? []}
          loading={query.isLoading}
          onOpen={(apiKey) => navigate(`/portal/keys/${apiKey.id}`, { state: { apiKey } })}
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
      <APIKeyCreatedModal created={created} onClose={() => setCreated(undefined)} />
    </div>
  );
}
