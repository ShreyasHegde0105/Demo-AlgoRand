"""Procure AI escrow contract built with AlgoPy ARC-4 patterns.

This contract is intentionally scoped as a single-order escrow application.
Each deployed app instance tracks one procurement order end-to-end.
"""

import algopy
from algopy import arc4


class ProcureEscrowContract(arc4.ARC4Contract):
    """Escrow contract for a single procurement order."""

    def __init__(self) -> None:
        self.order_id = algopy.GlobalState(algopy.Bytes, key="order_id")
        self.buyer = algopy.GlobalState(algopy.Account, key="buyer")
        self.seller = algopy.GlobalState(algopy.Account, key="seller")
        self.agent = algopy.GlobalState(algopy.Account, key="agent")
        self.approver = algopy.GlobalState(algopy.Account, key="approver")
        self.amount = algopy.GlobalState(algopy.UInt64, key="amount")
        self.status = algopy.GlobalState(algopy.UInt64, key="status")
        self.selected_supplier = algopy.GlobalState(
            algopy.Bytes, key="selected_supplier"
        )
        self.quote_id = algopy.GlobalState(algopy.Bytes, key="quote_id")
        self.quote_valid_until = algopy.GlobalState(
            algopy.UInt64, key="quote_valid_until"
        )
        self.approved = algopy.GlobalState(algopy.UInt64, key="approved")
        self.funded_at_round = algopy.GlobalState(algopy.UInt64, key="funded_at")
        self.funding_txn_id = algopy.GlobalState(algopy.Bytes, key="fund_txn")
        self.release_txn_id = algopy.GlobalState(algopy.Bytes, key="release_txn")
        self.refund_txn_id = algopy.GlobalState(algopy.Bytes, key="refund_txn")
        self.settlement_supplier = algopy.GlobalState(
            algopy.Bytes, key="settlement_supplier"
        )
        self.settlement_amount = algopy.GlobalState(
            algopy.UInt64, key="settlement_amount"
        )
        self.settlement_txn = algopy.GlobalState(algopy.Bytes, key="settlement_txn")

    @arc4.abimethod(create="require")
    def create_application(
        self,
        order_id: arc4.String,
        buyer: arc4.Address,
        seller: arc4.Address,
        agent: arc4.Address,
        approver: arc4.Address,
        amount: arc4.UInt64,
        quote_valid_until: arc4.UInt64,
    ) -> None:
        """Initialise escrow for a single order."""
        assert amount.native > algopy.UInt64(0), "amount must be greater than zero"
        assert buyer.native != seller.native, "buyer and seller must differ"
        assert quote_valid_until.native > algopy.Global.round, "quote must expire later"

        self.order_id.value = order_id.bytes
        self.buyer.value = buyer.native
        self.seller.value = seller.native
        self.agent.value = agent.native
        self.approver.value = approver.native
        self.amount.value = amount.native
        self.status.value = algopy.UInt64(0)
        self.selected_supplier.value = algopy.Bytes()
        self.quote_id.value = algopy.Bytes()
        self.quote_valid_until.value = quote_valid_until.native
        self.approved.value = algopy.UInt64(0)
        self.funded_at_round.value = algopy.UInt64(0)
        self.funding_txn_id.value = algopy.Bytes()
        self.release_txn_id.value = algopy.Bytes()
        self.refund_txn_id.value = algopy.Bytes()
        self.settlement_supplier.value = algopy.Bytes()
        self.settlement_amount.value = algopy.UInt64(0)
        self.settlement_txn.value = algopy.Bytes()

    @arc4.abimethod()
    def fund(self, payment: algopy.gtxn.PaymentTransaction) -> None:
        """Record buyer funding when a grouped payment sends ALGO to the app."""
        assert self.status.value == algopy.UInt64(0), "escrow is not awaiting funding"
        assert payment.sender == self.buyer.value, "payment sender must be the buyer"
        assert (
            payment.receiver == algopy.Global.current_application_address
        ), "payment receiver must be app account"
        assert payment.amount == self.amount.value, "payment amount mismatch"

        self.funding_txn_id.value = payment.txn_id
        self.funded_at_round.value = algopy.Global.round
        self.status.value = algopy.UInt64(1)

    @arc4.abimethod()
    def set_selected_supplier(
        self,
        selected_supplier: arc4.String,
        quote_id: arc4.String,
    ) -> None:
        """Persist the AI-selected supplier and quote reference."""
        assert algopy.Txn.sender == self.agent.value, "only agent can select supplier"
        assert self.status.value != algopy.UInt64(3), "payment already released"
        assert self.status.value != algopy.UInt64(4), "escrow already refunded"
        assert selected_supplier.bytes != algopy.Bytes(), "selected supplier is required"
        assert quote_id.bytes != algopy.Bytes(), "quote id is required"

        self.selected_supplier.value = selected_supplier.bytes
        self.quote_id.value = quote_id.bytes
        self.approved.value = algopy.UInt64(0)

    @arc4.abimethod()
    def approve(self) -> None:
        """Human approver authorises escrow release."""
        assert algopy.Txn.sender == self.approver.value, "only approver can approve"
        assert self.status.value != algopy.UInt64(3), "payment already released"
        assert self.status.value != algopy.UInt64(4), "escrow already refunded"
        assert self.selected_supplier.value != algopy.Bytes(), "supplier not selected"
        assert self.quote_id.value != algopy.Bytes(), "quote id not set"

        self.approved.value = algopy.UInt64(1)

    @arc4.abimethod()
    def confirm_delivery(self) -> None:
        """Buyer confirms the goods arrived and unlocks release."""
        assert self.status.value == algopy.UInt64(1), "escrow must be funded first"
        assert algopy.Txn.sender == self.buyer.value, "only buyer can confirm delivery"

        self.status.value = algopy.UInt64(2)

    @arc4.abimethod()
    def release_payment(self) -> None:
        """Release escrowed ALGO to the seller after delivery confirmation."""
        assert self.status.value == algopy.UInt64(2), "delivery has not been confirmed"
        assert algopy.Txn.sender == self.approver.value, "only approver can release payment"
        assert self.approved.value == algopy.UInt64(1), "payment not approved"
        assert self.selected_supplier.value != algopy.Bytes(), "supplier not selected"
        assert self.quote_id.value != algopy.Bytes(), "quote id not set"
        assert algopy.Global.round <= self.quote_valid_until.value, "quote has expired"

        payment_result = algopy.itxn.Payment(
            receiver=self.seller.value,
            amount=self.amount.value,
            fee=0,
        ).submit()

        self.release_txn_id.value = payment_result.txn_id
        self.settlement_supplier.value = self.selected_supplier.value
        self.settlement_amount.value = self.amount.value
        self.settlement_txn.value = payment_result.txn_id
        self.status.value = algopy.UInt64(3)

    @arc4.abimethod()
    def refund_buyer(self) -> None:
        """Refund buyer before delivery confirmation.

        For the demo this is creator-controlled so an admin workflow can resolve disputes.
        """
        assert self.status.value == algopy.UInt64(1), "refund only allowed for funded escrow"
        assert (
            algopy.Txn.sender == algopy.Global.creator_address
        ), "only creator can trigger refund"

        payment_result = algopy.itxn.Payment(
            receiver=self.buyer.value,
            amount=self.amount.value,
            fee=0,
        ).submit()

        self.refund_txn_id.value = payment_result.txn_id
        self.status.value = algopy.UInt64(4)

    @arc4.abimethod(readonly=True)
    def get_status(self) -> arc4.UInt64:
        return arc4.UInt64(self.status.value)

    @arc4.abimethod(readonly=True)
    def get_amount(self) -> arc4.UInt64:
        return arc4.UInt64(self.amount.value)
