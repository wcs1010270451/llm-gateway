export interface Provider {
  id: number;
  name: string;
  slug: string;
  vendor: "openai" | "anthropic" | "google" | "custom";
  adapter_type: "openai_compatible" | "anthropic" | "claude_code" | "gemini" | "vertexai";
  auth_type: "api_key" | "local_oauth" | "adc" | "claude_oauth" | "none";
  base_url: string;
  config_json: Record<string, unknown>;
  status: "active" | "disabled";
  description: string;
  created_at: string;
  updated_at: string;
}

export interface ProviderInput {
  name: string;
  slug: string;
  vendor: Provider["vendor"];
  adapter_type: Provider["adapter_type"];
  auth_type: Provider["auth_type"];
  base_url: string;
  api_key_encrypted?: string;
  config_json: Record<string, unknown>;
  status: Provider["status"];
  description: string;
}
