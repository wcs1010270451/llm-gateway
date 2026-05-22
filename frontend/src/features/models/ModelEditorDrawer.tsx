import { Button, Drawer, Form, Input, Select, Space } from "antd";
import { useEffect } from "react";

import type { Model, ModelFamily, ModelInput } from "../../types";
import { formatJSON, parseJSONObject } from "../../utils/json";
import { pricingFieldsToJSON, pricingJSONToFields, type PricingFields } from "../../utils/pricing";
import { PricingFormItems } from "./PricingFormItems";

interface ModelEditorDrawerProps {
  open: boolean;
  model?: Model;
  families: ModelFamily[];
  submitting: boolean;
  onClose: () => void;
  onSubmit: (input: ModelInput) => Promise<void>;
}

interface ModelFormValues extends PricingFields {
  name: string;
  display_name: string;
  family: string;
  modality: Model["modality"];
  status: Model["status"];
  description: string;
  config_json_text: string;
}

export function ModelEditorDrawer({ open, model, families, submitting, onClose, onSubmit }: ModelEditorDrawerProps) {
  const [form] = Form.useForm<ModelFormValues>();
  const hasActiveRoute = Boolean(model?.active_provider_model_id);

  useEffect(() => {
    if (!open) {
      return;
    }
    form.setFieldsValue({
      name: model?.name ?? "",
      display_name: model?.display_name ?? "",
      family: model?.family ?? families[0]?.name ?? "",
      modality: model?.modality ?? "text",
      status: model?.status ?? "disabled",
      description: model?.description ?? "",
      ...pricingJSONToFields(model?.pricing_json, "CNY"),
      config_json_text: formatJSON(model?.config_json),
    });
  }, [families, form, open, model]);

  async function handleFinish(values: ModelFormValues) {
    const pricingJSON = pricingFieldsToJSON(values);
    const configJSON = parseJSONObject(values.config_json_text);
    await onSubmit({
      name: values.name,
      display_name: values.display_name ?? "",
      family: values.family,
      modality: values.modality,
      status: values.status,
      description: values.description ?? "",
      pricing_json: pricingJSON,
      config_json: configJSON,
    });
  }

  return (
    <Drawer
      title={model ? "编辑模型" : "新建模型"}
      width={560}
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
        <Form.Item name="name" label="对外模型名" rules={[{ required: true, message: "请输入模型名" }]}>
          <Input placeholder="gpt5.5" />
        </Form.Item>
        <Form.Item name="display_name" label="显示名称">
          <Input placeholder="GPT-5.5" />
        </Form.Item>
        <Form.Item name="family" label="系列" rules={[{ required: true, message: "请选择模型系列" }]}>
          <Select
            placeholder="选择模型系列"
            options={families.map((family) => ({
              value: family.name,
              label: family.display_name ? `${family.display_name} (${family.name})` : family.name,
            }))}
          />
        </Form.Item>
        <Form.Item name="modality" label="能力" rules={[{ required: true }]}>
          <Select
            options={[
              { value: "text", label: "Text" },
              { value: "vision", label: "Vision" },
              { value: "embedding", label: "Embedding" },
              { value: "multimodal", label: "Multimodal" },
            ]}
          />
        </Form.Item>
        <Form.Item
          name="status"
          label="状态"
          rules={[{ required: true }]}
          tooltip="新建模型默认禁用。只有设置当前上游后，模型才可以启用。"
        >
          <Select
            options={[
              { value: "enabled", label: "启用", disabled: !hasActiveRoute },
              { value: "disabled", label: "禁用" },
            ]}
          />
        </Form.Item>
        <Form.Item name="description" label="备注">
          <Input.TextArea rows={3} />
        </Form.Item>
        <PricingFormItems currencyLabel="计价币种" />
        <Form.Item name="config_json_text" label="扩展配置 JSON">
          <Input.TextArea rows={6} spellCheck={false} />
        </Form.Item>
      </Form>
    </Drawer>
  );
}
