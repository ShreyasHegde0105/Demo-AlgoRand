# Procure AI Backend

Backend service for an autonomous procurement workflow built with Go, Gin, GORM, and PostgreSQL.

This backend does four main jobs:

1. stores supplier data
2. runs an agent-style vendor recommendation flow
3. creates and manages procurement orders
4. simulates payment lock/release and delivery confirmation

The current system is backend-first and hackathon-friendly:

- supplier quotes are mocked through seeded vendor data
- vendor recommendation is rule-based, not LLM-based
- payment/blockchain calls are mocked but already follow a clean `txID` pattern

## What The System Does

The current procurement flow is:

1. user submits procurement requirements
2. backend agent evaluates vendors and returns a ranked shortlist
3. frontend shows top vendors to the user
4. user selects one vendor from the shortlist
5. backend creates an order and stores decision metadata
6. order moves through approval, fund lock, QR verification, delivery, and payment release

This is important:

- the agent recommends vendors
- the user still chooses the vendor
- the backend validates that the chosen vendor was actually part of the saved shortlist

So the agent acts like a procurement advisor, not an auto-buyer.

## Tech Stack

- Go 1.23+
- Gin
- GORM
- PostgreSQL
- `go-qrcode`

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

Important folders:

- [controllers](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/Main-Backend/procure-ai/controllers)
- [db](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/Main-Backend/procure-ai/db)
- [models](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/Main-Backend/procure-ai/models)
- [routes](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/Main-Backend/procure-ai/routes)
- [services](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/Main-Backend/procure-ai/services)

## Current Backend Capabilities

### Vendor and Quote Layer

- stores vendor data in PostgreSQL
- seeds 20 mock vendors on startup if they do not already exist
- supports multiple vendor attributes such as:
  - price
  - trust
  - delivery days
  - stock
  - minimum order quantity
  - location
  - payment terms
  - reliability score
  - category

### Agent Recommendation Layer

- filters vendors using procurement constraints
- scores eligible vendors
- returns ranked vendors
- stores recommendation sessions in DB
- gives each recommendation a `recommendationId`

### Order Layer

- creates orders only from a saved recommendation shortlist
- stores agent metadata with the order
- starts new orders in `pending_approval`

### Payment and Delivery Layer

- approves orders
- locks funds with a mock blockchain transaction id
- generates and verifies QR codes
- confirms delivery
- releases payment with a mock blockchain transaction id

## Database Setup

Current PostgreSQL connection is defined in [database.go](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/Main-Backend/procure-ai/db/database.go).

Current values:

- host: `localhost`
- port: `5432`
- user: `postgres`
- password: `LowKey7642`
- db name: `procure_ai`

The DSN is currently hardcoded in source, so if your local database config is different, update [database.go](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/Main-Backend/procure-ai/db/database.go) before running the app.

## Startup Behavior

On startup, the backend does the following:

1. connects to PostgreSQL
2. runs auto-migration
3. seeds vendors if they are missing
4. starts the Gin server on port `8080`

Current migrated models include:

- vendors
- orders
- qr records
- recommendation sessions

## Run Locally

### Option 1: normal Go commands

```bash
go mod tidy
go run .
```

### Option 2: if Go build cache permissions cause issues

Use a workspace-local cache:

```powershell
$env:GOCACHE=(Join-Path (Get-Location) '.gocache')
& 'C:\Program Files\Go\bin\go.exe' run .
```

Server URL:

```text
http://localhost:8080
```

## API Overview

Current routes from [routes.go](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/Main-Backend/procure-ai/routes/routes.go):

- `GET /vendors`
- `POST /select-vendor`
- `POST /agent/recommend-vendors`
- `POST /create-order`
- `POST /approve-order`
- `POST /lock-funds`
- `POST /release-payment`
- `POST /generate-qr`
- `POST /verify-qr`
- `POST /confirm-delivery`

## Recommended Product Flow

This is the intended frontend/backend flow:

1. call `POST /agent/recommend-vendors`
2. show top 5 vendors to the user
3. if user wants more options, call the same route again with larger `topN`
4. user selects one vendor from the shortlist
5. call `POST /create-order`
6. order is created as `pending_approval`
7. continue with approval and payment lifecycle

## Core Concepts

### 1. Recommendation Session

When `POST /agent/recommend-vendors` is called:

- the backend computes the shortlist
- the shortlist is saved in DB
- the response includes `recommendationId`

That `recommendationId` must later be used when creating the order.

### 2. Shortlist Validation

When `POST /create-order` is called:

- backend checks the saved recommendation session
- backend confirms the chosen vendor exists in that shortlist
- backend rejects vendors that were never recommended for that session

This prevents bypassing the agent flow.

### 3. Order Metadata

Orders store agent decision metadata such as:

- `recommendationId`
- `selectionReason`
- `agentScore`
- `shortlistSnapshot`

This makes the order explainable and audit-friendly.

## Order Status Flow

Current state progression:

1. `pending_approval`
2. `approved`
3. `funds_locked`
4. `delivered`
5. `payment_released`

Rules:

- `lock-funds` requires status `approved`
- `confirm-delivery` requires status `funds_locked`
- `release-payment` requires status `delivered`
- `confirm-delivery` currently also triggers payment release internally

## Detailed API Guide

### 1. Get Vendors

`GET /vendors`

Purpose:

- fetch all vendors currently stored in DB
- useful for inspection and debugging

Example response:

```json
{
  "vendors": [
    {
      "id": 1,
      "name": "Alpha Industrial Supply",
      "price": 94.5,
      "trust": 4.8,
      "deliveryDays": 2,
      "stock": 450,
      "minOrderQty": 20,
      "location": "Mumbai",
      "paymentTerms": "Net 15",
      "reliabilityScore": 96,
      "category": "electronics"
    }
  ]
}
```

### 2. Basic Manual Vendor Selection

`POST /select-vendor`

Purpose:

- legacy/basic scoring endpoint
- works on raw vendor input sent directly in request body

Example request:

```json
{
  "vendors": [
    {
      "name": "Vendor A",
      "price": 95,
      "trust": 4.5
    },
    {
      "name": "Vendor B",
      "price": 100,
      "trust": 4.8
    }
  ]
}
```

This route is not the main agent workflow anymore.

### 3. Agent Recommendation

`POST /agent/recommend-vendors`

Purpose:

- compute ranked vendor shortlist based on procurement requirements
- save shortlist as a recommendation session
- return `recommendationId`

Example request:

```json
{
  "category": "electronics",
  "quantity": 40,
  "budget": 4000,
  "maxDeliveryDays": 5,
  "preferredCities": ["Mumbai", "Delhi", "Pune"],
  "topN": 5
}
```

Example response shape:

```json
{
  "recommendationId": "REC-0001",
  "recommendedVendor": {
    "rank": 1,
    "vendor": {
      "name": "Alpha Industrial Supply"
    }
  },
  "topVendors": [],
  "rejectedVendors": [],
  "appliedWeights": {
    "price": 0.35,
    "delivery": 0.25,
    "trust": 0.2,
    "reliability": 0.2
  },
  "summary": "Ranked 2 eligible vendors and shortlisted the top 2 for category \"electronics\"."
}
```

Use `topN` for frontend behavior:

- `topN: 5` for initial shortlist
- `topN: 10` or more for "Show more"

### 4. Create Order From User Selection

`POST /create-order`

Purpose:

- create an order only after user selects a vendor from the saved shortlist

Required request:

```json
{
  "recommendationId": "REC-0001",
  "vendor": "Vertex Trade Links",
  "quantity": 40
}
```

Important behavior:

- `recommendationId` must exist
- `vendor` must be part of that recommendation's shortlist
- `quantity` must match the recommendation session quantity

What backend stores on the order:

- chosen vendor
- category
- quantity
- unit price
- total amount
- `recommendationId`
- `selectionReason`
- `agentScore`
- `shortlistSnapshot`
- initial status `pending_approval`

### 5. Approve Order

`POST /approve-order`

Purpose:

- move order from `pending_approval` to `approved`

Request:

```json
{
  "orderId": "ORD-0001"
}
```

### 6. Lock Funds

`POST /lock-funds`

Purpose:

- simulate blockchain fund lock
- store mock transaction id in the order's `paymentTxId` field

Request:

```json
{
  "orderId": "ORD-0001"
}
```

Example response:

```json
{
  "orderId": "ORD-0001",
  "txID": "lock-ORD-0001-abcdef12",
  "status": "funds_locked"
}
```

### 7. Generate QR

`POST /generate-qr`

Purpose:

- generate QR data for order handoff or delivery verification

Request:

```json
{
  "orderId": "ORD-0001"
}
```

### 8. Verify QR

`POST /verify-qr`

Purpose:

- validate QR for the given order

Request:

```json
{
  "orderId": "ORD-0001",
  "qrCode": "PROCURE-ORDER:ORD-0001"
}
```

### 9. Confirm Delivery

`POST /confirm-delivery`

Purpose:

- first mark order as delivered
- then trigger payment release

Request:

```json
{
  "orderId": "ORD-0001"
}
```

Current behavior:

- briefly sets order to `delivered`
- then calls payment release logic
- final order status becomes `payment_released`

### 10. Release Payment

`POST /release-payment`

Purpose:

- explicit payment release endpoint
- currently used for mocked payment flow

Request:

```json
{
  "orderId": "ORD-0001"
}
```

## End-To-End Test Flow

Use this sequence in Postman:

### Step 1

`POST /agent/recommend-vendors`

```json
{
  "category": "electronics",
  "quantity": 40,
  "budget": 4000,
  "maxDeliveryDays": 5,
  "preferredCities": ["Mumbai", "Delhi", "Pune"],
  "topN": 5
}
```

Save the returned:

- `recommendationId`
- one vendor name from `topVendors`

### Step 2

`POST /create-order`

```json
{
  "recommendationId": "REC-0001",
  "vendor": "Vertex Trade Links",
  "quantity": 40
}
```

Save the returned `orderId`.

### Step 3

`POST /approve-order`

```json
{
  "orderId": "ORD-0001"
}
```

### Step 4

`POST /lock-funds`

```json
{
  "orderId": "ORD-0001"
}
```

### Step 5

`POST /generate-qr`

```json
{
  "orderId": "ORD-0001"
}
```

Copy the returned QR code.

### Step 6

`POST /verify-qr`

```json
{
  "orderId": "ORD-0001",
  "qrCode": "PROCURE-ORDER:ORD-0001"
}
```

### Step 7

`POST /confirm-delivery`

```json
{
  "orderId": "ORD-0001"
}
```

## What Is Mocked Right Now

The following parts are still mocked:

- vendor quote sourcing
- blockchain settlement
- transaction ids for lock/release

Current blockchain-style behavior is still useful because:

- backend already follows a clean integration pattern
- lock and release return `txID`
- order stores payment transaction information

Later, the internal implementation of the blockchain service can be replaced with a real Algorand integration without redesigning the whole backend flow.

## Why We Still Use PostgreSQL

Even if a blockchain layer is added later, PostgreSQL is still needed for normal application data.

Use DB for:

- vendors
- recommendation sessions
- shortlist snapshots
- orders
- approval state
- QR records
- audit metadata

Use blockchain later for:

- fund lock
- fund release
- transaction proof

Database is the operational application store.
Blockchain is the settlement/trust layer.

## Build And Test

If Go is on PATH:

```bash
go build ./...
go test ./...
```

If Go cache permissions are an issue:

```powershell
$env:GOCACHE=(Join-Path (Get-Location) '.gocache')
& 'C:\Program Files\Go\bin\go.exe' test ./...
```

## Current Limitations

- no authentication
- no user roles
- no pagination on vendor/order data
- no real blockchain integration yet
- no real supplier quote aggregation yet
- no frontend session management in this repo

## Summary

This backend is now set up for a solid hackathon procurement demo:

- vendors are seeded and queryable
- the agent returns a ranked shortlist
- the shortlist is stored with a recommendation id
- the user can choose from recommended vendors
- the backend validates that choice
- the order stores decision metadata
- the order lifecycle is enforced through status transitions

That is a strong base for frontend integration and later blockchain replacement.
