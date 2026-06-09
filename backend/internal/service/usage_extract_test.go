package service

import "testing"

func TestExtractOpenAIUsageFromStream(t *testing.T) {
	body := []byte("data: {\"choices\":[{\"delta\":{\"content\":\"hi\"}}]}\n\n" +
		"data: {\"choices\":[],\"usage\":{\"prompt_tokens\":12,\"completion_tokens\":8,\"total_tokens\":20}}\n\n" +
		"data: [DONE]\n\n")

	usage := extractOpenAIUsage(body)
	if usage.prompt != 12 || usage.completion != 8 || usage.total != 20 {
		t.Fatalf("extractOpenAIUsage() = (%d, %d, %d), want (12, 8, 20)", usage.prompt, usage.completion, usage.total)
	}
}

func TestExtractOpenAIUsageDetails(t *testing.T) {
	body := []byte(`{"usage":{"prompt_tokens":120,"completion_tokens":40,"total_tokens":160,"prompt_tokens_details":{"cached_tokens":80},"completion_tokens_details":{"reasoning_tokens":12}}}`)

	usage := extractOpenAIUsage(body)
	if usage.prompt != 120 || usage.completion != 40 || usage.cachedPrompt != 80 || usage.reasoning != 12 || usage.total != 160 {
		t.Fatalf("extractOpenAIUsage() = %+v, want prompt=120 completion=40 cached=80 reasoning=12 total=160", usage)
	}
}

func TestEnsureOpenAIStreamUsage(t *testing.T) {
	payload := map[string]any{
		"stream_options": map[string]any{
			"existing": true,
		},
	}

	ensureOpenAIStreamUsage(payload)

	streamOptions, ok := payload["stream_options"].(map[string]any)
	if !ok {
		t.Fatal("stream_options was not preserved as an object")
	}
	if streamOptions["existing"] != true {
		t.Fatal("existing stream_options field was not preserved")
	}
	if streamOptions["include_usage"] != true {
		t.Fatal("include_usage was not enabled")
	}
}

func TestExtractAnthropicUsageFromStream(t *testing.T) {
	body := []byte("event: message_start\n" +
		"data: {\"type\":\"message_start\",\"message\":{\"usage\":{\"input_tokens\":31,\"output_tokens\":1}}}\n\n" +
		"event: message_delta\n" +
		"data: {\"type\":\"message_delta\",\"usage\":{\"output_tokens\":17}}\n\n")

	usage := extractAnthropicUsage(body)
	if usage.input != 31 || usage.output != 17 {
		t.Fatalf("extractAnthropicUsage() = (%d, %d), want (31, 17)", usage.input, usage.output)
	}
}

func TestExtractAnthropicUsageCacheTokens(t *testing.T) {
	body := []byte("event: message_start\n" +
		"data: {\"type\":\"message_start\",\"message\":{\"usage\":{\"input_tokens\":504,\"cache_creation_input_tokens\":213,\"cache_read_input_tokens\":49691,\"output_tokens\":3}}}\n\n" +
		"event: message_delta\n" +
		"data: {\"type\":\"message_delta\",\"usage\":{\"output_tokens\":11}}\n\n")

	usage := extractAnthropicUsage(body)
	if usage.input != 504 || usage.output != 11 || usage.cacheCreation != 213 || usage.cacheRead != 49691 || usage.total() != 50419 {
		t.Fatalf("extractAnthropicUsage() = %+v, total=%d; want input=504 output=11 cacheCreation=213 cacheRead=49691 total=50419", usage, usage.total())
	}
}

func TestExtractGeminiUsageFromStream(t *testing.T) {
	body := []byte("data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"hi\"}]}}]}\n\n" +
		"data: {\"usageMetadata\":{\"promptTokenCount\":11,\"candidatesTokenCount\":9,\"totalTokenCount\":20}}\n\n")

	usage := extractGeminiUsage(body)
	if usage.prompt != 11 || usage.completion != 9 || usage.total != 20 {
		t.Fatalf("extractGeminiUsage() = (%d, %d, %d), want (11, 9, 20)", usage.prompt, usage.completion, usage.total)
	}
}

func TestExtractGeminiUsageDetails(t *testing.T) {
	body := []byte(`{"usageMetadata":{"promptTokenCount":30,"candidatesTokenCount":12,"cachedContentTokenCount":20,"thoughtsTokenCount":7,"toolUsePromptTokenCount":3,"totalTokenCount":52}}`)

	usage := extractGeminiUsage(body)
	if usage.prompt != 30 || usage.completion != 12 || usage.cached != 20 || usage.thoughts != 7 || usage.tool != 3 || usage.total != 52 {
		t.Fatalf("extractGeminiUsage() = %+v, want prompt=30 completion=12 cached=20 thoughts=7 tool=3 total=52", usage)
	}
}

