"""Procure AI escrow contract built with AlgoPy ARC-4 patterns."""

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
        self.quote_valid_until = algopy.GlobalState(algopy.UInt64, key="quote_valid_until")

        self.status = algopy.GlobalState(algopy.UInt64, key="status")
        self.selected_supplier = algopy.GlobalState(algopy.Bytes, key="selected_supplier")
        self.supplier_address = algopy.GlobalState(algopy.Account, key="supplier_addr")
        self.selected_at_round = algopy.GlobalState(algopy.UInt64, key="selected_at")
        self.quote_id = algopy.GlobalState(algopy.Bytes, key="quote_id")
        self.quote_amount = algopy.GlobalState(algopy.UInt64, key="quote_amount")

        self.approved = algopy.GlobalState(algopy.UInt64, key="approved")
        self.approved_by = algopy.GlobalState(algopy.Account, key="approved_by")
        self.approved_at_round = algopy.GlobalState(algopy.UInt64, key="approved_at")

        self.funded_at_round = algopy.GlobalState(algopy.UInt64, key="funded_at")
        self.funding_txn_id = algopy.GlobalState(algopy.Bytes, key="fund_txn")
        self.delivery_confirmed_at = algopy.GlobalState(algopy.UInt64, key="delivered_at")
        self.release_txn_id = algopy.GlobalState(algopy.Bytes, key="release_txn")
        self.refund_txn_id = algopy.GlobalState(algopy.Bytes, key="refund_txn")
        self.refund_triggered_by = algopy.GlobalState(algopy.Account, key="refund_by")
        self.settlement_supplier = algopy.GlobalState(algopy.Bytes, key="settlement_supplier")
        self.settlement_amount = algopy.GlobalState(algopy.UInt64, key="settlement_amount")
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
        self.quote_valid_until.value = quote_valid_until.native

        self.status.value = algopy.UInt64(0)
        self.selected_supplier.value = algopy.Bytes()
        self.supplier_address.value = algopy.Global.zero_address
        self.selected_at_round.value = algopy.UInt64(0)
        self.quote_id.value = algopy.Bytes()
        self.quote_amount.value = algopy.UInt64(0)
        self.approved.value = algopy.UInt64(0)
        self.approved_by.value = algopy.Global.zero_address
        self.approved_at_round.value = algopy.UInt64(0)
        self.funded_at_round.value = algopy.UInt64(0)
        self.funding_txn_id.value = algopy.Bytes()
        self.delivery_confirmed_at.value = algopy.UInt64(0)
        self.release_txn_id.value = algopy.Bytes()
        self.refund_txn_id.value = algopy.Bytes()
        self.refund_triggered_by.value = algopy.Global.zero_address
        self.settlement_supplier.value = algopy.Bytes()
        self.settlement_amount.value = algopy.UInt64(0)
        self.settlement_txn.value = algopy.Bytes()

    @arc4.abimethod()
    def set_selected_supplier(
        self,
        selected_supplier: arc4.String,
        supplier_address: arc4.Address,
        quote_id: arc4.String,
        quote_amount: arc4.UInt64,
    ) -> None:
        """Persist the AI-selected supplier and bind it to the payout address."""
        assert algopy.Txn.sender == self.agent.value, "only agent can select supplier"
        assert self.status.value == algopy.UInt64(0), "supplier selection only allowed from created"
        assert algopy.Global.round <= self.quote_valid_until.value, "quote has expired"
        assert selected_supplier.bytes != algopy.Bytes(), "selected supplier is required"
        assert quote_id.bytes != algopy.Bytes(), "quote id is required"
        assert supplier_address.native != algopy.Global.zero_address, "supplier address required"
        assert supplier_address.native == self.seller.value, "supplier address must match seller"
        assert quote_amount.native == self.amount.value, "quoted amount must match escrowed amount"

        self.selected_supplier.value = selected_supplier.bytes
        self.supplier_address.value = supplier_address.native
        self.selected_at_round.value = algopy.Global.round
        self.quote_id.value = quote_id.bytes
        self.quote_amount.value = quote_amount.native
        self.approved.value = algopy.UInt64(0)
        self.approved_by.value = algopy.Global.zero_address
        self.approved_at_round.value = algopy.UInt64(0)
        self.status.value = algopy.UInt64(1)

    @arc4.abimethod()
    def approve(self) -> None:
        """Human approver authorises funding and later release."""
        assert algopy.Txn.sender == self.approver.value, "only approver can approve"
        assert self.status.value == algopy.UInt64(1), "approval only allowed after supplier selection"
        assert algopy.Global.round <= self.quote_valid_until.value, "quote has expired"
        assert self.selected_supplier.value != algopy.Bytes(), "supplier not selected"
        assert self.quote_id.value != algopy.Bytes(), "quote id not set"

        self.approved.value = algopy.UInt64(1)
        self.approved_by.value = algopy.Txn.sender
        self.approved_at_round.value = algopy.Global.round
        self.status.value = algopy.UInt64(2)

    @arc4.abimethod()
    def fund(self, payment: algopy.gtxn.PaymentTransaction) -> None:
        """Record buyer funding when a grouped payment sends ALGO to the app."""
        assert self.status.value == algopy.UInt64(2), "escrow is not ready for funding"
        assert payment.sender == self.buyer.value, "payment sender must be the buyer"
        assert payment.receiver == algopy.Global.current_application_address, "payment receiver must be app account"
        assert payment.amount == self.amount.value, "payment amount mismatch"
        assert algopy.Global.round <= self.quote_valid_until.value, "quote has expired before funding"

        self.funding_txn_id.value = payment.txn_id
        self.funded_at_round.value = algopy.Global.round
        self.status.value = algopy.UInt64(3)

    @arc4.abimethod()
    def confirm_delivery(self) -> None:
        """Buyer confirms the goods arrived and unlocks release."""
        assert self.status.value == algopy.UInt64(3), "escrow must be funded before delivery confirmation"
        assert algopy.Txn.sender == self.buyer.value, "only buyer can confirm delivery"

        self.delivery_confirmed_at.value = algopy.Global.round
        self.status.value = algopy.UInt64(4)

    @arc4.abimethod()
    def release_payment(self) -> None:
        """Release escrowed ALGO to the seller after delivery confirmation."""
        assert self.status.value == algopy.UInt64(4), "delivery has not been confirmed"
        assert algopy.Txn.sender == self.approver.value, "only approver can release payment"
        assert self.approved.value == algopy.UInt64(1), "payment not approved"
        assert self.selected_supplier.value != algopy.Bytes(), "supplier not selected"
        assert self.quote_id.value != algopy.Bytes(), "quote id not set"
        assert self.supplier_address.value == self.seller.value, "supplier address does not match seller"
        assert self.quote_amount.value == self.amount.value, "quote amount does not match escrowed amount"

        self.status.value = algopy.UInt64(5)
        payment_result = algopy.itxn.Payment(
            receiver=self.seller.value,
            amount=self.amount.value,
            fee=0,
        ).submit()

        self.release_txn_id.value = payment_result.txn_id
        self.settlement_supplier.value = self.selected_supplier.value
        self.settlement_amount.value = self.amount.value
        self.settlement_txn.value = payment_result.txn_id

    @arc4.abimethod()
    def refund_buyer_on_expiry(self) -> None:
        """Allow the buyer to refund after expiry while funds are still locked."""
        assert self.status.value == algopy.UInt64(3), "refund only allowed for funded escrow"
        assert algopy.Txn.sender == self.buyer.value, "only buyer can trigger expiry refund"
        assert algopy.Global.round > self.quote_valid_until.value, "quote has not expired yet"

        self.status.value = algopy.UInt64(6)
        payment_result = algopy.itxn.Payment(
            receiver=self.buyer.value,
            amount=self.amount.value,
            fee=0,
        ).submit()

        self.refund_txn_id.value = payment_result.txn_id
        self.refund_triggered_by.value = algopy.Txn.sender

    @arc4.abimethod()
    def refund_buyer_dispute(self) -> None:
        """Emergency refund path for the app creator."""
        assert self.status.value == algopy.UInt64(3), "refund only allowed for funded escrow"
        assert algopy.Txn.sender == algopy.Global.creator_address, "only creator can trigger dispute refund"

        self.status.value = algopy.UInt64(6)
        payment_result = algopy.itxn.Payment(
            receiver=self.buyer.value,
            amount=self.amount.value,
            fee=0,
        ).submit()

        self.refund_txn_id.value = payment_result.txn_id
        self.refund_triggered_by.value = algopy.Txn.sender

    @arc4.abimethod(readonly=True)
    def get_status(self) -> arc4.UInt64:
        return arc4.UInt64(self.status.value)

    @arc4.abimethod(readonly=True)
    def get_amount(self) -> arc4.UInt64:
        return arc4.UInt64(self.amount.value)

    @arc4.abimethod(readonly=True)
    def get_supplier_address(self) -> arc4.Address:
        return arc4.Address(self.supplier_address.value)

    @arc4.abimethod(readonly=True)
    def get_quote_amount(self) -> arc4.UInt64:
        return arc4.UInt64(self.quote_amount.value)
