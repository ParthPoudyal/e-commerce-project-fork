# Auction Studio

Full-stack online auction and marketplace for antiques and collectibles.

- **Frontend:** React 18, TypeScript, Tailwind CSS, Vite (McMaster-Carr + Swiss design)
- **Backend:** Go 1.25, PostgreSQL 16, `pgx/v5`, `go-chi/chi/v5`, WebSockets
- **Currency & Payments:** Nepalese Rupee (NPR), Khalti payment gateway

---

## Quickstart

### 1. Docker Compose (All Services)

```bash
docker compose up --build -d
```

- **Frontend:** http://localhost:5173
- **Backend API:** http://localhost:8080/api/v1
- **WebSocket:** ws://localhost:8080/ws
- **PostgreSQL:** localhost:5432

### 2. Local Host Setup

#### Backend (`Backend/`)
1. Ensure PostgreSQL is running and database (e.g. `ecom_db`) exists.
2. Configure `.env` in `Backend/` (supports `DATABASE_URL`, `DB_DSN`, or individual `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`).
3. Run:
   ```bash
   cd Backend
   go run ./cmd/server
   ```

#### Frontend (`Frontend/`)
1. Run:
   ```bash
   cd Frontend
   npm install
   npm run dev
   ```
2. Open `http://localhost:5173`.

---

## Pre-Seeded Accounts

- `jane@example.com` / `password123` (Buyer)
- `buyer@auctionstudio.com` / `buyer123` (Buyer)
- `seller@auctionstudio.com` / `seller123` (Seller)
- `admin@auctionstudio.com` / `admin123` (Admin)

For detailed documentation, see [Backend/README.md](file:///c:/Users/acer/Desktop/DEERWALK%20STUFF/SEM%206/AuctionStudio/e-commerce-project/Backend/README.md) and [Frontend/README.md](file:///c:/Users/acer/Desktop/DEERWALK%20STUFF/SEM%206/AuctionStudio/e-commerce-project/Frontend/README.md).
