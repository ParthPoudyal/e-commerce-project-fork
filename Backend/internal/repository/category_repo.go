package repository

import (
	"context"

	"github.com/auction-studio/backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) GetAll(ctx context.Context) ([]models.Category, error) {
	query := `SELECT id, label, slug, parent_id FROM categories ORDER BY id ASC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	allCats := make([]models.Category, 0)
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Label, &c.Slug, &c.ParentID); err != nil {
			return nil, err
		}
		allCats = append(allCats, c)
	}

	// Build tree: top-level categories have parent_id == nil
	topLevel := make([]models.Category, 0)
	childrenMap := make(map[string][]models.Category)

	for _, c := range allCats {
		if c.ParentID != nil && *c.ParentID != "" {
			childrenMap[*c.ParentID] = append(childrenMap[*c.ParentID], c)
		}
	}

	for _, c := range allCats {
		if c.ParentID == nil || *c.ParentID == "" {
			c.Children = childrenMap[c.ID]
			if c.Children == nil {
				c.Children = []models.Category{}
			}
			topLevel = append(topLevel, c)
		}
	}

	return topLevel, nil
}
