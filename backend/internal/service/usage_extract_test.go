package service

import "testing"

func TestExtractOpenAIUsageFromStream(t *testing.T) {
	body := []byte("data: {\"choices\":[{\"delta\":{\"content\":\"hi\"}}]}\n\n" +
		"data: {\"choices\":[],\"usage\":{\"prompt_tokens\":12,\"completion_tokens\":8,\"total_tokens\":20}}\n\n" +
		"data: [DONE]\n\n")

	prompt, completion, total := extractOpenAIUsage(body)
	if prompt != 12 || completion != 8 || total != 20 {
		t.Fatalf("extractOpenAIUsage() = (%d, %d, %d), want (12, 8, 20)", prompt, completion, total)
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

	input, output := extractAnthropicUsage(body)
	if input != 31 || output != 17 {
		t.Fatalf("extractAnthropicUsage() = (%d, %d), want (31, 17)", input, output)
	}
}

func TestExtractGeminiUsageFromStream(t *testing.T) {
	body := []byte("data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"hi\"}]}}]}\n\n" +
		"data: {\"usageMetadata\":{\"promptTokenCount\":11,\"candidatesTokenCount\":9,\"totalTokenCount\":20}}\n\n")

	prompt, completion, total := extractGeminiUsage(body)
	if prompt != 11 || completion != 9 || total != 20 {
		t.Fatalf("extractGeminiUsage() = (%d, %d, %d), want (11, 9, 20)", prompt, completion, total)
	}
}
