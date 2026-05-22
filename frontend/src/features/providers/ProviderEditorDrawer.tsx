import { Button, Drawer, Form, Input, Select, Space } from "antd";
import { useEffect } from "react";

import type { Provider, ProviderInput } from "../../types";
import { formatJSON, parseJSONObject } from "../../utils/json";

interface ProviderEditorDrawerProps {
  open: boolean;
  provider?: Provider;
  submitting: boolean;
  onClose: () => void;
  onSubmit: (input: ProviderInput) => Promise<void>;
}

interface ProviderFormValues {
  name: string;
  slug: string;
  vendor: Provider["vendor"];
  adapter_type: Provider["adapter_type"];
  auth_type: Provider["auth_type"];
  base_url: string;
  api_key_encrypted?: string;
  config_json_text: string;
  status: Provider["status"];
  description: string;
}

export function ProviderEditorDrawer({ open, provider, submitting, onClose, onSubmit }: ProviderEditorDrawerProps) {
  const [form] = Form.useForm<ProviderFormValues>();

  useEffect(() => {
    if (!open) {
      return;
    }

    form.setFieldsValue({
      name: provider?.name ?? "",
      slug: provider?.slug ?? "",
      vendor: provider?.vendor ?? "anthropic",
      adapter_type: provider?.adapter_type ?? "anthropic",
      auth_type: provider?.auth_type ?? "api_key",
      base_url: provider?.base_url ?? "",
      api_key_encrypted: "",
      config_json_text: formatJSON(provider?.config_json),
      status: provider?.status ?? "active",
      description: provider?.description ?? "",
    });
  }, [form, open, provider]);

  async function handleFinish(values: ProviderFormValues) {
    const configJSON = parseJSONObject(values.config_json_text);
    await onSubmit({
      name: values.name,
      slug: values.slug,
      vendor: values.vendor,
      adapter_type: values.adapter_type,
      auth_type: values.auth_type,
      base_url: values.base_url ?? "",
      api_key_encrypted: values.api_key_encrypted?.trim() || undefined,
      config_json: configJSON,
      status: values.status,
      description: values.description ?? "",
    });
  }

  return (
    <Drawer
      title={provider ? "编辑供应商" : "新建供应商"}
      width={520}
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
        <Form.Item name="name" label="名称" rules={[{ required: true, message: "请输入名称" }]}>
          <Input placeholder="Anthropic 官方" />
        </Form.Item>

        <Form.Item name="slug" label="唯一标识" rules={[{ required: true, message: "请输入唯一标识" }]}>
          <Input placeholder="anthropic-official" />
        </Form.Item>

        <Form.Item name="vendor" label="厂商" rules={[{ required: true }]}>
          <Select
            options={[
              { value: "openai", label: "OpenAI" },
              { value: "anthropic", label: "Anthropic" },
              { value: "google", label: "Google" },
              { value: "custom", label: "Custom" },
            ]}
          />
        </Form.Item>

        <Form.Item name="adapter_type" label="适配器" rules={[{ required: true }]}>
          <Select
            options={[
              { value: "openai_compatible", label: "OpenAI Compatible" },
              { value: "anthropic", label: "Anthropic API" },
              { value: "claude_code", label: "Claude Code" },
              { value: "gemini", label: "Gemini API" },
              { value: "vertexai", label: "Vertex AI" },
            ]}
          />
        </Form.Item>

        <Form.Item name="auth_type" label="鉴权方式" rules={[{ required: true }]}>
          <Select
            options={[
              { value: "api_key", label: "API Key" },
              { value: "local_oauth", label: "Local OAuth" },
              { value: "adc", label: "ADC" },
              { value: "none", label: "None" },
            ]}
          />
        </Form.Item>

        <Form.Item name="base_url" label="Base URL">
          <Input placeholder="https://api.anthropic.com" />
        </Form.Item>

        <Form.Item name="api_key_encrypted" label="上游 API Key">
          <Input.Password placeholder={provider ? "留空则保留原值" : "需要 API Key 时填写"} />
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

        <Form.Item name="config_json_text" label="扩展配置 JSON">
          <Input.TextArea rows={6} spellCheck={false} />
        </Form.Item>
      </Form>
    </Drawer>
  );
}
