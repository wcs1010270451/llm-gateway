export function formatJSON(value: Record<string, unknown> | undefined) {
  return JSON.stringify(value ?? {}, null, 2);
}

export function parseJSONObject(value: string | undefined) {
  if (!value || value.trim() === "") {
    return {};
  }

  const parsed = JSON.parse(value);
  if (parsed === null || Array.isArray(parsed) || typeof parsed !== "object") {
    throw new Error("JSON 必须是对象");
  }
  return parsed as Record<string, unknown>;
}
