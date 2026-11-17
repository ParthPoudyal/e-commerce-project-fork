package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/auction-studio/backend/internal/dto"
	"github.com/auction-studio/backend/internal/models"
	"github.com/auction-studio/backend/internal/payment"
	"github.com/auction-studio/backend/internal/repository"
	"github.com/google/uuid"
)

type PaymentService struct {
	paymentRepo *repository.PaymentRepository
	orderRepo   *repository.OrderRepository
	paymentProv payment.Provider
}

func NewPaymentService(paymentRepo *repository.PaymentRepository, orderRepo *repository.OrderRepository, paymentProv payment.Provider) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
		paymentProv: paymentProv,
	}
}

func (s *PaymentService) InitiatePayment(ctx context.Context, orderID string) (*dto.PaymentInitiateResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil || order == nil {
		return nil, errors.New("order not found")
	}

	initReq := payment.InitiateRequest{
		ReturnURL:         "http://localhost:5173/order/" + orderID,
		WebsiteURL:        "http://localhost:5173",
		Amount:            order.Total * 100, // NPR to Paisa
		PurchaseOrderID:   orderID,
		PurchaseOrderName: fmt.Sprintf("Auction Studio Order %s", orderID),
		CustomerInfo: payment.CustomerInfo{
			Name:  fmt.Sprintf("%s %s", order.ContactInfo.FirstName, order.ContactInfo.LastName),
			Email: order.ContactInfo.Email,
			Phone: order.ContactInfo.Phone,
		},
	}

	resp, err := s.paymentProv.Initiate(ctx, initReq)
	if err != nil {
		return nil, fmt.Errorf("initiate payment: %w", err)
	}

	pay := &models.Payment{
		ID:       "pay-" + uuid.New().String()[:8],
		OrderID:  orderID,
		Provider: s.paymentProv.Name(),
		PIDX:     resp.PIDX,
		Amount:   order.Total,
		Status:   models.PaymentStatusInitiated,
	}
	_ = s.paymentRepo.Create(ctx, pay)

	return &dto.PaymentInitiateResponse{
		PIDX:       resp.PIDX,
		PaymentURL: resp.PaymentURL,
		ExpiresAt:  resp.ExpiresAt,
	}, nil
}

func (s *PaymentService) VerifyPayment(ctx context.Context, pidx string) (*dto.PaymentVerifyResponse, error) {
	vResp, err := s.paymentProv.Verify(ctx, pidx)
	if err != nil {
		return nil, fmt.Errorf("verify payment with gateway: %w", err)
	}

	pay, err := s.paymentRepo.GetByPIDX(ctx, pidx)
	if err != nil || pay == nil {
		return nil, errors.New("payment record not found")
	}

	if vResp.Status == "Completed" {
		_ = s.paymentRepo.UpdateStatus(ctx, pidx, models.PaymentStatusCompleted, vResp.TransactionID)
		_ = s.orderRepo.UpdateStatus(ctx, pay.OrderID, models.OrderStatusPaid)
	} else {
		_ = s.paymentRepo.UpdateStatus(ctx, pidx, models.PaymentStatusFailed, vResp.TransactionID)
	}

	return &dto.PaymentVerifyResponse{
		Status:        vResp.Status,
		TransactionID: vResp.TransactionID,
		Amount:        vResp.TotalAmount / 100, // paisa to NPR
		Message:       fmt.Sprintf("Payment %s", vResp.Status),
	}, nil
}
