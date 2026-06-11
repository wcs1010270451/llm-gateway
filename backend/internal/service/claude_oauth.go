package service

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/bits"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// owner_claude_max provider：用 owner 自有的 Claude Max 订阅直连 api.anthropic.com。
// 请求被伪装成 Claude Code CLI 发出（system 迁移 + cch 签名 + OAuth Bearer），
// 从而走订阅额度而非 API credits。Token 由宿主机上的定时任务刷新写入凭证文件，
// 本服务只负责读取最新 accessToken，不实现 OAuth 刷新流程。
//
// 这是 ../claude-max-proxy/proxy.py 的 Go 移植；行为需与之保持一致。

const (
	defaultAnthropicBaseURL = "https://api.anthropic.com"
	// 目录式挂载下的凭证文件路径，详见 docker-compose.prod.yml 与 README。
	defaultClaudeCredsFile = "/secrets/claude/.credentials.json"
	defaultClaudeCCVersion  = "2.1.133"
	defaultClaudeCCBuild    = "190"
	defaultClaudeBeta       = "claude-code-20250219,oauth-2025-04-20,interleaved-thinking-2025-05-14,context-1m-2025-08-07,context-management-2025-06-27,prompt-caching-scope-2026-01-05,effort-2025-11-24"

	// 与 proxy.py 的 CCH_SEED 保持一致，改动会导致 cch 不匹配、伪装失效。
	cchSeed uint64 = 0x6E52736AC806831E
)

type claudeOAuthConfig struct {
	CredentialsFile string
	CCVersion       string
	CCBuild         string
	Beta            string
}

// fullVersion 对应 proxy.py 的 CC_FULL_VERSION（四段），用于 billing header。
func (c claudeOAuthConfig) fullVersion() string {
	return c.CCVersion + "." + c.CCBuild
}

func parseClaudeOAuthConfig(raw datatypes.JSON) claudeOAuthConfig {
	cfg := claudeOAuthConfig{
		CredentialsFile: defaultClaudeCredsFile,
		CCVersion:       defaultClaudeCCVersion,
		CCBuild:         defaultClaudeCCBuild,
		Beta:            defaultClaudeBeta,
	}
	if len(raw) == 0 {
		return cfg
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return cfg
	}
	if v, ok := m["credentials_file"].(string); ok && strings.TrimSpace(v) != "" {
		cfg.CredentialsFile = strings.TrimSpace(v)
	}
	if v, ok := m["cc_version"].(string); ok && strings.TrimSpace(v) != "" {
		cfg.CCVersion = strings.TrimSpace(v)
	}
	if v, ok := m["cc_build"].(string); ok && strings.TrimSpace(v) != "" {
		cfg.CCBuild = strings.TrimSpace(v)
	}
	if v, ok := m["beta"].(string); ok && strings.TrimSpace(v) != "" {
		cfg.Beta = strings.TrimSpace(v)
	}
	return cfg
}

// loadClaudeAccessToken 读取凭证文件，取出当前 accessToken。
// 文件由定时任务刷新，本函数每次现读以拿到最新 token。
func loadClaudeAccessToken(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read credentials file %q: %w", path, err)
	}
	var creds struct {
		ClaudeAiOauth struct {
			AccessToken string `json:"accessToken"`
			ExpiresAt   int64  `json:"expiresAt"`
		} `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", fmt.Errorf("parse credentials file %q: %w", path, err)
	}
	token := strings.TrimSpace(creds.ClaudeAiOauth.AccessToken)
	if token == "" {
		return "", fmt.Errorf("accessToken missing in credentials file %q", path)
	}
	return token, nil
}

// injectClaudeSystemAndCCH 移植自 proxy.py 的 inject_system_and_cch：
// 把客户端原始 system prompt 迁移进第一条 user message（包在 <system_instructions> 里），
// system 只保留标准 Claude Code 的 billing + identity 两块，再计算并写入 cch 签名。
// 入参 payload 会被就地修改；返回最终要发往上游的请求体字节。
func injectClaudeSystemAndCCH(payload map[string]any, fullVersion string) ([]byte, error) {
	sysTexts := collectSystemTexts(payload["system"])
	combined := strings.Join(sysTexts, "\n\n")

	if len(sysTexts) > 0 {
		prefix := "<system_instructions>\n" + combined + "\n</system_instructions>\n\n"
		prefixBlock := map[string]any{
			"type":          "text",
			"text":          prefix,
			"cache_control": map[string]any{"type": "ephemeral", "ttl": "1h"},
		}
		injectIntoFirstUserMessage(payload, prefixBlock)
	}

	// cch 必须在同一会话内保持稳定。它落在缓存前缀里（billing 块在带 cache_control 的
	// identity 块之前），一旦每轮都变，prompt 缓存就永远命不中，导致每轮把整个
	// system+tools 前缀按全额 input 重复计费——这正是独立 proxy.py 消耗异常快的根因。
	// 因此只对会话内稳定的内容（tools + system 文本 + 版本号）取哈希，绝不掺入会增长的 messages。
	cch := computeCCH(cchStableInput(payload["tools"], combined, fullVersion))

	billing := map[string]any{
		"type": "text",
		"text": "x-anthropic-billing-header: cc_version=" + fullVersion + "; cc_entrypoint=sdk-cli; cch=" + cch + ";",
	}
	identity := map[string]any{
		"type":          "text",
		"text":          "You are Claude Code, Anthropic's official CLI for Claude.",
		"cache_control": map[string]any{"type": "ephemeral", "ttl": "1h"},
	}
	payload["system"] = []any{billing, identity}

	return marshalNoEscape(payload)
}

// cchStableInput 拼出用于计算 cch 的稳定输入：tools + 原始 system 文本 + 版本号。
// 这些在一个会话内不变，从而保证 cch 稳定、缓存可命中。
func cchStableInput(tools any, combinedSystem string, fullVersion string) []byte {
	var buf bytes.Buffer
	if tools != nil {
		if tb, err := marshalNoEscape(tools); err == nil {
			buf.Write(tb)
		}
	}
	buf.WriteString(combinedSystem)
	buf.WriteString(fullVersion)
	return buf.Bytes()
}

// collectSystemTexts 把 system 字段（字符串或块数组）抽取成文本片段列表。
func collectSystemTexts(system any) []string {
	switch v := system.(type) {
	case string:
		if strings.TrimSpace(v) != "" {
			return []string{v}
		}
	case []any:
		texts := make([]string, 0, len(v))
		for _, block := range v {
			switch b := block.(type) {
			case map[string]any:
				if t, ok := b["text"].(string); ok && t != "" {
					texts = append(texts, t)
				}
			case string:
				if b != "" {
					texts = append(texts, b)
				}
			}
		}
		return texts
	}
	return nil
}

// injectIntoFirstUserMessage 把 prefixBlock 插到第一条 user message 内容的最前面。
func injectIntoFirstUserMessage(payload map[string]any, prefixBlock map[string]any) {
	messages, ok := payload["messages"].([]any)
	if !ok {
		return
	}
	for _, raw := range messages {
		msg, ok := raw.(map[string]any)
		if !ok || msg["role"] != "user" {
			continue
		}
		switch content := msg["content"].(type) {
		case string:
			msg["content"] = []any{
				prefixBlock,
				map[string]any{"type": "text", "text": content},
			}
		case []any:
			msg["content"] = append([]any{prefixBlock}, content...)
		default:
			msg["content"] = []any{prefixBlock}
		}
		return
	}
}

// applyClaudeOAuthHeaders 移植自 proxy.py 的 build_headers，把请求伪装成 Claude Code CLI。
func applyClaudeOAuthHeaders(h http.Header, token string, cfg claudeOAuthConfig, stream bool) {
	accept := "application/json"
	if stream {
		accept = "text/event-stream"
	}
	h.Set("Accept", accept)
	h.Set("Authorization", "Bearer "+token)
	h.Set("Content-Type", "application/json")
	h.Set("User-Agent", "claude-cli/"+cfg.CCVersion+" (external, cli)")
	h.Set("X-Claude-Code-Session-Id", uuid.NewString())
	h.Set("x-app", "cli")
	h.Set("anthropic-dangerous-direct-browser-access", "true")
	h.Set("anthropic-beta", cfg.Beta)
	h.Set("anthropic-version", "2023-06-01")
	h.Set("X-Stainless-Lang", "js")
	h.Set("X-Stainless-Package-Version", "0.80.0")
	h.Set("X-Stainless-OS", "Linux")
	h.Set("X-Stainless-Arch", "x64")
	h.Set("X-Stainless-Runtime", "node")
	h.Set("X-Stainless-Runtime-Version", "v24.3.0")
}

// marshalNoEscape 序列化为紧凑 JSON 且不对 <、>、& 做 HTML 转义，
// 与 proxy.py 的 json.dumps(separators=(",",":"), ensure_ascii=False) 对齐，
// 避免 <system_instructions> 的尖括号被转义。
func marshalNoEscape(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	// Encoder 会在末尾追加一个换行，去掉以保持紧凑。
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// computeCCH 复刻 proxy.py：xxh64(body, seed) 取低 20 位，格式化为 5 位十六进制。
func computeCCH(body []byte) string {
	h := xxhash64(body, cchSeed)
	return fmt.Sprintf("%05x", h&0xFFFFF)
}

// ---- 内联 xxHash64（canonical 实现，避免引入外部依赖）----
// 已用 Python xxhash 的参考向量在 claude_oauth_test.go 中对拍验证。

const (
	xxPrime64_1 uint64 = 11400714785074694791
	xxPrime64_2 uint64 = 14029467366897019727
	xxPrime64_3 uint64 = 1609587929392839161
	xxPrime64_4 uint64 = 9650029242287828579
	xxPrime64_5 uint64 = 2870177450012600261
)

func xxRound(acc, input uint64) uint64 {
	acc += input * xxPrime64_2
	acc = bits.RotateLeft64(acc, 31)
	acc *= xxPrime64_1
	return acc
}

func xxMergeRound(acc, val uint64) uint64 {
	val = xxRound(0, val)
	acc ^= val
	acc = acc*xxPrime64_1 + xxPrime64_4
	return acc
}

func xxhash64(data []byte, seed uint64) uint64 {
	n := len(data)
	var h64 uint64
	i := 0

	if n >= 32 {
		v1 := seed + xxPrime64_1 + xxPrime64_2
		v2 := seed + xxPrime64_2
		v3 := seed
		v4 := seed - xxPrime64_1
		for ; i+32 <= n; i += 32 {
			v1 = xxRound(v1, binary.LittleEndian.Uint64(data[i:]))
			v2 = xxRound(v2, binary.LittleEndian.Uint64(data[i+8:]))
			v3 = xxRound(v3, binary.LittleEndian.Uint64(data[i+16:]))
			v4 = xxRound(v4, binary.LittleEndian.Uint64(data[i+24:]))
		}
		h64 = bits.RotateLeft64(v1, 1) + bits.RotateLeft64(v2, 7) +
			bits.RotateLeft64(v3, 12) + bits.RotateLeft64(v4, 18)
		h64 = xxMergeRound(h64, v1)
		h64 = xxMergeRound(h64, v2)
		h64 = xxMergeRound(h64, v3)
		h64 = xxMergeRound(h64, v4)
	} else {
		h64 = seed + xxPrime64_5
	}

	h64 += uint64(n)

	for ; i+8 <= n; i += 8 {
		k1 := xxRound(0, binary.LittleEndian.Uint64(data[i:]))
		h64 ^= k1
		h64 = bits.RotateLeft64(h64, 27)*xxPrime64_1 + xxPrime64_4
	}
	if i+4 <= n {
		h64 ^= uint64(binary.LittleEndian.Uint32(data[i:])) * xxPrime64_1
		h64 = bits.RotateLeft64(h64, 23)*xxPrime64_2 + xxPrime64_3
		i += 4
	}
	for ; i < n; i++ {
		h64 ^= uint64(data[i]) * xxPrime64_5
		h64 = bits.RotateLeft64(h64, 11) * xxPrime64_1
	}

	h64 ^= h64 >> 33
	h64 *= xxPrime64_2
	h64 ^= h64 >> 29
	h64 *= xxPrime64_3
	h64 ^= h64 >> 32
	return h64
}
