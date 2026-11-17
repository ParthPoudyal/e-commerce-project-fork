package handler

import (
	"net/http"

	"github.com/auction-studio/backend/internal/service"
)

type CategoryHandler struct {
	catService *service.CategoryService
}

func NewCategoryHandler(catService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{catService: catService}
}

func (h *CategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	cats, err := h.catService.GetAll(r.Context())
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, cats)
}
