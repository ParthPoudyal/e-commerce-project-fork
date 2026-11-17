# Auction Studio Backend

High-performance Go backend for **Auction Studio**, an online auction and marketplace for antiques and collectibles.

## Tech Stack

- **Language:** Go 1.25
- **Database:** PostgreSQL 16 (with connection pooling via `pgxpool`)
- **Driver & Migration:** `pgx/v5`, `golang-migrate`
- **Routing & Middleware:** `go-chi/chi/v5`, `go-chi/cors`
- **Authentication:** JWT (HMAC-SHA256) + Argon2id/Bcrypt password hashing + Refresh Tokens
- **Real-Time:** WebSockets (`gorilla/websocket`) for live bid updates and auction expiration events
- **Payment Gateway:** Khalti payment integration with provider abstraction interface (extensible for eSewa)
- **Currency:** Nepalese Rupee (NPR) with integer arithmetic

---

## Directory Structure

```text
Backend/
├── cmd/
│   └── server/          # Main application entrypoint
├── docs/
│   └── openapi.yaml     # OpenAPI 3.0 specification
├── internal/
│   ├── auth/            # JWT issuance, claims, password hashing
│   ├── config/          # Environment configuration loading
│   ├── database/        # Connection pool, migrations, seed data (40 NPR items)
│   ├── dto/             # API request and response data transfer objects
│   ├── handler/         # HTTP handlers (auth, items, bids, cart, orders, etc.)
│   ├── middleware/      # JWT auth, role authorization, CORS
│   ├── models/          # Domain models
│   ├── payment/         # Payment provider interface & Khalti implementation
│   ├── repository/      # Database queries with PostgreSQL locking
│   ├── service/         # Business logic and auction expiration worker
│   └── ws/              # Real-time WebSocket Hub
├── uploads/             # Static file storage for item uploads
├── Dockerfile           # Multi-stage production container build
├── .env.example         # Environment template
└── go.mod               # Dependencies
```

---

## Quickstart & Installation

### 1. With Docker Compose (Recommended)

From the project root:

```bash
docker compose up --build -d
```

This automatically orchestrates:
- **PostgreSQL 16** at `localhost:5432`
- **Go Backend API** at `localhost:8080` (with real-time WebSockets at `/ws`)
- **React Frontend** at `localhost:5173`

Embedded migrations and initial seed data (~40 NPR antique lots + pre-configured test users) run automatically on boot.

---

### 2. Running Locally (Direct Host Execution)

#### Prerequisites
- Go 1.22+ installed
- PostgreSQL 14+ running locally (e.g., on `localhost:5432`)

#### Database Credentials & `.env` Setup

Copy `.env.example` to `.env` in the `Backend/` directory:

```bash
cd Backend
cp .env.example .env
```

The configuration loader automatically parses `.env` on startup and supports **either** a complete connection string or individual database parameters:

**Option A — Connection String (`DATABASE_URL` or `DB_DSN`):**
```ini
DATABASE_URL=postgres://postgres:parth%20grey1@localhost:5432/ecom_db?sslmode=disable
```

**Option B — Individual Variables:**
```ini
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ecom_db
DB_USER=postgres
DB_PASSWORD=parth grey1
```
*(Note: Passwords containing spaces or special characters are automatically URL-encoded by the configuration loader).*

#### Start the Server

```bash
go run ./cmd/server
```

You should see:
```text
Starting Auction Studio Backend on port 8080 (env: development)...
Connected to PostgreSQL
Migrations at version 1 (dirty: false)
Seeding complete. 40 items in NPR ready.
Server listening on http://localhost:8080
```

---

## Key Technical Features

### 1. Auction Concurrency & Race Condition Prevention
Bid placement uses PostgreSQL row-level locking:
```sql
SELECT ... FROM items WHERE id = $1 FOR UPDATE OF i
```
This guarantees that concurrent bids cannot overwrite each other or accept bids lower than the active increment.

### 2. Bid Cancellation
Bids can be cancelled prior to auction close:
- Soft-deletes the bid (`status = 'cancelled'`)
- Concurrently locks the item
- Recalculates the highest remaining active bid and decrements the active bid count
- Broadcasts the new current price to all connected WebSocket clients

### 3. Background Auction Expiration Worker
An automated background worker checks every 10 seconds for ended auctions:
- Updates status to `ended`
- Resolves the winning bidder
- Emits real-time `AUCTION_ENDED` WebSocket notifications

### 4. Khalti Payment Gateway
- Initiates payment with Khalti API (`/epayment/initiate/`)
- Verifies payment status via server-to-server lookup (`/epayment/lookup/`)
- In local development mode without secret keys, operates seamlessly in sandbox mock mode for continuous testing.

---

## Test Accounts

The database comes pre-seeded with:

| Email | Password | Role |
| :--- | :--- | :--- |
| `admin@auctionstudio.com` | `admin123` | Administrator |
| `seller@auctionstudio.com` | `seller123` | Verified Seller |
| `jane@example.com` | `password123` | Buyer |
| `buyer@auctionstudio.com` | `buyer123` | Buyer |
