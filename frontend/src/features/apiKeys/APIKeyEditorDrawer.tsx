import { Button, Modal, Form, Input, InputNumber, Select, Space } from "antd";
import { useEffect } from "react";

import type { APIKey, APIKeyInput } from "../../types";

interface APIKeyEditorDrawerProps {
  open: boolean;
  apiKey?: APIKey;
  submitting: boolean;
  onClose: () => void;
  onSubmit: (input: APIKeyInput) => Promise<void>;
}

interface APIKeyFormValues {
  name: string;
  status: APIKey["status"];
  rpm_limit: number;
  daily_request_limit: number;
  daily_token_limit: number;
}

export function APIKeyEditorDrawer({ open, apiKey, submitting, onClose, onSubmit }: APIKeyEditorDrawerProps) {
  const [form] = Form.useForm<APIKeyFormValues>();

  useEffect(() => {
    if (!open) {
      return;
    }
    form.setFieldsValue({
      name: apiKey?.name ?? "",
      status: apiKey?.status ?? "active",
      rpm_limit: apiKey?.rpm_limit ?? 0,
      daily_request_limit: apiKey?.daily_request_limit ?? 0,
      daily_token_limit: apiKey?.daily_token_limit ?? 0,
    });
  }, [apiKey, form, open]);

  async function handleFinish(values: APIKeyFormValues) {
    await onSubmit({
      name: values.name,
      status: values.status,
      rpm_limit: values.rpm_limit ?? 0,
      daily_request_limit: values.daily_request_limit ?? 0,
      daily_token_limit: values.daily_token_limit ?? 0,
      expires_at: null,
    });
  }

  return (
    <Modal
      title={apiKey ? "编辑 Key" : "新建 Key"}
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
        <Form.Item name="name" label="名称" rules={[{ required: true, message: "请输入名称" }]}>
          <Input placeholder="客服助手生产环境" />
        </Form.Item>
        <Form.Item name="status" label="状态" rules={[{ required: true }]}>
          <Select
            options={[
              { value: "active", label: "启用" },
              { value: "disabled", label: "禁用" },
            ]}
          />
        </Form.Item>
        <Form.Item name="rpm_limit" label="每分钟请求数">
          <InputNumber min={0} precision={0} style={{ width: "100%" }} />
        </Form.Item>
        <Form.Item name="daily_request_limit" label="每日请求数">
          <InputNumber min={0} precision={0} style={{ width: "100%" }} />
        </Form.Item>
        <Form.Item name="daily_token_limit" label="每日 Token 数">
          <InputNumber min={0} precision={0} style={{ width: "100%" }} />
        </Form.Item>
      </Form>
    </Modal>
  );
}
