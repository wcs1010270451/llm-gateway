package me

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"llm-gateway/backend/internal/entity"
	"llm-gateway/backend/internal/service"
)

type Handler struct {
	keys    *service.APIKeyService
	console *service.UserConsoleService
}

func NewHandler(keys *service.APIKeyService, console *service.UserConsoleService) *Handler {
	return &Handler{keys: keys, console: console}
}

func (h *Handler) ListAPIKeys(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "authentication_error", "login required")
		return
	}

	items, err := h.keys.ListForUser(c.Request.Context(), user)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *Handler) CreateAPIKey(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "authentication_error", "login required")
		return
	}

	var input service.APIKeyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	result, err := h.keys.CreateForUser(c.Request.Context(), user, input)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *Handler) UpdateAPIKey(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "authentication_error", "login required")
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}

	var input service.APIKeyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	item, err := h.keys.UpdateForUser(c.Request.Context(), user, id, input)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteAPIKey(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "authentication_error", "login required")
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.keys.DeleteForUser(c.Request.Context(), user, id); err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) GetAPIKey(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "authentication_error", "login required")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	item, err := h.console.GetAPIKey(c.Request.Context(), user, id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) ListModels(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "authentication_error", "login required")
		return
	}

	items, err := h.console.ListModels(c.Request.Context(), user)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *Handler) ListAPIKeyModelStats(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "authentication_error", "login required")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	items, err := h.console.ListAPIKeyModelStats(c.Request.Context(), user, id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *Handler) ListAPIKeyLogs(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "authentication_error", "login required")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	result, err := h.console.ListAPIKeyLogs(c.Request.Context(), user, id, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetLog(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "authentication_error", "login required")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	item, err := h.console.GetAPIKeyLog(c.Request.Context(), user, id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) DebugMessages(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "authentication_error", "login required")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		writeError(c, http.StatusBadRequest, "bad_request", "failed to read request body")
		return
	}

	result, err := h.console.OpenDebugAnthropicMessages(c.Request.Context(), user, id, service.AnthropicMessagesInput{
		Body:     body,
		Headers:  c.Request.Header,
		Path:     c.Request.URL.Path,
		ClientIP: c.ClientIP(),
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
			log.Printf("read anthropic debug upstream response: %v", readErr)
			h.console.LogMessagesResult(c.Request.Context(), result, nil, readErr)
			return
		}
		_, writeErr := c.Writer.Write(responseBody)
		if writeErr != nil {
			log.Printf("write anthropic debug upstream response: %v", writeErr)
		}
		h.console.LogMessagesResult(c.Request.Context(), result, responseBody, writeErr)
		return
	}

	// 流式响应不能完整缓冲，否则会破坏实时转发；这里只截取小片段用于日志预览。
	preview := &limitedBuffer{limit: 4096}
	writer := flushWriter{writer: c.Writer}
	_, copyErr := io.Copy(writer, io.TeeReader(result.Response.Body, preview))
	if copyErr != nil {
		log.Printf("copy anthropic debug upstream response: %v", copyErr)
	}
	h.console.LogMessagesResult(c.Request.Context(), result, preview.Bytes(), copyErr)
}

func (h *Handler) DebugChatCompletions(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "authentication_error", "login required")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		writeError(c, http.StatusBadRequest, "bad_request", "failed to read request body")
		return
	}

	result, err := h.console.OpenDebugOpenAIChatCompletions(c.Request.Context(), user, id, service.OpenAIChatCompletionsInput{
		Body:     body,
		Headers:  c.Request.Header,
		Path:     c.Request.URL.Path,
		ClientIP: c.ClientIP(),
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
			log.Printf("read openai debug upstream response: %v", readErr)
			h.console.LogChatCompletionsResult(c.Request.Context(), result, nil, readErr)
			return
		}
		_, writeErr := c.Writer.Write(responseBody)
		if writeErr != nil {
			log.Printf("write openai debug upstream response: %v", writeErr)
		}
		h.console.LogChatCompletionsResult(c.Request.Context(), result, responseBody, writeErr)
		return
	}

	// 流式响应不能完整缓冲，否则会破坏实时转发；这里只截取小片段用于日志预览。
	preview := &limitedBuffer{limit: 4096}
	writer := flushWriter{writer: c.Writer}
	_, copyErr := io.Copy(writer, io.TeeReader(result.Response.Body, preview))
	if copyErr != nil {
		log.Printf("copy openai debug upstream response: %v", copyErr)
	}
	h.console.LogChatCompletionsResult(c.Request.Context(), result, preview.Bytes(), copyErr)
}

func (h *Handler) DebugGeminiGenerateContent(c *gin.Context) {
	h.debugGemini(c, false)
}

func (h *Handler) DebugGeminiStreamGenerateContent(c *gin.Context) {
	h.debugGemini(c, true)
}

func (h *Handler) debugGemini(c *gin.Context, stream bool) {
	user, ok := currentUser(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "authentication_error", "login required")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		writeError(c, http.StatusBadRequest, "bad_request", "failed to read request body")
		return
	}

	input := service.GeminiGenerateContentInput{
		Body:     body,
		Headers:  c.Request.Header,
		Path:     c.Request.URL.Path,
		ClientIP: c.ClientIP(),
	}

	var result *service.GeminiGenerateContentResult
	if stream {
		result, err = h.console.OpenDebugGeminiStreamGenerateContent(c.Request.Context(), user, id, input)
	} else {
		result, err = h.console.OpenDebugGeminiGenerateContent(c.Request.Context(), user, id, input)
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
			log.Printf("read gemini debug upstream response: %v", readErr)
			h.console.LogGeminiGenerateContentResult(c.Request.Context(), result, nil, readErr)
			return
		}
		_, writeErr := c.Writer.Write(responseBody)
		if writeErr != nil {
			log.Printf("write gemini debug upstream response: %v", writeErr)
		}
		h.console.LogGeminiGenerateContentResult(c.Request.Context(), result, responseBody, writeErr)
		return
	}

	preview := &limitedBuffer{limit: 4096}
	writer := flushWriter{writer: c.Writer}
	_, copyErr := io.Copy(writer, io.TeeReader(result.Response.Body, preview))
	if copyErr != nil {
		log.Printf("copy gemini debug upstream response: %v", copyErr)
	}
	h.console.LogGeminiGenerateContentResult(c.Request.Context(), result, preview.Bytes(), copyErr)
}

func currentUser(c *gin.Context) (entity.User, bool) {
	value, ok := c.Get("current_user")
	if !ok {
		return entity.User{}, false
	}
	user, ok := value.(entity.User)
	return user, ok
}

func parseID(c *gin.Context, names ...string) (int64, bool) {
	name := "id"
	if len(names) > 0 && names[0] != "" {
		name = names[0]
	}
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, http.StatusBadRequest, "bad_request", "invalid id")
		return 0, false
	}
	return id, true
}

func handleError(c *gin.Context, err error) {
	var validationErr service.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeError(c, http.StatusBadRequest, "bad_request", validationErr.Message)
	case errors.Is(err, gorm.ErrRecordNotFound):
		writeError(c, http.StatusNotFound, "not_found", "record not found")
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
	limit int
}

func (b *limitedBuffer) Write(data []byte) (int, error) {
	if len(b.data) < b.limit {
		remaining := b.limit - len(b.data)
		if len(data) > remaining {
			b.data = append(b.data, data[:remaining]...)
		} else {
			b.data = append(b.data, data...)
		}
	}
	return len(data), nil
}

func (b *limitedBuffer) Bytes() []byte {
	return b.data
}
