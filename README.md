# Procure AI Backend

Gin + GORM backend for an autonomous procurement workflow with:

- Vendor discovery and scoring
- Agent-based vendor recommendation
- Order lifecycle with approval and payment stages
- QR generation and verification for order handoff

## Tech Stack

- Go 1.23+
- Gin (HTTP API)
- GORM + PostgreSQL
- go-qrcode

## Project Structure

```text
procure-ai/
	controllers/   HTTP handlers
	db/            PostgreSQL connection, migration, seed
	models/        Request/response and DB models
	routes/        API route mapping
	services/      Business logic
	main.go        App bootstrap
```

## Prerequisites

- Go installed
- PostgreSQL running locally
- Database named `procure_ai`

Current DB connection is hardcoded in [db/database.go](db/database.go):

- host: localhost
- port: 5432
- user: postgres
- db: procure_ai

If your local setup is different, update [db/database.go](db/database.go) before starting.

## Run Locally

```bash
go mod tidy
go run main.go
```

Server starts on:

```text
http://localhost:8080
```

On startup, the app will:

1. Connect to PostgreSQL
2. Auto-migrate `vendors`, `orders`, `qrs`
3. Seed vendor data

## Order Lifecycle

Valid state progression:

1. `pending_approval` (after create)
2. `approved`
3. `funds_locked`
4. `delivered`
5. `payment_released`

Important behavior:

- `POST /lock-funds` requires order status `approved`
- `POST /release-payment` requires order status `delivered`
- `POST /confirm-delivery` moves to delivered and immediately releases payment

## API Endpoints

### Health/Discovery

- `GET /vendors`

Response:

```json
{
	"vendors": []
}
```

### Vendor Selection

- `POST /select-vendor`

Request:

```json
{
	"vendors": [
		{ "name": "Alpha", "price": 95, "trust": 4.6 }
	]
}
```

### Agent Recommendation

- `POST /agent/recommend-vendors`

Request:

```json
{
	"category": "electronics",
	"quantity": 120,
	"budget": 12000,
	"maxDeliveryDays": 5,
	"preferredCities": ["Mumbai", "Pune"],
	"topN": 5
}
```

### Orders and Payment

- `POST /create-order`

Request:

```json
{
	"vendor": "Alpha Industrial Supply",
	"quantity": 50
}
```

- `POST /approve-order`

```json
{ "orderId": "ORD-0001" }
```

- `POST /lock-funds`

```json
{ "orderId": "ORD-0001" }
```

- `POST /confirm-delivery`

```json
{ "orderId": "ORD-0001" }
```

- `POST /release-payment`

```json
{ "orderId": "ORD-0001" }
```

### QR Operations

- `POST /generate-qr`

```json
{ "orderId": "ORD-0001" }
```

- `POST /verify-qr`

```json
{
	"orderId": "ORD-0001",
	"qrCode": "PROCURE-ORDER:ORD-0001"
}
```

## Quick Demo Flow

1. `GET /vendors`
2. `POST /agent/recommend-vendors`
3. `POST /create-order`
4. `POST /approve-order`
5. `POST /lock-funds`
6. `POST /generate-qr`
7. `POST /verify-qr`
8. `POST /confirm-delivery`

## Build and Test

```bash
go build ./...
go test ./...
```

## Notes

- Current payment/blockchain service stores transaction IDs in DB and generates mock IDs.
- This is suitable for local development demos, not production-grade on-chain settlement.
