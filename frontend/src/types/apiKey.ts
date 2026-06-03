import type { User } from "./user";

export interface APIKey {
  id: number;
  user_id: number;
  name: string;
  masked_key: string;
  status: "active" | "disabled";
  rpm_limit: number;
  daily_request_limit: number;
  daily_token_limit: number;
  expires_at?: string;
  last_used_at?: string;
  last_error_at?: string;
  last_error_message: string;
  created_at: string;
  updated_at: string;
  user?: User;
}

export interface APIKeyInput {
  name: string;
  status: APIKey["status"];
  rpm_limit: number;
  daily_request_limit: number;
  daily_token_limit: number;
  expires_at?: string | null;
}

export interface CreatedAPIKey {
  api_key: APIKey;
  plain_key: string;
}

export interface RevealedAPIKey {
  plain_key: string;
}

export interface KeyModelUsageStat {
  public_model_name: string;
  request_count: number;
  prompt_tokens: number;
  completion_tokens: number;
  cache_creation_input_tokens: number;
  cache_read_input_tokens: number;
  reasoning_tokens: number;
  tool_tokens: number;
  total_tokens: number;
  estimated_cost: number;
}

export interface RequestLog {
  id: number;
  request_id: string;
  trace_id: string;
  user_id?: number;
  api_key_id?: number;
  model_id?: number;
  public_model_name: string;
  provider_id?: number;
  provider_model_id?: number;
  adapter_type: string;
  upstream_model: string;
  request_type: string;
  stream: boolean;
  client_ip: string;
  request_method: string;
  request_path: string;
  http_status: number;
  success: boolean;
  latency_ms: number;
  prompt_tokens: number;
  completion_tokens: number;
  cache_creation_input_tokens: number;
  cache_read_input_tokens: number;
  reasoning_tokens: number;
  tool_tokens: number;
  total_tokens: number;
  estimated_cost: number;
  error_type: string;
  error_message: string;
  request_preview: unknown;
  response_preview: unknown;
  metadata: unknown;
  created_at: string;
  user?: User;
}
