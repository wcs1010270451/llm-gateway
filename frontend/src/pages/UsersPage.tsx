import { Button, Card, Modal, message } from "antd";
import { PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

import { createUser, deleteUser, fetchUsers, updateUser } from "../api/admin";
import { PageHeader } from "../components/PageHeader";
import { UserEditorDrawer } from "../features/users/UserEditorDrawer";
import { UserTable } from "../features/users/UserTable";
import type { User, UserInput } from "../types";

export function UsersPage() {
  const queryClient = useQueryClient();
  const query = useQuery({ queryKey: ["users"], queryFn: fetchUsers });
  const [editing, setEditing] = useState<User | undefined>();
  const [drawerOpen, setDrawerOpen] = useState(false);

  const saveMutation = useMutation({
    mutationFn: (input: UserInput) => (editing ? updateUser(editing.id, input) : createUser(input)),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["users"] });
      message.success("用户已保存");
      setDrawerOpen(false);
      setEditing(undefined);
    },
    onError: (error) => message.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: deleteUser,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["users"] });
      message.success("用户已删除");
    },
    onError: (error) => message.error(error.message),
  });

  function confirmDelete(user: User) {
    Modal.confirm({
      title: "删除用户",
      content: `确定删除 ${user.email} 吗？该用户的 API Key 也会被删除。`,
      okText: "删除",
      okButtonProps: { danger: true },
      cancelText: "取消",
      onOk: () => deleteMutation.mutateAsync(user.id),
    });
  }

  return (
    <div className="page-stack">
      <PageHeader
        title="用户"
        description="管理员添加和禁用用户；不提供公开注册。普通用户登录后只能管理自己的 Key。"
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
              新建用户
            </Button>
          </>
        }
      />
      <Card>
        <UserTable
          data={query.data?.items ?? []}
          loading={query.isLoading}
          onEdit={(user) => {
            setEditing(user);
            setDrawerOpen(true);
          }}
          onDelete={confirmDelete}
        />
      </Card>
      <UserEditorDrawer
        open={drawerOpen}
        user={editing}
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
