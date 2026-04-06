# Procure AI Backend

Backend service for an autonomous procurement workflow built with Go, Gin, GORM, PostgreSQL, and an Algorand ARC-4 escrow contract.

The current system supports:

1. supplier recommendation and shortlist storage
2. order creation from a validated shortlist
3. off-chain order approval
4. on-chain escrow creation, funding, approval, release, and settlement logging
5. QR generation and verification for delivery flow

## Current Lifecycle

The implemented procurement flow is:

1. user submits procurement requirements
2. backend ranks vendors and stores a recommendation session
3. user selects a vendor from the saved shortlist
4. backend creates an order in `pending_approval`
5. human approves the order
6. backend creates an Algorand escrow app for that order
7. agent stores `selectedSupplier` and `quoteId` on-chain
8. approver marks the escrow approved on-chain
9. buyer funds escrow
10. buyer confirms delivery
11. approver releases payment
12. backend verifies settlement state and stores settlement metadata

Important:

- the agent recommends suppliers, but the user still chooses the supplier off-chain
- the smart contract enforces agent and approver roles on-chain
- release only succeeds after approval and before quote expiry

## Tech Stack

- Go 1.23+
- Gin
- GORM
- PostgreSQL
- `go-qrcode`
- Algorand helper bridge via Python in `../procure-contracts`

## Project Structure

```text
procure-ai/
  controllers/   HTTP handlers
  db/            database connection, migration, seed logic
  models/        DB models and request/response models
  routes/        route registration
  services/      business logic
  main.go        application bootstrap
```

## Database Setup

Current PostgreSQL connection is defined in [database.go](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/main-backend/procure-ai/db/database.go).

Current values in source:

- host: `localhost`
- port: `5432`
- user: `postgres`
- password: `LowKey7642`
- db name: `procure_ai`

The DSN is currently hardcoded, so update [database.go](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/main-backend/procure-ai/db/database.go) if your local database differs.

Auto-migrated models:

- vendors
- orders
- qr records
- recommendation sessions

## Run Locally

From [procure-ai](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/main-backend/procure-ai):

```powershell
go mod tidy
go run .
```

If Go cache permissions cause issues:

```powershell
$env:GOCACHE=(Join-Path (Get-Location) '.gocache')
go run .
```

Server URL:

```text
http://localhost:8080
```

Gemini parsing is optional. To use `POST /agent/parse-and-recommend`, set:

```powershell
$env:GEMINI_API_KEY="your_api_key"
```

## Contract Dependency

The blockchain flow depends on the helper in [escrow_cli.py](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/main-backend/procure-contracts/scripts/escrow_cli.py) and the compiled ARC-56 app spec in [procure_escrow.arc56.json](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/main-backend/procure-contracts/artifacts/procure_escrow.arc56.json).

Before using blockchain routes:

1. compile the contract in `../procure-contracts`
2. ensure Algod env vars are loaded into the shell that starts this backend

## API Overview

Routes from [routes.go](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/main-backend/procure-ai/routes/routes.go):

- `GET /vendors`
- `POST /select-vendor`
- `POST /agent/recommend-vendors`
- `POST /agent/parse-and-recommend`
- `POST /create-order`
- `POST /approve-order`
- `POST /lock-funds`
- `POST /release-payment`
- `POST /generate-qr`
- `POST /verify-qr`
- `POST /confirm-delivery`
- `POST /blockchain/create-escrow`
- `POST /blockchain/prepare-select-supplier`
- `POST /blockchain/confirm-select-supplier`
- `POST /blockchain/prepare-approve`
- `POST /blockchain/confirm-approve`
- `POST /blockchain/prepare-fund`
- `POST /blockchain/confirm-fund`
- `POST /blockchain/prepare-release`
- `POST /blockchain/confirm-release`

## Request Ownership

In normal usage, request bodies come from the client:

- frontend UI
- Postman during testing
- any API consumer

General rule:

- `prepare-*` routes build unsigned transactions for a wallet signer
- `confirm-*` routes sync backend state after the signed transaction is submitted on-chain

## Core Concepts

### Recommendation Session

When `POST /agent/recommend-vendors` runs:

- backend computes the shortlist
- shortlist is stored in DB
- response returns `recommendationId`

That `recommendationId` is required later for `POST /create-order`.

### Shortlist Validation

When `POST /create-order` runs:

- backend loads the saved recommendation session
- backend confirms the selected vendor is actually in the shortlist
- backend rejects vendors outside the saved recommendation

### Order Metadata

Orders store procurement and escrow metadata including:

- `recommendationId`
- `selectionReason`
- `agentScore`
- `shortlistSnapshot`
- `buyerAddress`
- `sellerAddress`
- `agentAddress`
- `approverAddress`
- `selectedSupplier`
- `quoteId`
- `quoteValidUntil`
- `escrowApproved`
- `fundingTxId`
- `releaseTxId`
- `settlementSupplier`
- `settlementAmount`
- `settlementTxId`

## Order Status Flow

Current status progression:

1. `pending_approval`
2. `approved`
3. `funds_locked`
4. `delivered`
5. `payment_released`

Rules:

- `approve-order` requires `pending_approval`
- `prepare-fund` requires `approved`
- `confirm-delivery` requires `funds_locked`
- `prepare-release` requires `funds_locked` or `delivered`

Notes:

- `confirm-delivery` now only marks the order as delivered
- payment release is no longer triggered automatically by `confirm-delivery`
- on-chain approval is tracked separately from off-chain order approval

## Main API Flow

### 1. Recommend Vendors

`POST /agent/recommend-vendors`

```json
{
  "category": "electronics",
  "quantity": 10,
  "budget": 50000,
  "maxDeliveryDays": 7,
  "preferredCities": ["Bangalore", "Mumbai"],
  "topN": 3
}
```

Save:

- `recommendationId`
- selected vendor name from `topVendors`

### 2. Create Order

`POST /create-order`

```json
{
  "recommendationId": "REC-0001",
  "vendor": "Acme Supplies",
  "quantity": 10
}
```

Save:

- `orderId`

### 3. Approve Order

`POST /approve-order`

```json
{
  "orderId": "ORD-0001"
}
```

### 4. Create Escrow

`POST /blockchain/create-escrow`

```json
{
  "orderId": "ORD-0001",
  "buyerAddress": "ACCOUNT_A_ADDRESS",
  "sellerAddress": "ACCOUNT_B_ADDRESS",
  "agentAddress": "ACCOUNT_A_ADDRESS",
  "approverAddress": "ACCOUNT_A_ADDRESS",
  "escrowAmountMicroAlgos": 1000000,
  "quoteValidUntil": 999999999
}
```

### 5. Prepare Supplier Selection

`POST /blockchain/prepare-select-supplier`

```json
{
  "orderId": "ORD-0001",
  "selectedSupplier": "Acme Supplies",
  "quoteId": "QUOTE-001"
}
```

### 6. Confirm Supplier Selection

After the signed transaction is submitted:

`POST /blockchain/confirm-select-supplier`

```json
{
  "orderId": "ORD-0001"
}
```

### 7. Prepare Escrow Approval

`POST /blockchain/prepare-approve`

```json
{
  "orderId": "ORD-0001"
}
```

### 8. Confirm Escrow Approval

`POST /blockchain/confirm-approve`

```json
{
  "orderId": "ORD-0001"
}
```

### 9. Prepare Fund

`POST /blockchain/prepare-fund`

```json
{
  "orderId": "ORD-0001"
}
```

### 10. Confirm Fund

`POST /blockchain/confirm-fund`

```json
{
  "orderId": "ORD-0001",
  "txID": "SUBMITTED_GROUP_TX_ID"
}
```

### 11. Confirm Delivery

`POST /confirm-delivery`

```json
{
  "orderId": "ORD-0001"
}
```

### 12. Prepare Release

`POST /blockchain/prepare-release`

```json
{
  "orderId": "ORD-0001"
}
```

### 13. Confirm Release

`POST /blockchain/confirm-release`

```json
{
  "orderId": "ORD-0001",
  "txID": "SUBMITTED_RELEASE_TX_ID"
}
```

This step also syncs:

- `settlementSupplier`
- `settlementAmount`
- `settlementTxId`

## Legacy Routes

These still exist for backward compatibility:

- `POST /lock-funds`
- `POST /release-payment`

They are mock-style flows and should not be used for Algorand escrow orders.

## QR Flow

QR routes remain off-chain:

- `POST /generate-qr`
- `POST /verify-qr`

They can be used alongside the blockchain flow for delivery and handoff UX.

## Build And Test

```powershell
go build ./...
go test ./...
```

If Go cache permissions are an issue:

```powershell
$env:GOCACHE=(Join-Path (Get-Location) '.gocache')
go test ./...
```

## Current Limitations

- no authentication
- no user/session authorization layer in the Go API
- DB DSN is hardcoded
- supplier quotes are still mocked through seeded vendor data
- Postman can prepare blockchain transactions, but wallet signing still happens outside Postman
- contract env loading must be present in the shell that starts the backend

## Summary

This backend now supports a full procurement-to-settlement demo:

- recommendation sessions are stored and validated
- orders carry explainable agent metadata
- Algorand escrow is role-aware
- agent and approver actions are enforced on-chain
- settlement details are stored both on-chain and in the backend
