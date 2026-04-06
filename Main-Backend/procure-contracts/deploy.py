"""Deployment entrypoint for Procure AI escrow contracts.

This script is intentionally lightweight. It assumes contract compilation will
produce an ARC-56 app spec under artifacts/, then deploys that spec using
AlgoKit Utils.
"""

from __future__ import annotations

import json
import os
from pathlib import Path

from algokit_utils import AlgorandClient
from algokit_utils.applications.app_client import AppClientMethodCallCreateParams


ARTIFACT_PATH = Path("artifacts") / "procure_escrow.arc56.json"


def main() -> None:
    if not ARTIFACT_PATH.exists():
        raise FileNotFoundError(
            f"Expected compiled app spec at {ARTIFACT_PATH}. "
            "Compile the contract first, then rerun deployment."
        )

    algorand = AlgorandClient.from_environment()
    deployer = algorand.account.from_environment("DEPLOYER")

    app_spec = json.loads(ARTIFACT_PATH.read_text(encoding="utf-8"))
    factory = algorand.client.get_app_factory(
        app_spec=app_spec,
        default_sender=deployer.address,
        app_name=os.getenv("PROCURE_ESCROW_APP_NAME", "procure-escrow"),
    )

    order_id = os.getenv("PROCURE_ESCROW_ORDER_ID", "ORD-DEMO")
    buyer = os.getenv("PROCURE_ESCROW_BUYER")
    seller = os.getenv("PROCURE_ESCROW_SELLER")
    agent = os.getenv("PROCURE_ESCROW_AGENT", buyer or "")
    approver = os.getenv("PROCURE_ESCROW_APPROVER", buyer or "")
    amount = int(os.getenv("PROCURE_ESCROW_AMOUNT_MICROALGOS", "10000"))
    quote_valid_until = int(os.getenv("PROCURE_ESCROW_QUOTE_VALID_UNTIL", "999999999"))
    if not buyer or not seller or not agent or not approver:
        raise ValueError(
            "PROCURE_ESCROW_BUYER, PROCURE_ESCROW_SELLER, PROCURE_ESCROW_AGENT, and "
            "PROCURE_ESCROW_APPROVER must be set to create a per-order escrow app."
        )

    app_client, result = factory.deploy(
        create_params=AppClientMethodCallCreateParams(
            sender=deployer.address,
            method="create_application",
            args=[order_id, buyer, seller, agent, approver, amount, quote_valid_until],
        ),
        app_name=f"{os.getenv('PROCURE_ESCROW_APP_NAME', 'procure-escrow')}-{order_id}",
    )

    print("Deployment complete")
    print(f"app_id={app_client.app_id}")
    print(f"app_address={app_client.app_address}")
    print(f"operation={result.operation_performed}")


if __name__ == "__main__":
    main()
