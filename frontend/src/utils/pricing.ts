export interface PricingFields {
  pricing_currency: string;
  pricing_input: number;
  pricing_output: number;
  pricing_cache: number;
}

function readNumber(value: unknown) {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === "string" && value.trim() !== "") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

export function pricingJSONToFields(value: Record<string, unknown> | undefined, defaultCurrency = "CNY"): PricingFields {
  return {
    pricing_currency: typeof value?.currency === "string" && value.currency.trim() ? value.currency : defaultCurrency,
    pricing_input: readNumber(value?.input),
    pricing_output: readNumber(value?.output),
    pricing_cache: readNumber(value?.cache ?? value?.cache_read),
  };
}

export function pricingFieldsToJSON(values: PricingFields): Record<string, unknown> {
  return {
    currency: values.pricing_currency?.trim() || "CNY",
    input: values.pricing_input ?? 0,
    output: values.pricing_output ?? 0,
    cache: values.pricing_cache ?? 0,
  };
}

export function formatPricing(value: Record<string, unknown> | undefined, fallbackCurrency = "CNY") {
  const pricing = pricingJSONToFields(value, fallbackCurrency);
  return `${pricing.pricing_currency} 输入 ${pricing.pricing_input} / 输出 ${pricing.pricing_output} / 缓存 ${pricing.pricing_cache}`;
}

export function formatPricingAmount(value: number, currency: string) {
  const normalizedCurrency = currency.trim().toUpperCase();
  const symbol = normalizedCurrency === "USD" ? "$" : normalizedCurrency === "CNY" || normalizedCurrency === "RMB" ? "￥" : `${currency} `;

  return `${symbol}${value}`;
}
