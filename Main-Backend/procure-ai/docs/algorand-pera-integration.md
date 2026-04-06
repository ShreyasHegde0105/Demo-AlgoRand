# Algorand + Pera Integration Plan

This document defines the backend integration shape for replacing the mocked payment flow in `services/blockchain_service.go`.

## Goal

Keep recommendation and order logic in Go, but move escrowed money movement onto Algorand.

## Wallet Model

Use Pera Wallet for user signing.

- buyer signs the funding transaction
- buyer signs the delivery confirmation and release call
- backend does not hold the buyer's private key
- deployer/admin account can deploy the app and optionally trigger refunds

For the hackathon, this is the simplest secure story.

## Contract Model

Use one Algorand application per order.

Each order should store:

- `algorandAppId`
- `algorandAppAddress`
- `fundingTxId`
- `releaseTxId`
- `refundTxId`
- `algorandNetwork`

## Backend Changes Needed

### 1. Order model

Extend [models/order.go](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/Main-Backend/procure-ai/models/order.go) with:

- `AlgorandAppID uint64`
- `AlgorandAppAddress string`
- `RefundTxID string`
- `AlgorandNetwork string`

### 2. Blockchain service

Replace the mock implementation in [services/blockchain_service.go](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/Main-Backend/procure-ai/services/blockchain_service.go) with an Algorand-backed service that does two kinds of work:

1. server-side actions
2. wallet-prepared actions

Server-side actions:

- deploy or create the order app
- lookup app state
- verify on-chain confirmations

Wallet-prepared actions:

- prepare grouped payment + app call for `fund`
- prepare app call for `confirm_delivery`
- prepare app call for `release_payment`

## Recommended API Shape

Keep current lifecycle endpoints, but split wallet preparation from final status updates.

Suggested new endpoints:

- `POST /blockchain/create-escrow`
- `POST /blockchain/prepare-fund`
- `POST /blockchain/confirm-fund`
- `POST /blockchain/prepare-release`
- `POST /blockchain/confirm-release`

## Recommended Flow

### A. Create escrow

After order approval:

1. backend deploys or creates the escrow app for the order
2. backend stores `appId` and `appAddress`
3. frontend receives escrow metadata

### B. Fund escrow

1. frontend asks backend for prepared transactions
2. backend builds:
   - payment txn from buyer to app address
   - app call txn to `fund`
3. frontend sends those txns to Pera Wallet
4. buyer signs in Pera
5. signed txns are submitted
6. backend confirms and updates order status to `funds_locked`

### C. Release payment

1. frontend asks backend for prepared release call
2. buyer signs with Pera
3. contract sends inner payment to seller
4. backend confirms and updates order status to `payment_released`

## SDK Responsibilities

Backend should use Algorand SDK for:

- suggested params
- app create / app call transaction building
- tx confirmation
- state reads

Frontend should use Pera for:

- wallet connection
- signing buyer-controlled transactions

## Environment Variables

Add these to the backend:

- `ALGORAND_NETWORK=testnet`
- `ALGORAND_ALGOD_ADDRESS=...`
- `ALGORAND_ALGOD_TOKEN=...`
- `ALGORAND_INDEXER_ADDRESS=...`
- `ALGORAND_INDEXER_TOKEN=...`
- `ALGORAND_ESCROW_APP_ID=...` only if you move to a shared app model later

Add these to the deploy environment:

- `DEPLOYER_MNEMONIC=...`
- `DEPLOYER_SENDER=...` if you standardize account loading differently

## Notes

- Pera Wallet does not require a browser extension for this flow; WalletConnect/Pera Connect is sufficient.
- For hackathon speed, one-app-per-order is easier than a shared app with boxes.
