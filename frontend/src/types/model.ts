import type { Provider } from "./provider";

export interface Model {
  id: number;
  name: string;
  display_name: string;
  family: string;
  modality: "text" | "vision" | "embedding" | "multimodal";
  status: "enabled" | "disabled";
  active_provider_model_id?: number;
  description: string;
  pricing_json: Record<string, unknown>;
  config_json: Record<string, unknown>;
  active_provider_model?: ProviderModel;
  created_at: string;
  updated_at: string;
}

export interface ModelInput {
  name: string;
  display_name: string;
  family: string;
  modality: Model["modality"];
  status: Model["status"];
  description: string;
  pricing_json: Record<string, unknown>;
  config_json: Record<string, unknown>;
}

export interface ProviderModel {
  id: number;
  provider_id: number;
  model_id?: number;
  upstream_model: string;
  status: "enabled" | "disabled";
  max_tokens: number;
  timeout_seconds: number;
  input_cost_per_1m: number;
  output_cost_per_1m: number;
  pricing_json: Record<string, unknown>;
  config_json: Record<string, unknown>;
  provider?: Provider;
  model?: Model;
  created_at: string;
  updated_at: string;
}

export interface ProviderModelInput {
  provider_id: number;
  model_id?: number;
  upstream_model: string;
  status: ProviderModel["status"];
  max_tokens: number;
  timeout_seconds: number;
  input_cost_per_1m: number;
  output_cost_per_1m: number;
  pricing_json: Record<string, unknown>;
  config_json: Record<string, unknown>;
  set_active: boolean;
}
