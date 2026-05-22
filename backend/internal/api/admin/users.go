package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"llm-gateway/backend/internal/api/apierror"
	"llm-gateway/backend/internal/service"
)

func (h *Handler) ListUsers(c *gin.Context) {
	items, err := h.users.List(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": len(items),
	})
}

func (h *Handler) GetUser(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	item, err := h.users.Get(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var input service.UserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.BadRequest(c, err.Error())
		return
	}

	item, err := h.users.Create(c.Request.Context(), input)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	var input service.UserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.BadRequest(c, err.Error())
		return
	}

	item, err := h.users.Update(c.Request.Context(), id, input)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}

	if err := h.users.Delete(c.Request.Context(), id); err != nil {
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
