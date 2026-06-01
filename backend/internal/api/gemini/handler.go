package gemini

import (
	"bytes"
	"encoding/json"
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
	proxy   *service.GeminiProxyService
	apiKeys *service.APIKeyService
}

func NewHandler(proxy *service.GeminiProxyService, apiKeys *service.APIKeyService) *Handler {
	return &Handler{proxy: proxy, apiKeys: apiKeys}
}

func (h *Handler) GenerateContent(c *gin.Context) {
	h.handle(c, false)
}

func (h *Handler) StreamGenerateContent(c *gin.Context) {
	h.handle(c, true)
}

func (h *Handler) NativeAPI(c *gin.Context) {
	action := strings.TrimPrefix(c.Param("action"), "/")
	colonIndex := strings.LastIndex(action, ":")
	if colonIndex <= 0 {
		writeError(c, http.StatusNotFound, "not_found", "not found")
		return
	}

	model := strings.TrimSpace(action[:colonIndex])
	method := strings.TrimSpace(action[colonIndex+1:])
	if model == "" {
		writeError(c, http.StatusBadRequest, "invalid_request_error", "model is required in path")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request_error", "failed to read request body")
		return
	}
	payload := make(map[string]any)
	if len(bytes.TrimSpace(body)) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			writeError(c, http.StatusBadRequest, "invalid_request_error", "request body must be valid JSON")
			return
		}
	}
	payload["model"] = model
	body, _ = json.Marshal(payload)
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	switch method {
	case "generateContent":
		h.GenerateContent(c)
	case "streamGenerateContent":
		h.StreamGenerateContent(c)
	default:
		writeError(c, http.StatusNotFound, "not_found", "unsupported method: "+method)
	}
}

func (h *Handler) handle(c *gin.Context, stream bool) {
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

	input := service.GeminiGenerateContentInput{
		Body:     body,
		Headers:  c.Request.Header,
		Method:   c.Request.Method,
		Path:     c.Request.URL.Path,
		ClientIP: c.ClientIP(),
		UserID:   apiAuth.User.ID,
		APIKeyID: apiAuth.APIKey.ID,
	}

	var result *service.GeminiGenerateContentResult
	if stream {
		result, err = h.proxy.OpenStreamGenerateContent(c.Request.Context(), input)
	} else {
		result, err = h.proxy.OpenGenerateContent(c.Request.Context(), input)
	}
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
			log.Printf("read gemini upstream response: %v", readErr)
			h.proxy.LogGenerateContentResult(c.Request.Context(), result, nil, readErr)
			return
		}
		_, writeErr := c.Writer.Write(responseBody)
		if writeErr != nil {
			log.Printf("write gemini upstream response: %v", writeErr)
		}
		h.proxy.LogGenerateContentResult(c.Request.Context(), result, responseBody, writeErr)
		return
	}

	preview := &limitedBuffer{limit: 4096}
	writer := flushWriter{writer: c.Writer}
	_, copyErr := io.Copy(writer, io.TeeReader(result.Response.Body, preview))
	if copyErr != nil {
		log.Printf("copy gemini upstream response: %v", copyErr)
	}
	h.proxy.LogGenerateContentResult(c.Request.Context(), result, preview.Bytes(), copyErr)
}

func apiKeyFromRequest(c *gin.Context) string {
	if value := strings.TrimSpace(c.GetHeader("x-api-key")); value != "" {
		return value
	}
	if value := strings.TrimSpace(c.GetHeader("x-goog-api-key")); value != "" {
		return value
	}
	if value := strings.TrimSpace(c.Query("key")); value != "" {
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
	c.JSON(status, gin.H{"error": gin.H{"type": errorType, "message": message}})
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
