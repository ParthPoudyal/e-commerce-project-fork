package handler

import (
	"encoding/json"
	"net/http"

	"github.com/auction-studio/backend/internal/dto"
	"github.com/auction-studio/backend/internal/middleware"
	"github.com/auction-studio/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		ErrorJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.orderService.CreateOrder(r.Context(), userID, req)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, resp)
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		ErrorJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	orders, err := h.orderService.GetOrders(r.Context(), userID)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, orders)
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		ErrorJSON(w, http.StatusBadRequest, "order id is required")
		return
	}

	order, err := h.orderService.GetByID(r.Context(), orderID)
	if err != nil {
		ErrorJSON(w, http.StatusNotFound, err.Error())
		return
	}

	JSON(w, http.StatusOK, order)
}
