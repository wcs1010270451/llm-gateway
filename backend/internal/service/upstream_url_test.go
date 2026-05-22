package service

import "testing"

func TestBuildUpstreamURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		endpoint string
		want     string
	}{
		{
			name:     "base without version",
			baseURL:  "https://api.example.com",
			endpoint: "/v1/chat/completions",
			want:     "https://api.example.com/v1/chat/completions",
		},
		{
			name:     "base already ends with v1",
			baseURL:  "https://dashscope.aliyuncs.com/compatible-mode/v1",
			endpoint: "/v1/chat/completions",
			want:     "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
		},
		{
			name:     "base already includes endpoint",
			baseURL:  "https://api.example.com/v1/messages",
			endpoint: "/v1/messages",
			want:     "https://api.example.com/v1/messages",
		},
		{
			name:     "base already ends with v1beta",
			baseURL:  "https://generativelanguage.googleapis.com/v1beta",
			endpoint: "/v1beta/models/gemini-2.5-flash:generateContent",
			want:     "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildUpstreamURL(tt.baseURL, tt.endpoint)
			if got != tt.want {
				t.Fatalf("buildUpstreamURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
