# Main Backend

Workspace for the Procure AI backend and Algorand escrow contract.

Projects:

- [procure-ai](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/main-backend/procure-ai): Go API for recommendations, orders, QR flow, and blockchain orchestration
- [procure-contracts](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/main-backend/procure-contracts): AlgoPy ARC-4 escrow contract and Python helper CLI

## Flow

1. Recommend suppliers
2. Create order from shortlist
3. Approve order
4. Create escrow app
5. Agent sets selected supplier and quote id
6. Approver approves escrow
7. Buyer funds escrow
8. Buyer confirms delivery
9. Approver releases payment
10. Backend stores settlement metadata

## Quick Start

### 1. Configure Contract Env

From [procure-contracts](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/main-backend/procure-contracts):

```powershell
cd C:\Users\bhats\OneDrive\Desktop\Dr_Hannibal_Lecter\Projects\main-backend\procure-contracts
Copy-Item .env.example .env
notepad .env
```

Minimum `.env`:

```env
ALGOD_SERVER=https://testnet-api.algonode.cloud
ALGOD_TOKEN=
DEPLOYER_MNEMONIC=25-word algorand mnemonic
PROCURE_ESCROW_APP_NAME=procure-escrow
PROCURE_ESCROW_ORDER_ID=ORD-DEMO
PROCURE_ESCROW_BUYER=BUYER_ADDRESS
PROCURE_ESCROW_SELLER=SELLER_ADDRESS
PROCURE_ESCROW_AGENT=AGENT_ADDRESS
PROCURE_ESCROW_APPROVER=APPROVER_ADDRESS
PROCURE_ESCROW_AMOUNT_MICROALGOS=10000
PROCURE_ESCROW_QUOTE_VALID_UNTIL=999999999
ALGORAND_NETWORK=testnet
```

### 2. Compile And Deploy Contract

```powershell
Get-Content .env | ForEach-Object {
  if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
  $name, $value = $_ -split '=', 2
  Set-Item -Path "Env:$name" -Value $value
}

.venv\Scripts\python.exe scripts\compile_contract.py
.venv\Scripts\python.exe deploy.py
```

Backend helper artifact:

- [procure_escrow.arc56.json](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/main-backend/procure-contracts/artifacts/procure_escrow.arc56.json)

### 3. Start Backend

Important:

- load the same Algorand env vars into the shell that starts `procure-ai`

```powershell
Get-Content C:\Users\bhats\OneDrive\Desktop\Dr_Hannibal_Lecter\Projects\main-backend\procure-contracts\.env | ForEach-Object {
  if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
  $name, $value = $_ -split '=', 2
  Set-Item -Path "Env:$name" -Value $value
}

cd C:\Users\bhats\OneDrive\Desktop\Dr_Hannibal_Lecter\Projects\main-backend\procure-ai
go mod tidy
go run .
```

Backend URL:

```text
http://localhost:8080
```

## API Order

Use these routes in order:

1. `POST /agent/recommend-vendors`
2. `POST /create-order`
3. `POST /approve-order`
4. `POST /blockchain/create-escrow`
5. `POST /blockchain/prepare-select-supplier`
6. `POST /blockchain/confirm-select-supplier`
7. `POST /blockchain/prepare-approve`
8. `POST /blockchain/confirm-approve`
9. `POST /blockchain/prepare-fund`
10. `POST /blockchain/confirm-fund`
11. `POST /confirm-delivery`
12. `POST /blockchain/prepare-release`
13. `POST /blockchain/confirm-release`

Rule:

- `prepare-*` builds unsigned wallet transactions
- `confirm-*` verifies on-chain state and syncs backend records

## Minimal Demo Accounts

Fastest test setup:

- Account A: deployer + buyer + agent + approver
- Account B: seller

Demo amount:

- the backend now caps new escrow creation to `10000` microAlgos (`0.01 ALGO`) unless you send a smaller value
- this keeps demo funding cheap even when the business order amount is much larger off-chain

## Common Issues

`WrongMnemonicLengthError`

- `DEPLOYER_MNEMONIC` is not a valid 25-word Algorand mnemonic

`WinError 10061`

- wrong `ALGOD_SERVER`
- or backend started without Algorand env loaded

`Object of type bytes is not JSON serializable`

- already fixed in [escrow_cli.py](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/main-backend/procure-contracts/scripts/escrow_cli.py)

## More Detail

- [procure-ai/README.md](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/main-backend/procure-ai/README.md)
- [procure-contracts/README.md](C:/Users/bhats/OneDrive/Desktop/Dr_Hannibal_Lecter/Projects/main-backend/procure-contracts/README.md)
