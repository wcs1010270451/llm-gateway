import { apiClient } from "./http";
import type {
  ListResponse,
  KeyModelUsageStat,
  Model,
  ModelFamily,
  ModelFamilyInput,
  ModelInput,
  Provider,
  ProviderInput,
  ProviderModel,
  ProviderModelInput,
  RequestLog,
  User,
  UserInput,
  PageResponse,
} from "../types";

export interface AdminStats {
  provider_count: number;
  model_count: number;
  user_count: number;
  active_user_count: number;
  api_key_count: number;
  recent_usage: {
    request_count: number;
    success_count: number;
    active_user_count: number;
    active_key_count: number;
    total_tokens: number;
    average_latency_ms: number;
    estimated_cost: number;
  };
  top_models: KeyModelUsageStat[];
}

export interface ClaudeProxyStatus {
  provider_id: number;
  name: string;
  slug: string;
  status: Provider["status"];
  base_url: string;
  reachable: boolean;
  proxy_status: string;
  token_hours?: number;
  cc_version?: string;
  subscription_type?: string;
  rate_limit_tier?: string;
  http_status?: number;
  error?: string;
  checked_at: string;
}

export interface ClaudeProxyProbeResult extends ClaudeProxyStatus {
  probe_ok: boolean;
}

export interface ClaudeProxyRefreshResult extends ClaudeProxyStatus {
  refresh_ok: boolean;
}

export interface ProviderUsageStats {
  window: "24h" | "7d" | "30d";
  granularity: "hour" | "day";
  since: string;
  summary: {
    request_count: number;
    success_count: number;
    active_user_count: number;
    active_key_count: number;
    total_tokens: number;
    average_latency_ms: number;
    estimated_cost: number;
  };
  trend: Array<{
    period: string;
    request_count: number;
    total_tokens: number;
    estimated_cost: number;
  }>;
  models: Array<{
    provider_model_id?: number;
    upstream_model: string;
    public_model_name: string;
    request_count: number;
    success_count: number;
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
    estimated_cost: number;
    average_latency_ms: number;
    last_used_at?: string;
  }>;
}

export async function fetchAdminStats() {
  const response = await apiClient.get<AdminStats>("/api/admin/stats");
  return response.data;
}

export async function fetchClaudeProxyStatuses() {
  const response = await apiClient.get<ListResponse<ClaudeProxyStatus>>("/api/admin/claude-proxies");
  return response.data;
}

export async function probeClaudeProxy(id: number) {
  const response = await apiClient.post<ClaudeProxyProbeResult>(`/api/admin/claude-proxies/${id}/probe`);
  return response.data;
}

export async function refreshClaudeProxyToken(id: number) {
  const response = await apiClient.post<ClaudeProxyRefreshResult>(`/api/admin/claude-proxies/${id}/refresh`);
  return response.data;
}

export async function fetchProviders() {
  const response = await apiClient.get<ListResponse<Provider>>("/api/admin/providers");
  return response.data;
}

export async function fetchProvider(id: string) {
  const response = await apiClient.get<Provider>(`/api/admin/providers/${id}`);
  return response.data;
}

export async function fetchProviderModelRoutes(providerId: string) {
  const response = await apiClient.get<ListResponse<ProviderModel>>(
    `/api/admin/providers/${providerId}/provider-models`,
  );
  return response.data;
}

export async function fetchProviderUsageStats(
  providerId: string,
  params: { window: "24h" | "7d" | "30d"; granularity?: "hour" | "day" },
) {
  const response = await apiClient.get<ProviderUsageStats>(`/api/admin/providers/${providerId}/usage-stats`, {
    params,
  });
  return response.data;
}

export async function createProviderModelRoute(providerId: number, data: ProviderModelInput) {
  const response = await apiClient.post<ProviderModel>(`/api/admin/providers/${providerId}/provider-models`, data);
  return response.data;
}

export async function updateProviderModelRoute(providerId: number, providerModelId: number, data: ProviderModelInput) {
  const response = await apiClient.put<ProviderModel>(
    `/api/admin/providers/${providerId}/provider-models/${providerModelId}`,
    data,
  );
  return response.data;
}

export async function deleteProviderModelRoute(providerId: number, providerModelId: number) {
  await apiClient.delete(`/api/admin/providers/${providerId}/provider-models/${providerModelId}`);
}

export async function createProvider(data: ProviderInput) {
  const response = await apiClient.post<Provider>("/api/admin/providers", data);
  return response.data;
}

export async function updateProvider(id: number, data: ProviderInput) {
  const response = await apiClient.put<Provider>(`/api/admin/providers/${id}`, data);
  return response.data;
}

export async function deleteProvider(id: number) {
  await apiClient.delete(`/api/admin/providers/${id}`);
}

export async function fetchModels() {
  const response = await apiClient.get<ListResponse<Model>>("/api/admin/models");
  return response.data;
}

export async function fetchModelsByFamily(family: string) {
  const response = await apiClient.get<ListResponse<Model>>("/api/admin/models", { params: { family } });
  return response.data;
}

export async function fetchModelPage(params: { page: number; page_size: number; family?: string }) {
  const response = await apiClient.get<PageResponse<Model>>("/api/admin/models", { params });
  return response.data;
}

export async function fetchModelFamilies() {
  const response = await apiClient.get<ListResponse<ModelFamily>>("/api/admin/model-families");
  return response.data;
}

export async function fetchActiveModelFamilies() {
  const response = await apiClient.get<ListResponse<ModelFamily>>("/api/admin/model-families/active");
  return response.data;
}

export async function createModelFamily(data: ModelFamilyInput) {
  const response = await apiClient.post<ModelFamily>("/api/admin/model-families", data);
  return response.data;
}

export async function updateModelFamily(id: number, data: ModelFamilyInput) {
  const response = await apiClient.put<ModelFamily>(`/api/admin/model-families/${id}`, data);
  return response.data;
}

export async function deleteModelFamily(id: number) {
  await apiClient.delete(`/api/admin/model-families/${id}`);
}

export async function fetchModel(id: string) {
  const response = await apiClient.get<Model>(`/api/admin/models/${id}`);
  return response.data;
}

export async function createModel(data: ModelInput) {
  const response = await apiClient.post<Model>("/api/admin/models", data);
  return response.data;
}

export async function updateModel(id: number, data: ModelInput) {
  const response = await apiClient.put<Model>(`/api/admin/models/${id}`, data);
  return response.data;
}

export async function deleteModel(id: number) {
  await apiClient.delete(`/api/admin/models/${id}`);
}

export async function fetchProviderModels(modelId: string) {
  const response = await apiClient.get<ListResponse<ProviderModel>>(`/api/admin/models/${modelId}/provider-models`);
  return response.data;
}

export async function createProviderModel(modelId: number, data: ProviderModelInput) {
  const response = await apiClient.post<ProviderModel>(`/api/admin/models/${modelId}/provider-models`, data);
  return response.data;
}

export async function updateProviderModel(modelId: number, providerModelId: number, data: ProviderModelInput) {
  const response = await apiClient.put<ProviderModel>(
    `/api/admin/models/${modelId}/provider-models/${providerModelId}`,
    data,
  );
  return response.data;
}

export async function deleteProviderModel(modelId: number, providerModelId: number) {
  await apiClient.delete(`/api/admin/models/${modelId}/provider-models/${providerModelId}`);
}

export async function setActiveProviderModel(modelId: number, providerModelId: number) {
  const response = await apiClient.post<Model>(`/api/admin/models/${modelId}/provider-models/${providerModelId}/activate`);
  return response.data;
}

export async function fetchUsers() {
  const response = await apiClient.get<ListResponse<User>>("/api/admin/users");
  return response.data;
}

export async function createUser(data: UserInput) {
  const response = await apiClient.post<User>("/api/admin/users", data);
  return response.data;
}

export async function updateUser(id: number, data: UserInput) {
  const response = await apiClient.put<User>(`/api/admin/users/${id}`, data);
  return response.data;
}

export async function deleteUser(id: number) {
  await apiClient.delete(`/api/admin/users/${id}`);
}

export async function fetchLogs(params: { page: number; page_size: number }) {
  const response = await apiClient.get<PageResponse<RequestLog>>("/api/admin/logs", { params });
  return response.data;
}

export async function fetchLog(id: number) {
  const response = await apiClient.get<RequestLog>(`/api/admin/logs/${id}`);
  return response.data;
}
