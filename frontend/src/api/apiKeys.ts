import { useAuthStore } from "../store/authStore";
import type {
  APIKey,
  APIKeyInput,
  CreatedAPIKey,
  RevealedAPIKey,
  KeyModelUsageStat,
  ListResponse,
  Model,
  PageResponse,
  RequestLog,
} from "../types";
import { apiClient } from "./http";

export async function fetchMyAPIKeys() {
  const response = await apiClient.get<ListResponse<APIKey>>("/api/me/api-keys");
  return response.data;
}

export async function fetchMyAPIKey(id: number) {
  const response = await apiClient.get<APIKey>(`/api/me/api-keys/${id}`);
  return response.data;
}

export async function fetchMyModels() {
  const response = await apiClient.get<ListResponse<Model>>("/api/me/models");
  return response.data;
}

export async function fetchMyAPIKeyModelStats(id: number) {
  const response = await apiClient.get<ListResponse<KeyModelUsageStat>>(`/api/me/api-keys/${id}/model-stats`);
  return response.data;
}

export async function fetchMyAPIKeyLogs(id: number, params: { page: number; page_size: number }) {
  const response = await apiClient.get<PageResponse<RequestLog>>(`/api/me/api-keys/${id}/logs`, { params });
  return response.data;
}

export async function fetchMyLog(id: number) {
  const response = await apiClient.get<RequestLog>(`/api/me/logs/${id}`);
  return response.data;
}

export async function createMyAPIKey(data: APIKeyInput) {
  const response = await apiClient.post<CreatedAPIKey>("/api/me/api-keys", data);
  return response.data;
}

export async function updateMyAPIKey(id: number, data: APIKeyInput) {
  const response = await apiClient.put<APIKey>(`/api/me/api-keys/${id}`, data);
  return response.data;
}

export async function deleteMyAPIKey(id: number) {
  await apiClient.delete(`/api/me/api-keys/${id}`);
}

export async function revealMyAPIKey(id: number) {
  const response = await apiClient.post<RevealedAPIKey>(`/api/me/api-keys/${id}/reveal`);
  return response.data;
}

export interface DebugMessagesPayload {
  model: string;
  max_tokens: number;
  stream: boolean;
  messages: Array<{ role: "user" | "assistant"; content: string }>;
}

export interface DebugMessagesResult {
  data?: unknown;
  status?: number;
  assistantText: string;
  rawText: string;
  usage?: {
    prompt_tokens: number;
    completion_tokens: number;
    cache_creation_input_tokens: number;
    cache_read_input_tokens: number;
    reasoning_tokens: number;
    tool_tokens: number;
    total_tokens: number;
  } | null;
  error?: string;
}

export async function sendMyDebugMessages(
  keyId: number,
  payload: DebugMessagesPayload,
  options: { signal?: AbortSignal; onUpdate?: (result: DebugMessagesResult) => void } = {},
) {
  const token = useAuthStore.getState().token;
  const baseURL = apiClient.defaults.baseURL ?? "";
  const response = await fetch(`${baseURL}/api/me/api-keys/${keyId}/debug/messages`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "anthropic-version": "2023-06-01",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(payload),
    signal: options.signal,
  });

  const contentType = response.headers.get("content-type") ?? "";
  if (!contentType.toLowerCase().includes("text/event-stream")) {
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      const message = data?.error?.message ?? `Request failed with status ${response.status}`;
      throw new Error(message);
    }
    const result = parseAnthropicJSON(data, response.status);
    options.onUpdate?.(result);
    return result;
  }

  if (!response.ok || !response.body) {
    throw new Error(`Request failed with status ${response.status}`);
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let assistantText = "";
  let usage: DebugMessagesResult["usage"] = null;

  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }

    buffer += decoder.decode(value, { stream: true });
    const lines = buffer.split(/\r?\n/);
    buffer = lines.pop() ?? "";

    for (const line of lines) {
      if (!line.startsWith("data:")) {
        continue;
      }
      const raw = line.slice(5).trim();
      if (!raw || raw === "[DONE]") {
        continue;
      }

      const event = JSON.parse(raw);
      if (event.type === "content_block_delta" && event.delta?.type === "text_delta") {
        assistantText += event.delta.text ?? "";
      }
      if (event.type === "message_start" && event.message?.usage) {
        const inputTokens: number = event.message.usage.input_tokens ?? 0;
        const outputTokens: number = event.message.usage.output_tokens ?? 0;
        const cacheCreationTokens: number = event.message.usage.cache_creation_input_tokens ?? 0;
        const cacheReadTokens: number = event.message.usage.cache_read_input_tokens ?? 0;
        usage = {
          prompt_tokens: inputTokens,
          completion_tokens: outputTokens,
          cache_creation_input_tokens: cacheCreationTokens,
          cache_read_input_tokens: cacheReadTokens,
          reasoning_tokens: 0,
          tool_tokens: 0,
          total_tokens: inputTokens + outputTokens + cacheCreationTokens + cacheReadTokens,
        };
      }
      if (event.type === "message_delta" && event.usage) {
        const inputTokens: number = usage?.prompt_tokens ?? 0;
        const outputTokens: number = event.usage.output_tokens ?? usage?.completion_tokens ?? 0;
        const cacheCreationTokens: number = event.usage.cache_creation_input_tokens ?? usage?.cache_creation_input_tokens ?? 0;
        const cacheReadTokens: number = event.usage.cache_read_input_tokens ?? usage?.cache_read_input_tokens ?? 0;
        usage = {
          prompt_tokens: inputTokens,
          completion_tokens: outputTokens,
          cache_creation_input_tokens: cacheCreationTokens,
          cache_read_input_tokens: cacheReadTokens,
          reasoning_tokens: 0,
          tool_tokens: 0,
          total_tokens: inputTokens + outputTokens + cacheCreationTokens + cacheReadTokens,
        };
      }
      options.onUpdate?.({ assistantText, rawText: assistantText, usage, status: response.status });
    }
  }

  return { assistantText, rawText: assistantText, usage, status: response.status };
}

export async function sendMyDebugChatCompletions(
  keyId: number,
  payload: DebugMessagesPayload,
  options: { signal?: AbortSignal; onUpdate?: (result: DebugMessagesResult) => void } = {},
) {
  const token = useAuthStore.getState().token;
  const baseURL = apiClient.defaults.baseURL ?? "";
  const response = await fetch(`${baseURL}/api/me/api-keys/${keyId}/debug/chat/completions`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(payload),
    signal: options.signal,
  });

  const contentType = response.headers.get("content-type") ?? "";
  if (!contentType.toLowerCase().includes("text/event-stream")) {
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      const message = data?.error?.message ?? `Request failed with status ${response.status}`;
      throw new Error(message);
    }
    const result = parseOpenAIJSON(data, response.status);
    options.onUpdate?.(result);
    return result;
  }

  if (!response.ok || !response.body) {
    throw new Error(`Request failed with status ${response.status}`);
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let assistantText = "";

  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }

    buffer += decoder.decode(value, { stream: true });
    const lines = buffer.split(/\r?\n/);
    buffer = lines.pop() ?? "";

    for (const line of lines) {
      if (!line.startsWith("data:")) {
        continue;
      }
      const raw = line.slice(5).trim();
      if (!raw || raw === "[DONE]") {
        continue;
      }
      const event = JSON.parse(raw);
      assistantText += event.choices?.[0]?.delta?.content ?? "";
      options.onUpdate?.({ assistantText, rawText: assistantText, usage: null, status: response.status });
    }
  }

  return { assistantText, rawText: assistantText, usage: null, status: response.status };
}

export interface DebugGeminiPayload {
  model: string;
  contents: Array<{ role: "user" | "model"; parts: Array<{ text: string }> }>;
  generationConfig?: {
    maxOutputTokens?: number;
  };
}

export async function sendMyDebugGeminiGenerateContent(
  keyId: number,
  payload: DebugGeminiPayload,
  stream: boolean,
  options: { signal?: AbortSignal; onUpdate?: (result: DebugMessagesResult) => void } = {},
) {
  const token = useAuthStore.getState().token;
  const baseURL = apiClient.defaults.baseURL ?? "";
  const endpoint = stream ? "stream-generate-content" : "generate-content";
  const response = await fetch(`${baseURL}/api/me/api-keys/${keyId}/debug/gemini/${endpoint}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(payload),
    signal: options.signal,
  });

  const contentType = response.headers.get("content-type") ?? "";
  if (!contentType.toLowerCase().includes("text/event-stream")) {
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      const message = data?.error?.message ?? `Request failed with status ${response.status}`;
      throw new Error(message);
    }
    const result = parseGeminiJSON(data, response.status);
    options.onUpdate?.(result);
    return result;
  }

  if (!response.ok || !response.body) {
    throw new Error(`Request failed with status ${response.status}`);
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let assistantText = "";
  let usage: DebugMessagesResult["usage"] = null;

  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }

    buffer += decoder.decode(value, { stream: true });
    const events = buffer.split(/\r?\n\r?\n/);
    buffer = events.pop() ?? "";

    for (const eventText of events) {
      const line = eventText
        .split(/\r?\n/)
        .map((lineItem) => lineItem.trim())
        .find((lineItem) => lineItem.startsWith("data:"));
      if (!line) {
        continue;
      }
      const raw = line.slice(5).trim();
      if (!raw || raw === "[DONE]") {
        continue;
      }
      const event = JSON.parse(raw);
      assistantText += geminiText(event);
      usage = geminiUsage(event) ?? usage;
      options.onUpdate?.({ assistantText, rawText: assistantText, usage, status: response.status });
    }
  }

  return { assistantText, rawText: assistantText, usage, status: response.status };
}

function parseAnthropicJSON(data: any, status: number): DebugMessagesResult {
  const assistantText = Array.isArray(data?.content)
    ? data.content
        .filter((block: any) => block?.type === "text")
        .map((block: any) => block.text ?? "")
        .join("")
    : "";
  const inputTokens = data?.usage?.input_tokens ?? 0;
  const outputTokens = data?.usage?.output_tokens ?? 0;
  const cacheCreationTokens = data?.usage?.cache_creation_input_tokens ?? 0;
  const cacheReadTokens = data?.usage?.cache_read_input_tokens ?? 0;
  return {
    data,
    status,
    assistantText,
    rawText: JSON.stringify(data, null, 2),
    usage: data?.usage
      ? {
          prompt_tokens: inputTokens,
          completion_tokens: outputTokens,
          cache_creation_input_tokens: cacheCreationTokens,
          cache_read_input_tokens: cacheReadTokens,
          reasoning_tokens: 0,
          tool_tokens: 0,
          total_tokens: inputTokens + outputTokens + cacheCreationTokens + cacheReadTokens,
        }
      : null,
  };
}

function parseOpenAIJSON(data: any, status: number): DebugMessagesResult {
  const assistantText = data?.choices?.[0]?.message?.content ?? "";
  const promptTokens = data?.usage?.prompt_tokens ?? 0;
  const completionTokens = data?.usage?.completion_tokens ?? 0;
  const totalTokens = data?.usage?.total_tokens ?? promptTokens + completionTokens;
  const cachedTokens = data?.usage?.prompt_tokens_details?.cached_tokens ?? 0;
  const reasoningTokens = data?.usage?.completion_tokens_details?.reasoning_tokens ?? 0;
  return {
    data,
    status,
    assistantText,
    rawText: JSON.stringify(data, null, 2),
    usage: data?.usage
      ? {
          prompt_tokens: promptTokens,
          completion_tokens: completionTokens,
          cache_creation_input_tokens: 0,
          cache_read_input_tokens: cachedTokens,
          reasoning_tokens: reasoningTokens,
          tool_tokens: 0,
          total_tokens: totalTokens,
        }
      : null,
  };
}

function parseGeminiJSON(data: any, status: number): DebugMessagesResult {
  return {
    data,
    status,
    assistantText: geminiText(data),
    rawText: JSON.stringify(data, null, 2),
    usage: geminiUsage(data),
  };
}

function geminiText(data: any) {
  const parts = data?.candidates?.[0]?.content?.parts;
  if (!Array.isArray(parts)) {
    return "";
  }
  return parts.map((part: any) => part?.text ?? "").join("");
}

function geminiUsage(data: any): DebugMessagesResult["usage"] {
  const usage = data?.usageMetadata;
  if (!usage) {
    return null;
  }
  const promptTokens = usage.promptTokenCount ?? 0;
  const completionTokens = usage.candidatesTokenCount ?? 0;
  const thoughtsTokens = usage.thoughtsTokenCount ?? 0;
  const toolTokens = usage.toolUsePromptTokenCount ?? 0;
  return {
    prompt_tokens: promptTokens,
    completion_tokens: completionTokens,
    cache_creation_input_tokens: 0,
    cache_read_input_tokens: usage.cachedContentTokenCount ?? 0,
    reasoning_tokens: thoughtsTokens,
    tool_tokens: toolTokens,
    total_tokens: usage.totalTokenCount ?? promptTokens + completionTokens + thoughtsTokens + toolTokens,
  };
}
