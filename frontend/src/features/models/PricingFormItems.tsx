import { Form, InputNumber, Select } from "antd";

export function PricingFormItems({ currencyLabel = "币种" }: { currencyLabel?: string }) {
  return (
    <>
      <Form.Item name="pricing_currency" label={currencyLabel} rules={[{ required: true, message: "请输入币种" }]}>
        <Select
          options={[
            { value: "CNY", label: "人民币（CNY）" },
            { value: "USD", label: "美元（USD）" },
          ]}
        />
      </Form.Item>
      <Form.Item name="pricing_input" label="输入价格 / 1M Token" rules={[{ required: true, message: "请输入输入价格" }]}>
        <InputNumber min={0} precision={6} style={{ width: "100%" }} />
      </Form.Item>
      <Form.Item name="pricing_output" label="输出价格 / 1M Token" rules={[{ required: true, message: "请输入输出价格" }]}>
        <InputNumber min={0} precision={6} style={{ width: "100%" }} />
      </Form.Item>
      <Form.Item name="pricing_cache" label="缓存价格 / 1M Token" rules={[{ required: true, message: "请输入缓存价格" }]}>
        <InputNumber min={0} precision={6} style={{ width: "100%" }} />
      </Form.Item>
    </>
  );
}
