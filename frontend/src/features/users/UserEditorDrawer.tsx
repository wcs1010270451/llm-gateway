import { Button, Drawer, Form, Input, Select, Space } from "antd";
import { useEffect } from "react";

import type { User, UserInput } from "../../types";

interface UserEditorDrawerProps {
  open: boolean;
  user?: User;
  submitting: boolean;
  onClose: () => void;
  onSubmit: (input: UserInput) => Promise<void>;
}

interface UserFormValues {
  email: string;
  password?: string;
  display_name: string;
  role: User["role"];
  status: User["status"];
}

export function UserEditorDrawer({ open, user, submitting, onClose, onSubmit }: UserEditorDrawerProps) {
  const [form] = Form.useForm<UserFormValues>();

  useEffect(() => {
    if (!open) {
      return;
    }
    form.setFieldsValue({
      email: user?.email ?? "",
      password: "",
      display_name: user?.display_name ?? "",
      role: user?.role ?? "user",
      status: user?.status ?? "active",
    });
  }, [form, open, user]);

  async function handleFinish(values: UserFormValues) {
    await onSubmit({
      email: values.email,
      password: values.password?.trim() || undefined,
      display_name: values.display_name ?? "",
      role: values.role,
      status: values.status,
    });
  }

  return (
    <Drawer
      title={user ? "编辑用户" : "新建用户"}
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
      <Form layout="vertical" form={form} onFinish={handleFinish}>
        <Form.Item name="email" label="邮箱" rules={[{ required: true, message: "请输入邮箱" }]}>
          <Input placeholder="user@example.com" />
        </Form.Item>
        <Form.Item
          name="password"
          label="密码"
          rules={user ? [] : [{ required: true, message: "新建用户必须填写密码" }]}
        >
          <Input.Password placeholder={user ? "留空则保留原密码" : "请输入初始密码"} />
        </Form.Item>
        <Form.Item name="display_name" label="显示名称">
          <Input />
        </Form.Item>
        <Form.Item name="role" label="角色" rules={[{ required: true }]}>
          <Select
            options={[
              { value: "admin", label: "管理员" },
              { value: "user", label: "用户" },
            ]}
          />
        </Form.Item>
        <Form.Item name="status" label="状态" rules={[{ required: true }]}>
          <Select
            options={[
              { value: "active", label: "启用" },
              { value: "disabled", label: "禁用" },
            ]}
          />
        </Form.Item>
      </Form>
    </Drawer>
  );
}
