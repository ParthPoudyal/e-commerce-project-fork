package handler

import (
	"encoding/json"
	"net/http"

	"github.com/auction-studio/backend/internal/dto"
	"github.com/auction-studio/backend/internal/middleware"
	"github.com/auction-studio/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type WishlistHandler struct {
	wishlistService *service.WishlistService
}

func NewWishlistHandler(wishlistService *service.WishlistService) *WishlistHandler {
	return &WishlistHandler{wishlistService: wishlistService}
}

func (h *WishlistHandler) GetWishlist(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		ErrorJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.wishlistService.GetWishlist(r.Context(), userID)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, items)
}

func (h *WishlistHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		ErrorJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.WishlistActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	inWishlist, err := h.wishlistService.Toggle(r.Context(), userID, req.ItemID)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"itemId":     req.ItemID,
		"inWishlist": inWishlist,
	})
}

func (h *WishlistHandler) Remove(w http.ResponseWriter, r *http.Request) {
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

	if err := h.wishlistService.Remove(r.Context(), userID, itemID); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "removed from wishlist"})
}
