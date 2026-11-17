package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type InitiateRequest struct {
	ReturnURL      string `json:"return_url"`
	WebsiteURL     string `json:"website_url"`
	Amount         int64  `json:"amount"` // in paisa (1 NPR = 100 paisa)
	PurchaseOrderID string `json:"purchase_order_id"`
	PurchaseOrderName string `json:"purchase_order_name"`
	CustomerInfo   CustomerInfo `json:"customer_info,omitempty"`
}

type CustomerInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type InitiateResponse struct {
	PIDX       string `json:"pidx"`
	PaymentURL string `json:"payment_url"`
	ExpiresAt  string `json:"expires_at"`
}

type VerifyResponse struct {
	PIDX          string `json:"pidx"`
	Status        string `json:"status"` // Completed, Pending, Expired, User canceled
	TransactionID string `json:"transaction_id"`
	TotalAmount   int64  `json:"total_amount"`
	Fee           int64  `json:"fee"`
	Refunded      bool   `json:"refunded"`
}

type Provider interface {
	Name() string
	Initiate(ctx context.Context, req InitiateRequest) (*InitiateResponse, error)
	Verify(ctx context.Context, pidx string) (*VerifyResponse, error)
}

type KhaltiProvider struct {
	SecretKey string
	PublicKey string
	BaseURL   string
	client    *http.Client
}

func NewKhaltiProvider(secretKey, publicKey, baseURL string) *KhaltiProvider {
	if baseURL == "" {
		baseURL = "https://a.khalti.com/api/v2"
	}
	return &KhaltiProvider{
		SecretKey: secretKey,
		PublicKey: publicKey,
		BaseURL:   baseURL,
		client:    &http.Client{Timeout: 15 * time.Second},
	}
}

func (k *KhaltiProvider) Name() string {
	return "khalti"
}

func (k *KhaltiProvider) Initiate(ctx context.Context, req InitiateRequest) (*InitiateResponse, error) {
	// If no secret key is provided (e.g. dev/test mode), simulate a successful initiation
	if k.SecretKey == "" || k.SecretKey == "change-me" {
		pidx := fmt.Sprintf("mock-pidx-%d", time.Now().UnixNano())
		return &InitiateResponse{
			PIDX:       pidx,
			PaymentURL: fmt.Sprintf("%s?pidx=%s&status=mock_ready", req.ReturnURL, pidx),
			ExpiresAt:  time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		}, nil
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/epayment/initiate/", k.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Key %s", k.SecretKey))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := k.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("khalti api error (%d): %s", resp.StatusCode, string(respBody))
	}

	var initResp InitiateResponse
	if err := json.Unmarshal(respBody, &initResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &initResp, nil
}

func (k *KhaltiProvider) Verify(ctx context.Context, pidx string) (*VerifyResponse, error) {
	// If mock/sandbox without key
	if k.SecretKey == "" || k.SecretKey == "change-me" {
		return &VerifyResponse{
			PIDX:          pidx,
			Status:        "Completed",
			TransactionID: fmt.Sprintf("MOCK-TXN-%d", time.Now().Unix()),
			TotalAmount:   1000,
			Refunded:      false,
		}, nil
	}

	body, err := json.Marshal(map[string]string{"pidx": pidx})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/epayment/lookup/", k.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Key %s", k.SecretKey))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := k.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("khalti lookup error (%d): %s", resp.StatusCode, string(respBody))
	}

	var vResp VerifyResponse
	if err := json.Unmarshal(respBody, &vResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &vResp, nil
}
