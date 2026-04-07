# Procure UI

Frontend console for the Procure AI demo. This app drives the procurement flow from supplier discovery to escrow setup, on-chain actions, and QR verification.

## What It Does

- collect a procurement brief
- rank suppliers through the backend agent
- create and approve an order
- create an Algorand escrow app for that order
- prepare and confirm on-chain steps
- generate and verify delivery QR codes

## Stack

- React 19
- TypeScript
- Vite
- Tailwind CSS 4

## Main Flow

1. Discover suppliers
2. Create order
3. Approve order
4. Create escrow
5. Select supplier on-chain
6. Approve escrow
7. Fund escrow
8. Confirm delivery
9. Release payment

## Run Locally

```powershell
npm install
npm run dev
```

Backend API URL defaults to:

```text
http://localhost:8080
```

To override it:

```powershell
$env:VITE_API_URL="http://localhost:8080"
npm run dev
```

## Build

```powershell
npx tsc -b
npm run build
```

If `npm run build` fails with a Tailwind or Vite native binary error on Windows, that is usually a local dependency/runtime issue with `@tailwindcss/oxide`, not a TypeScript error in the app.

## Related Repos

- backend workspace: `../Main-Backend`
- Go API: `../Main-Backend/procure-ai`
- Algorand contract: `../Main-Backend/procure-contracts`
