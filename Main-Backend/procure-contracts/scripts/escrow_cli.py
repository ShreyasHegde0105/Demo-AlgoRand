from __future__ import annotations

import argparse
import base64
import json
import os
import sys
from pathlib import Path

from algokit_utils import AlgorandClient
from algokit_utils.applications.app_client import (
    AppClientMethodCallCreateParams,
    AppClientMethodCallParams,
    FundAppAccountParams,
)
from algokit_utils.models.amount import AlgoAmount
from algokit_utils.transactions.transaction_composer import NULL_SIGNER
from algosdk import encoding, logic, transaction
from algosdk.atomic_transaction_composer import TransactionWithSigner


ARTIFACT_PATH = Path(__file__).resolve().parents[1] / "artifacts" / "procure_escrow.arc56.json"


def load_local_env() -> None:
    env_path = Path(__file__).resolve().parents[1] / ".env"
    if not env_path.exists():
        return

    for raw_line in env_path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue

        name, value = line.split("=", 1)
        name = name.strip()
        value = value.strip().strip('"').strip("'")
        if name:
            os.environ.setdefault(name, value)


load_local_env()


def load_app_spec() -> dict:
    if not ARTIFACT_PATH.exists():
        raise FileNotFoundError(
            f"Missing app spec at {ARTIFACT_PATH}. Compile the contract first with AlgoKit."
        )
    return json.loads(ARTIFACT_PATH.read_text(encoding="utf-8"))


def get_algorand() -> AlgorandClient:
    return AlgorandClient.from_environment()


def get_network_name(algorand: AlgorandClient) -> str:
    return os.getenv("ALGORAND_NETWORK", "custom")


def json_safe(value: object) -> object:
    if isinstance(value, bytes):
        try:
            return value.decode("utf-8")
        except UnicodeDecodeError:
            return base64.b64encode(value).decode("ascii")
    if isinstance(value, dict):
        return {str(key): json_safe(item) for key, item in value.items()}
    if isinstance(value, list):
        return [json_safe(item) for item in value]
    return value


def encode_transactions(txns: list[transaction.Transaction], descriptions: list[str]) -> list[dict[str, object]]:
    transaction.assign_group_id(txns)
    return [
        {
            "index": idx,
            "description": descriptions[idx],
            "txnBase64": encoding.msgpack_encode(txn),
        }
        for idx, txn in enumerate(txns)
    ]


def create_escrow(args: argparse.Namespace) -> dict[str, object]:
    algorand = get_algorand()
    deployer = algorand.account.from_environment("DEPLOYER")
    app_spec = load_app_spec()
    app_name = f"{os.getenv('PROCURE_ESCROW_APP_NAME', 'procure-escrow')}-{args.order_id}"
    factory = algorand.client.get_app_factory(
        app_spec=app_spec,
        default_sender=deployer.address,
        app_name=app_name,
    )

    app_client, result = factory.deploy(
        create_params=AppClientMethodCallCreateParams(
            sender=deployer.address,
            method="create_application",
            args=[
                args.order_id,
                args.buyer,
                args.seller,
                args.agent,
                args.approver,
                args.amount,
                args.quote_valid_until,
            ],
        ),
        app_name=app_name,
    )

    return {
        "orderId": args.order_id,
        "appId": app_client.app_id,
        "appAddress": app_client.app_address,
        "buyerAddress": args.buyer,
        "sellerAddress": args.seller,
        "agentAddress": args.agent,
        "approverAddress": args.approver,
        "escrowAmountMicroAlgos": args.amount,
        "quoteValidUntil": args.quote_valid_until,
        "algorandNetwork": get_network_name(algorand),
        "operation": str(result.operation_performed),
    }


def prepare_fund(args: argparse.Namespace) -> dict[str, object]:
    algorand = get_algorand()
    app_spec = load_app_spec()
    app_client = algorand.client.get_app_client_by_id(
        app_spec=app_spec,
        app_id=args.app_id,
        app_name=f"{os.getenv('PROCURE_ESCROW_APP_NAME', 'procure-escrow')}-{args.order_id}",
        default_sender=args.buyer,
    )

    payment_txn = app_client.create_transaction.fund_app_account(
        FundAppAccountParams(
            sender=args.buyer,
            amount=AlgoAmount.from_micro_algo(args.amount),
        )
    )
    grouped = app_client.create_transaction.call(
        AppClientMethodCallParams(
            sender=args.buyer,
            method="fund",
            args=[TransactionWithSigner(txn=payment_txn, signer=NULL_SIGNER)],
        )
    )

    transactions = encode_transactions(
        grouped.transactions,
        [
            "Buyer payment into escrow app account",
            "ARC-4 fund() app call",
        ],
    )

    return {
        "orderId": args.order_id,
        "appId": args.app_id,
        "appAddress": logic.get_application_address(args.app_id),
        "action": "fund_escrow",
        "algorandNetwork": get_network_name(algorand),
        "transactions": transactions,
    }


def prepare_release(args: argparse.Namespace) -> dict[str, object]:
    algorand = get_algorand()
    app_spec = load_app_spec()
    app_client = algorand.client.get_app_client_by_id(
        app_spec=app_spec,
        app_id=args.app_id,
        app_name=f"{os.getenv('PROCURE_ESCROW_APP_NAME', 'procure-escrow')}-{args.order_id}",
        default_sender=args.approver,
    )

    confirm_delivery = app_client.create_transaction.call(
        AppClientMethodCallParams(
            sender=args.buyer,
            method="confirm_delivery",
            args=[],
        )
    )
    release_payment = app_client.create_transaction.call(
        AppClientMethodCallParams(
            sender=args.approver,
            method="release_payment",
            args=[],
        )
    )

    transactions = confirm_delivery.transactions + release_payment.transactions
    encoded_transactions = encode_transactions(
        transactions,
        [
            "ARC-4 confirm_delivery() app call",
            "ARC-4 release_payment() app call",
        ],
    )

    return {
        "orderId": args.order_id,
        "appId": args.app_id,
        "appAddress": logic.get_application_address(args.app_id),
        "action": "confirm_and_release",
        "algorandNetwork": get_network_name(algorand),
        "transactions": encoded_transactions,
    }


def prepare_select_supplier(args: argparse.Namespace) -> dict[str, object]:
    algorand = get_algorand()
    app_spec = load_app_spec()
    app_client = algorand.client.get_app_client_by_id(
        app_spec=app_spec,
        app_id=args.app_id,
        app_name=f"{os.getenv('PROCURE_ESCROW_APP_NAME', 'procure-escrow')}-{args.order_id}",
        default_sender=args.agent,
    )

    select_supplier = app_client.create_transaction.call(
        AppClientMethodCallParams(
            sender=args.agent,
            method="set_selected_supplier",
            args=[args.selected_supplier, args.quote_id],
        )
    )

    return {
        "orderId": args.order_id,
        "appId": args.app_id,
        "appAddress": logic.get_application_address(args.app_id),
        "action": "set_selected_supplier",
        "algorandNetwork": get_network_name(algorand),
        "transactions": encode_transactions(
            select_supplier.transactions,
            ["ARC-4 set_selected_supplier() app call"],
        ),
    }


def prepare_approve(args: argparse.Namespace) -> dict[str, object]:
    algorand = get_algorand()
    app_spec = load_app_spec()
    app_client = algorand.client.get_app_client_by_id(
        app_spec=app_spec,
        app_id=args.app_id,
        app_name=f"{os.getenv('PROCURE_ESCROW_APP_NAME', 'procure-escrow')}-{args.order_id}",
        default_sender=args.approver,
    )

    approve = app_client.create_transaction.call(
        AppClientMethodCallParams(
            sender=args.approver,
            method="approve",
            args=[],
        )
    )

    return {
        "orderId": args.order_id,
        "appId": args.app_id,
        "appAddress": logic.get_application_address(args.app_id),
        "action": "approve_escrow",
        "algorandNetwork": get_network_name(algorand),
        "transactions": encode_transactions(
            approve.transactions,
            ["ARC-4 approve() app call"],
        ),
    }


def get_state(args: argparse.Namespace) -> dict[str, object]:
    algorand = get_algorand()
    app_spec = load_app_spec()
    app_client = algorand.client.get_app_client_by_id(
        app_spec=app_spec,
        app_id=args.app_id,
        app_name=f"{os.getenv('PROCURE_ESCROW_APP_NAME', 'procure-escrow')}-{args.order_id}",
    )

    state = json_safe(app_client.state.global_state.get_all())
    return {
        "orderId": args.order_id,
        "appId": args.app_id,
        "appAddress": logic.get_application_address(args.app_id),
        "algorandNetwork": get_network_name(algorand),
        "state": state,
        "status": state.get("status"),
    }


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Procure escrow helper CLI")
    subparsers = parser.add_subparsers(dest="command", required=True)

    create_parser = subparsers.add_parser("create-escrow")
    create_parser.add_argument("--order-id", required=True)
    create_parser.add_argument("--buyer", required=True)
    create_parser.add_argument("--seller", required=True)
    create_parser.add_argument("--agent", required=True)
    create_parser.add_argument("--approver", required=True)
    create_parser.add_argument("--amount", type=int, required=True)
    create_parser.add_argument("--quote-valid-until", type=int, required=True)
    create_parser.set_defaults(handler=create_escrow)

    fund_parser = subparsers.add_parser("prepare-fund")
    fund_parser.add_argument("--order-id", required=True)
    fund_parser.add_argument("--app-id", type=int, required=True)
    fund_parser.add_argument("--buyer", required=True)
    fund_parser.add_argument("--amount", type=int, required=True)
    fund_parser.set_defaults(handler=prepare_fund)

    release_parser = subparsers.add_parser("prepare-release")
    release_parser.add_argument("--order-id", required=True)
    release_parser.add_argument("--app-id", type=int, required=True)
    release_parser.add_argument("--buyer", required=True)
    release_parser.add_argument("--approver", required=True)
    release_parser.set_defaults(handler=prepare_release)

    select_parser = subparsers.add_parser("prepare-select-supplier")
    select_parser.add_argument("--order-id", required=True)
    select_parser.add_argument("--app-id", type=int, required=True)
    select_parser.add_argument("--agent", required=True)
    select_parser.add_argument("--selected-supplier", required=True)
    select_parser.add_argument("--quote-id", required=True)
    select_parser.set_defaults(handler=prepare_select_supplier)

    approve_parser = subparsers.add_parser("prepare-approve")
    approve_parser.add_argument("--order-id", required=True)
    approve_parser.add_argument("--app-id", type=int, required=True)
    approve_parser.add_argument("--approver", required=True)
    approve_parser.set_defaults(handler=prepare_approve)

    state_parser = subparsers.add_parser("get-state")
    state_parser.add_argument("--order-id", required=True)
    state_parser.add_argument("--app-id", type=int, required=True)
    state_parser.set_defaults(handler=get_state)

    return parser


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()
    try:
        result = args.handler(args)
        print(json.dumps(result, separators=(",", ":")))
        return 0
    except Exception as exc:  # pragma: no cover - CLI error surface
        print(json.dumps({"error": str(exc)}), file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
