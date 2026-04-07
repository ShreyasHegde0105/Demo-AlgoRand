# Procure AI Workspace

Workspace for the Procure AI backend and Algorand escrow contract.

## Projects

- `procure-ai`: Go API for supplier recommendation, orders, QR flow, and blockchain orchestration
- `procure-contracts`: AlgoPy ARC-4 escrow contract, compiled artifacts, and Python helper CLI

## End-to-End Flow

1. AI ranks suppliers
2. User creates an order from the shortlist
3. Order is approved off-chain
4. Backend deploys an Algorand escrow app
5. Agent selects supplier on-chain
6. Approver approves the escrow
7. Buyer funds the escrow
8. Buyer confirms delivery
9. Approver releases payment

## Quick Start

### 1. Start the contract project

```powershell
cd C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/Main-Backend/procure-contracts
.venv/Scripts/python.exe scripts/compile_contract.py
```

### 2. Start the backend

```powershell
cd C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/Main-Backend/procure-ai
go run .
```

Backend URL:

```text
http://localhost:8080
```

## Notes

- The blockchain flow depends on the compiled ARC-56 artifact in `procure-contracts/artifacts`.
- Wallet signing happens outside the Go API; the backend prepares transactions and later confirms state.
- The frontend UI lives in `../procure-ui`.
