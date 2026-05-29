package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListClaudeProxyStatus(c *gin.Context) {
	items, err := h.claude.ListStatus(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": len(items),
	})
}

func (h *Handler) ProbeClaudeProxy(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	result, err := h.claude.Probe(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
