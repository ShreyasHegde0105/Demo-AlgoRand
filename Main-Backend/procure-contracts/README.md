# Procure Contracts

Algorand smart contracts for the Procure AI demo.

This folder is intentionally separate from the Go backend in `../procure-ai`.

## Scope

The first contract replaces the mocked payment flow in the backend with an escrow-style application that supports:

- creating an escrow for an order
- funding an order
- confirming delivery
- releasing payment
- refunding if needed

## Structure

```text
procure-contracts/
  smart_contracts/
    escrow_contract/
      contract.py
  tests/
  pyproject.toml
```

## Python Setup

Install the contract dependencies from this directory:

```powershell
python -m pip install -e .
```

Use the official Algorand Python package:

- `algorand-python`
- `algokit-utils`
- `py-algorand-sdk`
- `puyapy`

## Compile Contract

Regenerate the ARC-56 app spec and TEAL outputs before deployment whenever the ABI changes:

```powershell
.venv\Scripts\python.exe scripts\compile_contract.py
```

The backend helper expects:

- `artifacts/procure_escrow.arc56.json`

## Contract Helper CLI

The Go backend now shells into [scripts/escrow_cli.py](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/Main-Backend/procure-contracts/scripts/escrow_cli.py) to bridge into AlgoKit and the Algorand SDK.

Supported commands:

- `create-escrow`
- `prepare-fund`
- `prepare-select-supplier`
- `prepare-approve`
- `prepare-release`
- `get-state`

These commands are designed for the one-app-per-order flow:

1. create the escrow app for a specific order
2. prepare unsigned funding transactions for the buyer wallet
3. prepare unsigned supplier-selection transaction for the agent wallet
4. prepare unsigned approval transaction for the approver wallet
5. prepare unsigned delivery confirmation + release transactions
6. read global state to verify that funding, approval, selection, or release completed on-chain

## Deployment

[deploy.py](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/Main-Backend/procure-contracts/deploy.py) now creates a per-order app instance using the ARC-4 `create_application` method.

Required environment variables:

- `ALGOD_SERVER`
- `ALGOD_TOKEN`
- `DEPLOYER_MNEMONIC`
- `PROCURE_ESCROW_BUYER`
- `PROCURE_ESCROW_SELLER`
- `PROCURE_ESCROW_AGENT`
- `PROCURE_ESCROW_APPROVER`
- `PROCURE_ESCROW_AMOUNT_MICROALGOS`
- `PROCURE_ESCROW_QUOTE_VALID_UNTIL`

Optional environment variables:

- `PROCURE_ESCROW_APP_NAME`
- `PROCURE_ESCROW_ORDER_ID`
- `ALGORAND_NETWORK`

## Contract Model

The current contract is designed as one app instance per order. That keeps the hackathon story simple:

- create app for order `ORD-xxxx`
- agent stores the selected supplier and quote id
- approver marks the procurement approved
- buyer funds the app account
- buyer confirms delivery
- approver releases payment to seller
- app stores settlement supplier, amount, and transaction id on-chain

This avoids needing box maps or a shared multi-order app on day one.

For demo use, `PROCURE_ESCROW_AMOUNT_MICROALGOS=10000` is recommended so funding uses only `0.01 ALGO`.
