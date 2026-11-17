package repository

import (
	"context"
	"errors"
	"time"

	"github.com/auction-studio/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Create(ctx context.Context, p *models.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, provider, pidx, provider_txn_id, amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	_, err := r.pool.Exec(ctx, query, p.ID, p.OrderID, p.Provider, p.PIDX, p.ProviderTxnID, p.Amount, p.Status, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *PaymentRepository) GetByPIDX(ctx context.Context, pidx string) (*models.Payment, error) {
	query := `
		SELECT id, order_id, provider, pidx, provider_txn_id, amount, status, created_at, updated_at
		FROM payments
		WHERE pidx = $1
	`
	var p models.Payment
	var pidxVal *string
	err := r.pool.QueryRow(ctx, query, pidx).Scan(
		&p.ID, &p.OrderID, &p.Provider, &pidxVal, &p.ProviderTxnID, &p.Amount, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if pidxVal != nil {
		p.PIDX = *pidxVal
	}
	return &p, err
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*models.Payment, error) {
	query := `
		SELECT id, order_id, provider, pidx, provider_txn_id, amount, status, created_at, updated_at
		FROM payments
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var p models.Payment
	var pidxVal *string
	err := r.pool.QueryRow(ctx, query, orderID).Scan(
		&p.ID, &p.OrderID, &p.Provider, &pidxVal, &p.ProviderTxnID, &p.Amount, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if pidxVal != nil {
		p.PIDX = *pidxVal
	}
	return &p, err
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, pidx string, status models.PaymentStatus, txnID string) error {
	query := `
		UPDATE payments
		SET status = $1, provider_txn_id = $2, updated_at = NOW()
		WHERE pidx = $3
	`
	_, err := r.pool.Exec(ctx, query, status, txnID, pidx)
	return err
}
