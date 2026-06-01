package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"llm-gateway/backend/internal/api/apierror"
	"llm-gateway/backend/internal/service"
)

func (h *Handler) ListProviders(c *gin.Context) {
	items, err := h.providers.List(c.Request.Context())
	if err != nil {
		apierror.Database(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": len(items),
	})
}

func (h *Handler) GetProvider(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	item, err := h.providers.Get(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) ListProviderModelsByProvider(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	items, err := h.models.ListProviderModelsByProvider(c.Request.Context(), id)
	if err != nil {
		apierror.Database(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": len(items),
	})
}

func (h *Handler) GetProviderUsageStats(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	if _, err := h.providers.Get(c.Request.Context(), id); err != nil {
		handleServiceError(c, err)
		return
	}

	stats, err := h.logs.ProviderUsageStats(c.Request.Context(), id, c.Query("window"), c.Query("granularity"))
	if err != nil {
		apierror.Database(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) CreateProviderModelForProvider(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	var input service.ProviderModelInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.BadRequest(c, err.Error())
		return
	}

	item, err := h.models.CreateProviderModelForProvider(c.Request.Context(), id, input)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *Handler) UpdateProviderModelForProvider(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	providerModelID, ok := parseID(c, "providerModelID")
	if !ok {
		return
	}

	var input service.ProviderModelInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.BadRequest(c, err.Error())
		return
	}

	item, err := h.models.UpdateProviderModelForProvider(c.Request.Context(), id, providerModelID, input)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteProviderModelForProvider(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	providerModelID, ok := parseID(c, "providerModelID")
	if !ok {
		return
	}

	if err := h.models.DeleteProviderModelForProvider(c.Request.Context(), id, providerModelID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) CreateProvider(c *gin.Context) {
	var input service.ProviderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.BadRequest(c, err.Error())
		return
	}

	item, err := h.providers.Create(c.Request.Context(), input)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *Handler) UpdateProvider(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	var input service.ProviderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.BadRequest(c, err.Error())
		return
	}

	item, err := h.providers.Update(c.Request.Context(), id, input)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteProvider(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	if err := h.providers.Delete(c.Request.Context(), id); err != nil {
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
