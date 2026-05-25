import { Tag } from "antd";

const colorByStatus: Record<string, string> = {
  active: "success",
  enabled: "success",
  disabled: "default",
  rate_limited: "warning",
  invalid: "error",
};

const labelByStatus: Record<string, string> = {
  active: "启用",
  enabled: "启用",
  disabled: "禁用",
  rate_limited: "受限",
  invalid: "无效",
};

export function StatusTag({ value }: { value: string }) {
  return <Tag color={colorByStatus[value] ?? "default"}>{labelByStatus[value] ?? value}</Tag>;
}
