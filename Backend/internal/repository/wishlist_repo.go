package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WishlistRepository struct {
	pool *pgxpool.Pool
}

func NewWishlistRepository(pool *pgxpool.Pool) *WishlistRepository {
	return &WishlistRepository{pool: pool}
}

func (r *WishlistRepository) GetItemIDsByUserID(ctx context.Context, userID string) ([]string, error) {
	query := `SELECT item_id FROM wishlist_items WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *WishlistRepository) Add(ctx context.Context, userID, itemID string) error {
	query := `
		INSERT INTO wishlist_items (user_id, item_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, item_id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, userID, itemID)
	return err
}

func (r *WishlistRepository) Remove(ctx context.Context, userID, itemID string) error {
	query := `DELETE FROM wishlist_items WHERE user_id = $1 AND item_id = $2`
	_, err := r.pool.Exec(ctx, query, userID, itemID)
	return err
}

func (r *WishlistRepository) Exists(ctx context.Context, userID, itemID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM wishlist_items WHERE user_id = $1 AND item_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, itemID).Scan(&exists)
	return exists, err
}
