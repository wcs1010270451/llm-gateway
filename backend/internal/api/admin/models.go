package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"llm-gateway/backend/internal/api/apierror"
	"llm-gateway/backend/internal/service"
)

func (h *Handler) ListModels(c *gin.Context) {
	if c.Query("page") != "" || c.Query("page_size") != "" || c.Query("family") != "" {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
		result, err := h.models.ListPage(c.Request.Context(), c.Query("family"), page, pageSize)
		if err != nil {
			apierror.Database(c, err)
			return
		}

		c.JSON(http.StatusOK, result)
		return
	}

	items, err := h.models.List(c.Request.Context())
	if err != nil {
		apierror.Database(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": len(items),
	})
}

func (h *Handler) GetModel(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	item, err := h.models.Get(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) CreateModel(c *gin.Context) {
	var input service.ModelInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.BadRequest(c, err.Error())
		return
	}

	item, err := h.models.Create(c.Request.Context(), input)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *Handler) UpdateModel(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	var input service.ModelInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.BadRequest(c, err.Error())
		return
	}

	item, err := h.models.Update(c.Request.Context(), id, input)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteModel(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	if err := h.models.Delete(c.Request.Context(), id); err != nil {
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ListProviderModels(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	items, err := h.models.ListProviderModels(c.Request.Context(), id)
	if err != nil {
		apierror.Database(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": len(items),
	})
}

func (h *Handler) CreateProviderModel(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	var input service.ProviderModelInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.BadRequest(c, err.Error())
		return
	}

	item, err := h.models.CreateProviderModel(c.Request.Context(), id, input)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *Handler) UpdateProviderModel(c *gin.Context) {
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

	item, err := h.models.UpdateProviderModel(c.Request.Context(), id, providerModelID, input)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteProviderModel(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	providerModelID, ok := parseID(c, "providerModelID")
	if !ok {
		return
	}

	if err := h.models.DeleteProviderModel(c.Request.Context(), id, providerModelID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) SetActiveProviderModel(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	providerModelID, ok := parseID(c, "providerModelID")
	if !ok {
		return
	}

	item, err := h.models.SetActiveProviderModel(c.Request.Context(), id, providerModelID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}
