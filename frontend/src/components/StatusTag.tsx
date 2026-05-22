import { Tag } from "antd";

const colorByStatus: Record<string, string> = {
  active: "success",
  enabled: "success",
  disabled: "default",
  rate_limited: "warning",
  invalid: "error",
};

export function StatusTag({ value }: { value: string }) {
  return <Tag color={colorByStatus[value] ?? "default"}>{value}</Tag>;
}
