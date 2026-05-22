package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.stats.Get(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}
