# Procure Contracts

Algorand smart-contract project for the Procure AI demo.

## What It Contains

- AlgoPy ARC-4 escrow contract
- compiled ARC-56 app spec and TEAL artifacts
- helper CLI used by the Go backend
- deploy and compile scripts

## Contract Purpose

The escrow contract enforces the commercial trust layer for a single procurement order.

It stores and enforces:

- buyer, seller, agent, and approver roles
- escrow amount and quote expiry
- supplier selection
- approval state
- funding state
- delivery confirmation
- release or refund outcome

## Contract Lifecycle

1. `created`
2. `supplier_selected`
3. `approved`
4. `funded`
5. `delivered`
6. `released`
7. or `refunded`

## Security Highlights

- strict state transitions
- supplier address must match the registered seller
- quoted amount must match the escrow amount
- timeout-aware refund path
- on-chain audit fields for funding, approval, delivery, release, and refund

## Compile

```powershell
cd C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/Main-Backend/procure-contracts
.venv/Scripts/python.exe scripts/compile_contract.py
```

Generated files:

- `artifacts/procure_escrow.arc56.json`
- `artifacts/procure_escrow.approval.teal`
- `artifacts/procure_escrow.clear.teal`

## Helper CLI

The Go backend uses `scripts/escrow_cli.py` to:

- create escrow apps
- prepare supplier-selection transactions
- prepare approval transactions
- prepare funding transactions
- prepare release transactions
- read app state

## Notes

- This project uses a one-app-per-order model.
- More contract state improves security and auditability, but also increases Algorand minimum balance requirements for app creation.
