import { Button, Checkbox, Modal, Form, Input, InputNumber, Select, Space } from "antd";
import { useEffect } from "react";

import { PricingFormItems } from "../models/PricingFormItems";
import type { Model, ProviderModel, ProviderModelInput } from "../../types";
import { formatJSON, parseJSONObject } from "../../utils/json";
import { pricingFieldsToJSON, pricingJSONToFields, type PricingFields } from "../../utils/pricing";

export interface ProviderModelRouteSubmit {
  input: ProviderModelInput;
}

interface ProviderModelRouteDrawerProps {
  open: boolean;
  providerId: number;
  models: Model[];
  providerModel?: ProviderModel;
  submitting: boolean;
  onClose: () => void;
  onSubmit: (payload: ProviderModelRouteSubmit) => Promise<void>;
}

interface ProviderModelRouteFormValues extends PricingFields {
  model_id?: number;
  upstream_model: string;
  status: ProviderModel["status"];
  max_tokens: number;
  timeout_seconds: number;
  set_active: boolean;
  config_json_text: string;
}

export function ProviderModelRouteDrawer({
  open,
  providerId,
  models,
  providerModel,
  submitting,
  onClose,
  onSubmit,
}: ProviderModelRouteDrawerProps) {
  const [form] = Form.useForm<ProviderModelRouteFormValues>();
  const selectedModelID = Form.useWatch("model_id", form);

  useEffect(() => {
    if (!open) {
      return;
    }
    form.setFieldsValue({
      model_id: providerModel?.model_id,
      upstream_model: providerModel?.upstream_model ?? "",
      status: providerModel?.status ?? "enabled",
      max_tokens: providerModel?.max_tokens ?? 0,
      timeout_seconds: providerModel?.timeout_seconds ?? 300,
      set_active: false,
      ...pricingJSONToFields(providerModel?.pricing_json, "USD"),
      config_json_text: formatJSON(providerModel?.config_json),
    });
  }, [form, open, providerModel]);

  async function handleFinish(values: ProviderModelRouteFormValues) {
    const pricingJSON = pricingFieldsToJSON(values);
    const configJSON = parseJSONObject(values.config_json_text);
    await onSubmit({
      input: {
        provider_id: providerId,
        model_id: values.model_id,
        upstream_model: values.upstream_model,
        status: values.status,
        max_tokens: values.max_tokens ?? 0,
        timeout_seconds: values.timeout_seconds ?? 300,
        input_cost_per_1m: values.pricing_input ?? 0,
        output_cost_per_1m: values.pricing_output ?? 0,
        pricing_json: pricingJSON,
        config_json: configJSON,
        set_active: values.set_active ?? false,
      },
    });
  }

  return (
    <Modal
      title={providerModel ? "编辑上游模型" : "新增上游模型"}
      width={580}
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
        <Form.Item name="model_id" label="平台模型">
          <Select
            allowClear
            showSearch
            optionFilterProp="label"
            placeholder="暂不关联平台模型"
            options={models.map((model) => ({
              value: model.id,
              label: model.display_name ? `${model.display_name} (${model.name})` : model.name,
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
          <Checkbox disabled={!selectedModelID}>保存后设为该平台模型的当前上游</Checkbox>
        </Form.Item>
        <Form.Item name="config_json_text" label="扩展配置 JSON">
          <Input.TextArea rows={6} spellCheck={false} />
        </Form.Item>
      </Form>
    </Modal>
  );
}
