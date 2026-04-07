# Procure AI Backend

Go API for the Procure AI demo. It handles supplier recommendation, order management, QR flow, and orchestration of the Algorand escrow contract.

## What It Does

- recommend suppliers from a procurement request
- store recommendation sessions
- create and approve orders
- create and manage Algorand escrow apps
- prepare unsigned wallet transactions for on-chain actions
- confirm on-chain state and sync backend records
- generate and verify QR codes for delivery flow

## Main Services

- `services/agent_service.go`: supplier ranking and recommendation sessions
- `services/order_service.go`: order lifecycle and persistence
- `services/procurement_service.go`: end-to-end procurement workflow
- `services/blockchain_service.go`: contract helper integration

## Order Flow

1. `pending_approval`
2. `approved`
3. `funds_locked`
4. `delivered`
5. `payment_released`

## Blockchain Flow

1. create escrow app
2. prepare supplier selection
3. confirm supplier selection
4. prepare approval
5. confirm approval
6. prepare funding
7. confirm funding
8. confirm delivery
9. prepare release
10. confirm release

## Run Locally

```powershell
cd C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/Main-Backend/procure-ai
go mod tidy
go run .
```

Server URL:

```text
http://localhost:8080
```

## Requirements

- PostgreSQL running locally
- compiled contract artifact available in `../procure-contracts/artifacts`
- Algorand environment loaded when using blockchain routes
- optional `GEMINI_API_KEY` for natural-language procurement parsing

## Important Notes

- `prepare-*` routes return unsigned transactions for the correct wallet.
- `confirm-*` routes read on-chain state and update backend records.
- Legacy `lock-funds` and `release-payment` routes are mock compatibility paths and should not be used for Algorand escrow orders.
