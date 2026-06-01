package anthropic

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"llm-gateway/backend/internal/service"
)

type Handler struct {
	proxy   *service.AnthropicProxyService
	apiKeys *service.APIKeyService
}

func NewHandler(proxy *service.AnthropicProxyService, apiKeys *service.APIKeyService) *Handler {
	return &Handler{proxy: proxy, apiKeys: apiKeys}
}

func (h *Handler) Messages(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request_error", "failed to read request body")
		return
	}

	apiAuth, err := h.apiKeys.Authenticate(c.Request.Context(), apiKeyFromRequest(c))
	if err != nil {
		handleAPIKeyError(c, err)
		return
	}

	result, err := h.proxy.OpenMessages(c.Request.Context(), service.AnthropicMessagesInput{
		Body:     body,
		Headers:  c.Request.Header,
		Method:   c.Request.Method,
		Path:     c.Request.URL.Path,
		ClientIP: c.ClientIP(),
		UserID:   apiAuth.User.ID,
		APIKeyID: apiAuth.APIKey.ID,
	})
	if err != nil {
		handleProxyError(c, err)
		return
	}
	defer result.Response.Body.Close()

	copyResponseHeaders(c, result.Response.Header)
	c.Status(result.Response.StatusCode)

	if !result.Stream {
		responseBody, readErr := io.ReadAll(result.Response.Body)
		if readErr != nil {
			log.Printf("read anthropic upstream response: %v", readErr)
			h.proxy.LogMessagesResult(c.Request.Context(), result, nil, readErr)
			return
		}
		_, writeErr := c.Writer.Write(responseBody)
		if writeErr != nil {
			log.Printf("write anthropic upstream response: %v", writeErr)
		}
		h.proxy.LogMessagesResult(c.Request.Context(), result, responseBody, writeErr)
		return
	}

	preview := &limitedBuffer{limit: 4096}
	writer := flushWriter{writer: c.Writer}
	_, copyErr := io.Copy(writer, io.TeeReader(result.Response.Body, preview))
	if copyErr != nil {
		log.Printf("copy anthropic upstream response: %v", copyErr)
	}

	h.proxy.LogMessagesResult(c.Request.Context(), result, preview.Bytes(), copyErr)
}

func apiKeyFromRequest(c *gin.Context) string {
	if value := strings.TrimSpace(c.GetHeader("x-api-key")); value != "" {
		return value
	}
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[len("Bearer "):])
	}
	return ""
}

func handleAPIKeyError(c *gin.Context, err error) {
	var validationErr service.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeError(c, http.StatusUnauthorized, "authentication_error", validationErr.Message)
	case errors.Is(err, gorm.ErrRecordNotFound):
		writeError(c, http.StatusUnauthorized, "authentication_error", "invalid api key")
	default:
		writeError(c, http.StatusInternalServerError, "internal_error", err.Error())
	}
}

func handleProxyError(c *gin.Context, err error) {
	var validationErr service.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeError(c, http.StatusBadRequest, "invalid_request_error", validationErr.Message)
	case errors.Is(err, gorm.ErrRecordNotFound):
		writeError(c, http.StatusNotFound, "invalid_request_error", "model not found")
	default:
		writeError(c, http.StatusBadGateway, "upstream_error", err.Error())
	}
}

func writeError(c *gin.Context, status int, errorType string, message string) {
	c.JSON(status, gin.H{
		"type": "error",
		"error": gin.H{
			"type":    errorType,
			"message": message,
		},
	})
}

func copyResponseHeaders(c *gin.Context, headers http.Header) {
	for key, values := range headers {
		if isHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}
}

func isHopByHopHeader(key string) bool {
	switch http.CanonicalHeaderKey(key) {
	case "Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization",
		"Te", "Trailer", "Transfer-Encoding", "Upgrade", "Content-Length":
		return true
	default:
		return false
	}
}

type flushWriter struct {
	writer gin.ResponseWriter
}

func (w flushWriter) Write(data []byte) (int, error) {
	n, err := w.writer.Write(data)
	w.writer.Flush()
	return n, err
}

type limitedBuffer struct {
	data  []byte
	tail  []byte
	limit int
	total int
}

func (b *limitedBuffer) Write(data []byte) (int, error) {
	b.total += len(data)
	if len(b.data) < b.limit {
		remaining := b.limit - len(b.data)
		if len(data) > remaining {
			b.data = append(b.data, data[:remaining]...)
		} else {
			b.data = append(b.data, data...)
		}
	}
	b.tail = append(b.tail, data...)
	if len(b.tail) > b.limit {
		b.tail = b.tail[len(b.tail)-b.limit:]
	}
	return len(data), nil
}

func (b *limitedBuffer) Bytes() []byte {
	if b.total <= b.limit || len(b.tail) == 0 {
		return b.data
	}
	result := make([]byte, 0, len(b.data)+len(b.tail)+2)
	result = append(result, b.data...)
	result = append(result, '\n', '\n')
	result = append(result, b.tail...)
	return result
}
