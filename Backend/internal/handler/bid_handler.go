package handler

import (
	"encoding/json"
	"net/http"

	"github.com/auction-studio/backend/internal/dto"
	"github.com/auction-studio/backend/internal/middleware"
	"github.com/auction-studio/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type BidHandler struct {
	bidService *service.BidService
}

func NewBidHandler(bidService *service.BidService) *BidHandler {
	return &BidHandler{bidService: bidService}
}

func (h *BidHandler) PlaceBid(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		ErrorJSON(w, http.StatusUnauthorized, "must be signed in to place a bid")
		return
	}

	itemID := chi.URLParam(r, "id")
	if itemID == "" {
		ErrorJSON(w, http.StatusBadRequest, "itemId is required")
		return
	}

	var req dto.PlaceBidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.bidService.PlaceBid(r.Context(), userID, itemID, req.Amount)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, resp)
}

func (h *BidHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "id")
	if itemID == "" {
		ErrorJSON(w, http.StatusBadRequest, "itemId is required")
		return
	}

	history, err := h.bidService.GetHistory(r.Context(), itemID)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, history)
}

func (h *BidHandler) CancelBid(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		ErrorJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	bidID := chi.URLParam(r, "id")
	if bidID == "" {
		ErrorJSON(w, http.StatusBadRequest, "bidId is required")
		return
	}

	if err := h.bidService.CancelBid(r.Context(), userID, bidID); err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "bid cancelled successfully"})
}

func (h *BidHandler) GetUserBids(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		ErrorJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	bids, err := h.bidService.GetUserBids(r.Context(), userID)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, bids)
}

func (h *BidHandler) GetUserWonItems(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		ErrorJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.bidService.GetUserWonItems(r.Context(), userID)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, items)
}
