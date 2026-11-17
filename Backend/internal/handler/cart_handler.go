package handler

import (
	"encoding/json"
	"net/http"

	"github.com/auction-studio/backend/internal/dto"
	"github.com/auction-studio/backend/internal/middleware"
	"github.com/auction-studio/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type CartHandler struct {
	cartService *service.CartService
}

func NewCartHandler(cartService *service.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		ErrorJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cart, err := h.cartService.GetCart(r.Context(), userID)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, cart)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		ErrorJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.AddCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.cartService.AddItem(r.Context(), userID, req.ItemID, req.Quantity); err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "item added to cart"})
}

func (h *CartHandler) UpdateQuantity(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		ErrorJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	itemID := chi.URLParam(r, "itemId")
	if itemID == "" {
		ErrorJSON(w, http.StatusBadRequest, "itemId is required")
		return
	}

	var req dto.UpdateCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.cartService.UpdateQuantity(r.Context(), userID, itemID, req.Quantity); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "cart updated"})
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		ErrorJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	itemID := chi.URLParam(r, "itemId")
	if itemID == "" {
		ErrorJSON(w, http.StatusBadRequest, "itemId is required")
		return
	}

	if err := h.cartService.RemoveItem(r.Context(), userID, itemID); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "item removed from cart"})
}

func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		ErrorJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.cartService.ClearCart(r.Context(), userID); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "cart cleared"})
}
