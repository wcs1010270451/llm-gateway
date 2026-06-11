package service

import (
	"encoding/json"
	"strings"
	"testing"
)

// 参考向量由 ../claude-max-proxy/.venv 的 Python xxhash 生成：
// xxhash.xxh64(data, seed=0x6E52736AC806831E).intdigest()
// 两边必须完全一致，否则 cch 签名不匹配、Claude Code 伪装失效。
func TestXXHash64MatchesPython(t *testing.T) {
	cases := []struct {
		in   string
		full uint64
		cch  string
	}{
		{"", 7960540557538117613, "647ed"},
		{"hello", 10784765851560677930, "9922a"},
		{`{"model":"claude"}`, 11272023490732157612, "2eaac"},
		{"the quick brown fox", 14328597750081404361, "03dc9"},
	}
	for _, tc := range cases {
		got := xxhash64([]byte(tc.in), cchSeed)
		if got != tc.full {
			t.Errorf("xxhash64(%q): got %d, want %d", tc.in, got, tc.full)
		}
		if cch := computeCCH([]byte(tc.in)); cch != tc.cch {
			t.Errorf("computeCCH(%q): got %s, want %s", tc.in, cch, tc.cch)
		}
	}
}

// cch 必须在同一会话内（system/tools 不变、仅 messages 增长时）保持稳定，
// 否则缓存命不中，是 proxy.py 烧额度的根因。这条测试守住这个不变量。
func TestCCHStableAcrossMessages(t *testing.T) {
	build := func(userMsg string) map[string]any {
		raw := `{"model":"m","system":[{"type":"text","text":"sys prompt"}],` +
			`"tools":[{"name":"Read"},{"name":"Bash"}],` +
			`"messages":[{"role":"user","content":"` + userMsg + `"}]}`
		var p map[string]any
		if err := json.Unmarshal([]byte(raw), &p); err != nil {
			t.Fatal(err)
		}
		return p
	}

	b1, err := injectClaudeSystemAndCCH(build("hi"), "2.1.133.190")
	if err != nil {
		t.Fatal(err)
	}
	b2, err := injectClaudeSystemAndCCH(build("a much longer and totally different follow-up question"), "2.1.133.190")
	if err != nil {
		t.Fatal(err)
	}
	if c1, c2 := billingCCH(t, b1), billingCCH(t, b2); c1 != c2 {
		t.Errorf("cch must stay stable as messages grow: %q vs %q", c1, c2)
	}

	// system 变了，cch 应当随之变化（缓存正确分桶）。
	var changed map[string]any
	_ = json.Unmarshal([]byte(`{"model":"m","system":[{"type":"text","text":"DIFFERENT sys"}],`+
		`"tools":[{"name":"Read"},{"name":"Bash"}],"messages":[{"role":"user","content":"hi"}]}`), &changed)
	b3, _ := injectClaudeSystemAndCCH(changed, "2.1.133.190")
	if billingCCH(t, b1) == billingCCH(t, b3) {
		t.Errorf("cch should change when system changes")
	}
}

// billingCCH 从最终请求体里抽出 billing 块的 cch 值。
func billingCCH(t *testing.T, body []byte) string {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatal(err)
	}
	text := out["system"].([]any)[0].(map[string]any)["text"].(string)
	idx := strings.Index(text, "cch=")
	if idx < 0 {
		t.Fatalf("no cch in billing: %q", text)
	}
	return text[idx : idx+len("cch=")+5]
}

func TestInjectClaudeSystemAndCCH(t *testing.T) {
	raw := `{
		"model": "claude-sonnet-4-6",
		"system": [{"type": "text", "text": "You are OpenClaw."}],
		"messages": [{"role": "user", "content": "hello"}]
	}`
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatal(err)
	}

	body, err := injectClaudeSystemAndCCH(payload, "2.1.133.190")
	if err != nil {
		t.Fatal(err)
	}

	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	// system 应只剩 billing + identity 两块，且 billing 带真实 cch（非占位）。
	system, ok := out["system"].([]any)
	if !ok || len(system) != 2 {
		t.Fatalf("expected 2 system blocks, got %#v", out["system"])
	}
	billing := system[0].(map[string]any)["text"].(string)
	if !strings.Contains(billing, "cc_version=2.1.133.190") {
		t.Errorf("billing header missing version: %q", billing)
	}
	if strings.Contains(billing, "cch=00000") || !strings.Contains(billing, "cch=") {
		t.Errorf("cch not substituted: %q", billing)
	}

	// 原始 system 文本应迁移进首条 user message 的 <system_instructions> 块。
	messages := out["messages"].([]any)
	firstContent := messages[0].(map[string]any)["content"].([]any)
	prefix := firstContent[0].(map[string]any)["text"].(string)
	if !strings.Contains(prefix, "<system_instructions>") || !strings.Contains(prefix, "You are OpenClaw.") {
		t.Errorf("system not migrated into user message: %q", prefix)
	}

	// 尖括号不能被 HTML 转义：最终字节里应保留原始 <system_instructions>。
	if !strings.Contains(string(body), "<system_instructions>") {
		t.Errorf("angle brackets were HTML-escaped, raw tag missing: %s", body)
	}
}
