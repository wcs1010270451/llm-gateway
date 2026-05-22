import { Button, Checkbox, Drawer, Form, Input, InputNumber, Select, Space } from "antd";
import { useEffect } from "react";

import type { Provider, ProviderModel, ProviderModelInput } from "../../types";
import { formatJSON, parseJSONObject } from "../../utils/json";
import { pricingFieldsToJSON, pricingJSONToFields, type PricingFields } from "../../utils/pricing";
import { PricingFormItems } from "./PricingFormItems";

interface ProviderModelEditorDrawerProps {
  open: boolean;
  providers: Provider[];
  providerModel?: ProviderModel;
  submitting: boolean;
  onClose: () => void;
  onSubmit: (input: ProviderModelInput) => Promise<void>;
}

interface ProviderModelFormValues extends PricingFields {
  provider_id: number;
  upstream_model: string;
  status: ProviderModel["status"];
  max_tokens: number;
  timeout_seconds: number;
  set_active: boolean;
  config_json_text: string;
}

export function ProviderModelEditorDrawer({
  open,
  providers,
  providerModel,
  submitting,
  onClose,
  onSubmit,
}: ProviderModelEditorDrawerProps) {
  const [form] = Form.useForm<ProviderModelFormValues>();

  useEffect(() => {
    if (!open) {
      return;
    }
    form.setFieldsValue({
      provider_id: providerModel?.provider_id,
      upstream_model: providerModel?.upstream_model ?? "",
      status: providerModel?.status ?? "enabled",
      max_tokens: providerModel?.max_tokens ?? 0,
      timeout_seconds: providerModel?.timeout_seconds ?? 300,
      set_active: false,
      ...pricingJSONToFields(providerModel?.pricing_json, "USD"),
      config_json_text: formatJSON(providerModel?.config_json),
    });
  }, [form, open, providerModel]);

  async function handleFinish(values: ProviderModelFormValues) {
    const pricingJSON = pricingFieldsToJSON(values);
    const configJSON = parseJSONObject(values.config_json_text);
    await onSubmit({
      provider_id: values.provider_id,
      upstream_model: values.upstream_model,
      status: values.status,
      max_tokens: values.max_tokens ?? 0,
      timeout_seconds: values.timeout_seconds ?? 300,
      input_cost_per_1m: values.pricing_input ?? 0,
      output_cost_per_1m: values.pricing_output ?? 0,
      pricing_json: pricingJSON,
      config_json: configJSON,
      set_active: values.set_active ?? false,
    });
  }

  return (
    <Drawer
      title={providerModel ? "编辑供应商模型" : "新增供应商模型"}
      width={580}
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
        <Form.Item name="provider_id" label="供应商" rules={[{ required: true, message: "请选择供应商" }]}>
          <Select
            showSearch
            optionFilterProp="label"
            options={providers.map((provider) => ({
              value: provider.id,
              label: `${provider.name} (${provider.adapter_type})`,
            }))}
          />
        </Form.Item>
        <Form.Item
          name="upstream_model"
          label="上游真实模型名"
          rules={[{ required: true, message: "请输入上游模型名" }]}
        >
          <Input placeholder="例如 gpt5-5 / gpt_5_5 / claude-sonnet-4-6" />
        </Form.Item>
        <Form.Item name="status" label="状态" rules={[{ required: true }]}>
          <Select
            options={[
              { value: "enabled", label: "启用" },
              { value: "disabled", label: "禁用" },
            ]}
          />
        </Form.Item>
        <Form.Item name="max_tokens" label="最大输出 Token">
          <InputNumber min={0} precision={0} style={{ width: "100%" }} />
        </Form.Item>
        <Form.Item name="timeout_seconds" label="超时时间（秒）">
          <InputNumber min={1} precision={0} style={{ width: "100%" }} />
        </Form.Item>
        <PricingFormItems currencyLabel="成本币种" />
        <Form.Item name="set_active" valuePropName="checked">
          <Checkbox>保存后设为当前上游</Checkbox>
        </Form.Item>
        <Form.Item name="config_json_text" label="扩展配置 JSON">
          <Input.TextArea rows={6} spellCheck={false} />
        </Form.Item>
      </Form>
    </Drawer>
  );
}
