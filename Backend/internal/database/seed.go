package database

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/auction-studio/backend/internal/auth"
	"github.com/auction-studio/backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Seed(pool *pgxpool.Pool) error {
	ctx := context.Background()

	// Seed bids if missing for any auction items
	seedBidsIfMissing(ctx, pool)

	// Check if already seeded
	var count int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM items").Scan(&count)
	if err == nil && count >= 35 {
		log.Println("Database already seeded. Skipping item creation.")
		return nil
	}

	log.Println("Seeding database...")

	// 1. Seed Users
	defaultHash, _ := auth.HashPassword("password123")
	adminHash, _ := auth.HashPassword("admin123")

	users := []models.User{
		{ID: "usr-admin", Name: "Auction Administrator", Email: "admin@auctionstudio.com", PasswordHash: adminHash, Role: models.RoleAdmin, Location: "Kathmandu, Nepal", Rating: 5.0},
		{ID: "usr-seller-1", Name: "Harrington & Sons", Email: "harrington@auctionstudio.com", PasswordHash: defaultHash, Role: models.RoleSeller, Location: "Patan, Nepal", Rating: 4.9},
		{ID: "usr-seller-2", Name: "Himalayan Antiquities", Email: "himalayan@auctionstudio.com", PasswordHash: defaultHash, Role: models.RoleSeller, Location: "Kathmandu, Nepal", Rating: 4.8},
		{ID: "usr-seller-3", Name: "Zürich Watch Co.", Email: "zurich@auctionstudio.com", PasswordHash: defaultHash, Role: models.RoleSeller, Location: "Zürich / Kathmandu", Rating: 5.0},
		{ID: "usr-seller-4", Name: "Bauman Rare Books", Email: "bauman@auctionstudio.com", PasswordHash: defaultHash, Role: models.RoleSeller, Location: "Bhaktapur, Nepal", Rating: 5.0},
		{ID: "usr-buyer-1", Name: "Jane Collector", Email: "jane@example.com", PasswordHash: defaultHash, Role: models.RoleBuyer, Location: "Pokhara, Nepal", Rating: 4.9},
		{ID: "usr-buyer-2", Name: "Ram Shrestha", Email: "ram@example.com", PasswordHash: defaultHash, Role: models.RoleBuyer, Location: "Kathmandu, Nepal", Rating: 4.7},
	}

	userInsert := `
		INSERT INTO users (id, name, email, password_hash, role, phone, location, rating, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`
	for _, u := range users {
		_, _ = pool.Exec(ctx, userInsert, u.ID, u.Name, u.Email, u.PasswordHash, u.Role, u.Phone, u.Location, u.Rating)
	}

	// 2. Seed Categories
	categories := []struct {
		ID       string
		Label    string
		Slug     string
		ParentID *string
	}{
		// Top Level
		{"cat-1", "Antiques", "antiques", nil},
		{"cat-2", "Furniture", "furniture", nil},
		{"cat-3", "Ceramics & Glass", "ceramics-glass", nil},
		{"cat-4", "Clocks & Watches", "clocks-watches", nil},
		{"cat-5", "Art", "art", nil},
		{"cat-6", "Jewellery", "jewellery", nil},
		{"cat-7", "Books & Maps", "books-maps", nil},
		{"cat-8", "General", "general", nil},

		// Subcategories
		{"cat-1-1", "Decorative Arts", "decorative-arts", ptr("cat-1")},
		{"cat-1-2", "Metalware", "metalware", ptr("cat-1")},
		{"cat-1-3", "Textiles", "textiles", ptr("cat-1")},
		{"cat-1-4", "Scientific Instruments", "scientific-instruments", ptr("cat-1")},

		{"cat-2-1", "Tables", "tables", ptr("cat-2")},
		{"cat-2-2", "Chairs & Seating", "chairs-seating", ptr("cat-2")},
		{"cat-2-3", "Cabinets & Storage", "cabinets-storage", ptr("cat-2")},
		{"cat-2-4", "Desks", "desks", ptr("cat-2")},

		{"cat-3-1", "Porcelain", "porcelain", ptr("cat-3")},
		{"cat-3-2", "Pottery", "pottery", ptr("cat-3")},
		{"cat-3-3", "Art Glass", "art-glass", ptr("cat-3")},
		{"cat-3-4", "Stoneware", "stoneware", ptr("cat-3")},

		{"cat-4-1", "Mantel Clocks", "mantel-clocks", ptr("cat-4")},
		{"cat-4-2", "Wall Clocks", "wall-clocks", ptr("cat-4")},
		{"cat-4-3", "Pocket Watches", "pocket-watches", ptr("cat-4")},
		{"cat-4-4", "Wristwatches", "wristwatches", ptr("cat-4")},

		{"cat-5-1", "Paintings", "paintings", ptr("cat-5")},
		{"cat-5-2", "Prints", "prints", ptr("cat-5")},
		{"cat-5-3", "Sculpture", "sculpture", ptr("cat-5")},
		{"cat-5-4", "Photography", "photography", ptr("cat-5")},

		{"cat-6-1", "Rings", "rings", ptr("cat-6")},
		{"cat-6-2", "Necklaces", "necklaces", ptr("cat-6")},
		{"cat-6-3", "Brooches", "brooches", ptr("cat-6")},
		{"cat-6-4", "Watches", "jewellery-watches", ptr("cat-6")},

		{"cat-7-1", "Rare Books", "rare-books", ptr("cat-7")},
		{"cat-7-2", "Maps & Atlases", "maps-atlases", ptr("cat-7")},
		{"cat-7-3", "Manuscripts", "manuscripts", ptr("cat-7")},

		{"cat-8-1", "Collectibles", "collectibles", ptr("cat-8")},
		{"cat-8-2", "Vintage Electronics", "vintage-electronics", ptr("cat-8")},
		{"cat-8-3", "Tools & Equipment", "tools-equipment", ptr("cat-8")},
		{"cat-8-4", "Miscellaneous", "miscellaneous", ptr("cat-8")},
	}

	catInsert := `
		INSERT INTO categories (id, label, slug, parent_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`
	for _, c := range categories {
		_, _ = pool.Exec(ctx, catInsert, c.ID, c.Label, c.Slug, c.ParentID)
	}

	// 3. Seed 40 Items in NPR
	seedItems := getSeedItems()
	itemInsert := `
		INSERT INTO items (
			id, title, lot_number, category_slug, subcategory, description,
			era, origin, material, dimensions, condition, provenance,
			images, purchase_mode, price, stock, current_bid, starting_bid,
			minimum_increment, bid_count, end_time, status, seller_id,
			shipping_cost, shipping_estimated_days, shipping_methods,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22, $23,
			$24, $25, $26,
			NOW(), NOW()
		)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			price = EXCLUDED.price,
			current_bid = EXCLUDED.current_bid,
			starting_bid = EXCLUDED.starting_bid,
			shipping_cost = EXCLUDED.shipping_cost,
			updated_at = NOW()
	`

	for _, it := range seedItems {
		imagesJSON, _ := json.Marshal(it.Images)
		methodsJSON, _ := json.Marshal(it.Shipping.Methods)

		_, err := pool.Exec(ctx, itemInsert,
			it.ID, it.Title, it.LotNumber, it.CategorySlug, it.Subcategory, it.Description,
			it.Specs.Era, it.Specs.Origin, it.Specs.Material, it.Specs.Dimensions, it.Specs.Condition, it.Specs.Provenance,
			imagesJSON, it.PurchaseMode, it.Price, it.Stock, it.CurrentBid, it.StartingBid,
			it.MinimumIncrement, it.BidCount, it.EndTime, "active", it.SellerID,
			it.Shipping.Cost, it.Shipping.EstimatedDays, methodsJSON,
		)
		if err != nil {
			log.Printf("Failed to seed item %s: %v", it.ID, err)
		}
	}

	log.Println("Seeding complete. 40 items in NPR ready.")
	return nil
}

func seedBidsIfMissing(ctx context.Context, pool *pgxpool.Pool) {
	seedItems := getSeedItems()
	bidInsert := `
		INSERT INTO bids (id, item_id, user_id, amount, status, created_at)
		VALUES ($1, $2, $3, $4, 'active', $5)
		ON CONFLICT (id) DO NOTHING
	`
	for _, it := range seedItems {
		if it.PurchaseMode == models.PurchaseModeAuction && it.CurrentBid != nil && *it.CurrentBid > 0 {
			var bidCount int
			_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM bids WHERE item_id = $1", it.ID).Scan(&bidCount)
			if bidCount == 0 {
				startBid := int64(100000)
				if it.StartingBid != nil {
					startBid = *it.StartingBid
				}
				currBid := *it.CurrentBid
				diff := currBid - startBid
				step := diff / 2
				if step <= 0 {
					step = 10000
				}

				// First bid from Ram Shrestha
				_, _ = pool.Exec(ctx, bidInsert, "bid-init-1-"+it.ID, it.ID, "usr-buyer-2", startBid+step/2, time.Now().Add(-2*time.Hour))
				// Second/highest bid from Jane Collector
				_, _ = pool.Exec(ctx, bidInsert, "bid-init-2-"+it.ID, it.ID, "usr-buyer-1", currBid, time.Now().Add(-30*time.Minute))
			}
		}
	}
}

func ptr[T any](v T) *T {
	return &v
}
