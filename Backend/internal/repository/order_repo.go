package repository

import (
	"context"
	"errors"

	"github.com/auction-studio/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *OrderRepository) Create(ctx context.Context, tx pgx.Tx, o *models.Order) error {
	orderQuery := `
		INSERT INTO orders (
			id, user_id, status, first_name, last_name, email, phone,
			line1, line2, city, state, zip, country,
			subtotal, shipping_cost, buyers_premium, total,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17,
			$18, $19
		)
	`
	_, err := tx.Exec(ctx, orderQuery,
		o.ID, o.UserID, o.Status, o.ContactInfo.FirstName, o.ContactInfo.LastName, o.ContactInfo.Email, o.ContactInfo.Phone,
		o.ShippingAddress.Line1, o.ShippingAddress.Line2, o.ShippingAddress.City, o.ShippingAddress.State, o.ShippingAddress.Zip, o.ShippingAddress.Country,
		o.Subtotal, o.ShippingCost, o.BuyersPremium, o.Total,
		o.CreatedAt, o.UpdatedAt,
	)
	if err != nil {
		return err
	}

	itemQuery := `
		INSERT INTO order_items (id, order_id, item_id, title, quantity, price, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`
	for _, item := range o.Items {
		_, err := tx.Exec(ctx, itemQuery,
			o.ID+"-"+item.ItemID, o.ID, item.ItemID, item.Title, item.Quantity, item.Price,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	query := `
		SELECT 
			id, user_id, status, first_name, last_name, email, phone,
			line1, line2, city, state, zip, country,
			subtotal, shipping_cost, buyers_premium, total,
			created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	var o models.Order
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.UserID, &o.Status, &o.ContactInfo.FirstName, &o.ContactInfo.LastName, &o.ContactInfo.Email, &o.ContactInfo.Phone,
		&o.ShippingAddress.Line1, &o.ShippingAddress.Line2, &o.ShippingAddress.City, &o.ShippingAddress.State, &o.ShippingAddress.Zip, &o.ShippingAddress.Country,
		&o.Subtotal, &o.ShippingCost, &o.BuyersPremium, &o.Total,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	items, err := r.getOrderItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

func (r *OrderRepository) GetByUserID(ctx context.Context, userID string) ([]models.Order, error) {
	query := `
		SELECT 
			id, user_id, status, first_name, last_name, email, phone,
			line1, line2, city, state, zip, country,
			subtotal, shipping_cost, buyers_premium, total,
			created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
	for rows.Next() {
		var o models.Order
		err := rows.Scan(
			&o.ID, &o.UserID, &o.Status, &o.ContactInfo.FirstName, &o.ContactInfo.LastName, &o.ContactInfo.Email, &o.ContactInfo.Phone,
			&o.ShippingAddress.Line1, &o.ShippingAddress.Line2, &o.ShippingAddress.City, &o.ShippingAddress.State, &o.ShippingAddress.Zip, &o.ShippingAddress.Country,
			&o.Subtotal, &o.ShippingCost, &o.BuyersPremium, &o.Total,
			&o.CreatedAt, &o.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	for i := range orders {
		items, err := r.getOrderItems(ctx, orders[i].ID)
		if err == nil {
			orders[i].Items = items
		}
	}

	return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID string, status models.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, orderID)
	return err
}

func (r *OrderRepository) getOrderItems(ctx context.Context, orderID string) ([]models.OrderItem, error) {
	query := `SELECT item_id, title, quantity, price FROM order_items WHERE order_id = $1`
	rows, err := r.pool.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.OrderItem, 0)
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ItemID, &item.Title, &item.Quantity, &item.Price); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
