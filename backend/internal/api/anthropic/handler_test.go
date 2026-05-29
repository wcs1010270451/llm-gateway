package anthropic

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAPIKeyFromRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		headers http.Header
		want    string
	}{
		{
			name:    "uses x api key",
			headers: http.Header{"X-Api-Key": []string{"key-x"}},
			want:    "key-x",
		},
		{
			name:    "uses bearer token",
			headers: http.Header{"Authorization": []string{"Bearer key-bearer"}},
			want:    "key-bearer",
		},
		{
			name: "prefers x api key",
			headers: http.Header{
				"X-Api-Key":     []string{"key-x"},
				"Authorization": []string{"Bearer key-bearer"},
			},
			want: "key-x",
		},
		{
			name:    "ignores unsupported authorization",
			headers: http.Header{"Authorization": []string{"Basic value"}},
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
			request.Header = tt.headers
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = request

			if got := apiKeyFromRequest(context); got != tt.want {
				t.Fatalf("apiKeyFromRequest() = %q, want %q", got, tt.want)
			}
		})
	}
}
