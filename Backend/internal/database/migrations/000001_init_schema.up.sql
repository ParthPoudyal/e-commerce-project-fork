-- Enable UUID extension if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'buyer',
    phone VARCHAR(64) DEFAULT '',
    location VARCHAR(255) DEFAULT '',
    rating NUMERIC(3,2) DEFAULT 5.00,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Refresh tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);

-- Categories
CREATE TABLE IF NOT EXISTS categories (
    id VARCHAR(64) PRIMARY KEY,
    label VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    parent_id VARCHAR(64) REFERENCES categories(id) ON DELETE CASCADE
);

-- Items
CREATE TABLE IF NOT EXISTS items (
    id VARCHAR(64) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    lot_number VARCHAR(64) NOT NULL,
    category_slug VARCHAR(255) NOT NULL,
    subcategory VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    era VARCHAR(255) DEFAULT '',
    origin VARCHAR(255) DEFAULT '',
    material VARCHAR(255) DEFAULT '',
    dimensions VARCHAR(255) DEFAULT '',
    condition VARCHAR(64) NOT NULL DEFAULT 'Good',
    provenance TEXT DEFAULT '',
    images JSONB NOT NULL DEFAULT '[]'::jsonb,
    purchase_mode VARCHAR(32) NOT NULL, -- 'auction' | 'buy-now'
    price BIGINT, -- for buy-now items
    stock INT DEFAULT 1,
    current_bid BIGINT,
    starting_bid BIGINT,
    minimum_increment BIGINT DEFAULT 100,
    bid_count INT DEFAULT 0,
    end_time TIMESTAMPTZ,
    status VARCHAR(32) NOT NULL DEFAULT 'active', -- 'active', 'ended', 'sold', 'cancelled'
    seller_id VARCHAR(64) NOT NULL REFERENCES users(id),
    shipping_cost BIGINT NOT NULL DEFAULT 500,
    shipping_estimated_days VARCHAR(128) DEFAULT '3-5 business days',
    shipping_methods JSONB NOT NULL DEFAULT '["Standard parcel"]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_items_category ON items(category_slug);
CREATE INDEX IF NOT EXISTS idx_items_purchase_mode ON items(purchase_mode);
CREATE INDEX IF NOT EXISTS idx_items_status ON items(status);
CREATE INDEX IF NOT EXISTS idx_items_end_time ON items(end_time);

-- Bids
CREATE TABLE IF NOT EXISTS bids (
    id VARCHAR(64) PRIMARY KEY,
    item_id VARCHAR(64) NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active', -- 'active', 'cancelled', 'won', 'outbid'
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_bids_item_id ON bids(item_id);
CREATE INDEX IF NOT EXISTS idx_bids_user_id ON bids(user_id);
CREATE INDEX IF NOT EXISTS idx_bids_status ON bids(status);

-- Cart Items
CREATE TABLE IF NOT EXISTS cart_items (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id VARCHAR(64) NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, item_id)
);

-- Wishlist Items
CREATE TABLE IF NOT EXISTS wishlist_items (
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id VARCHAR(64) NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, item_id)
);

-- Orders
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id),
    status VARCHAR(32) NOT NULL DEFAULT 'pending', -- 'pending', 'paid', 'shipped', 'delivered', 'cancelled'
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(64) NOT NULL,
    line1 VARCHAR(500) NOT NULL,
    line2 VARCHAR(500) DEFAULT '',
    city VARCHAR(255) NOT NULL,
    state VARCHAR(255) NOT NULL,
    zip VARCHAR(64) NOT NULL,
    country VARCHAR(255) NOT NULL,
    subtotal BIGINT NOT NULL,
    shipping_cost BIGINT NOT NULL,
    buyers_premium BIGINT NOT NULL,
    total BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);

-- Order Items
CREATE TABLE IF NOT EXISTS order_items (
    id VARCHAR(64) PRIMARY KEY,
    order_id VARCHAR(64) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    item_id VARCHAR(64) NOT NULL REFERENCES items(id),
    title VARCHAR(500) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    price BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Payments
CREATE TABLE IF NOT EXISTS payments (
    id VARCHAR(64) PRIMARY KEY,
    order_id VARCHAR(64) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    provider VARCHAR(64) NOT NULL DEFAULT 'khalti',
    pidx VARCHAR(255) DEFAULT '',
    provider_txn_id VARCHAR(255) DEFAULT '',
    amount BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'initiated', -- 'initiated', 'completed', 'failed', 'refunded'
    raw_response JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_pidx ON payments(pidx);
