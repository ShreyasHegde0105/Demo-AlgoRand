import { Link2 } from 'lucide-react'
import { useEffect, useState } from 'react'
import { apiPost } from '../api/client'
import type { CreateEscrowResponse, Order } from '../api/types'
import { Button, Field, Panel, TextInput } from './Ui'

export function EscrowPanel({
  order,
  onEscrowCreated,
  notify,
}: {
  order: Order | null
  onEscrowCreated: (o: Order) => void
  notify: (message: string, ok?: boolean) => void
}) {
  const [busy, setBusy] = useState(false)
  const [buyer, setBuyer] = useState('')
  const [seller, setSeller] = useState('')
  const [agent, setAgent] = useState('')
  const [approver, setApprover] = useState('')
  const [microAlgos, setMicroAlgos] = useState('10000')
  const [quoteUntil, setQuoteUntil] = useState('999999999')

  useEffect(() => {
    setBuyer(order?.buyerAddress ?? '')
    setSeller(order?.sellerAddress ?? '')
    setAgent(order?.agentAddress ?? '')
    setApprover(order?.approverAddress ?? '')
    setMicroAlgos(order?.escrowAmountMicroAlgos ? String(order.escrowAmountMicroAlgos) : '10000')
    setQuoteUntil(order?.quoteValidUntil ? String(order.quoteValidUntil) : '999999999')
  }, [order])

  const canSubmit =
    order?.status === 'approved' &&
    !order.algorandAppId &&
    buyer &&
    seller &&
    agent &&
    approver &&
    Number.parseInt(microAlgos, 10) > 0 &&
    Number.parseInt(quoteUntil, 10) > 0

  async function createEscrow() {
    if (!order) return
    setBusy(true)
    try {
      const res = await apiPost<CreateEscrowResponse>('/blockchain/create-escrow', {
        orderId: order.id,
        buyerAddress: buyer,
        sellerAddress: seller,
        agentAddress: agent,
        approverAddress: approver,
        escrowAmountMicroAlgos: Number.parseInt(microAlgos, 10),
        quoteValidUntil: Number.parseInt(quoteUntil, 10),
      })
      notify(
        `Escrow app ${res.appId} created on ${res.algorandNetwork}. Continue to On-chain steps.`,
        true,
      )
      const merged: Order = {
        ...order,
        algorandAppId: res.appId,
        algorandAppAddress: res.appAddress,
        buyerAddress: buyer,
        sellerAddress: seller,
        agentAddress: agent,
        approverAddress: approver,
        escrowAmountMicroAlgos: Number.parseInt(microAlgos, 10),
        quoteValidUntil: Number.parseInt(quoteUntil, 10),
        algorandNetwork: res.algorandNetwork,
      }
      onEscrowCreated(merged)
    } catch (e) {
      notify(e instanceof Error ? e.message : 'Create escrow failed')
    } finally {
      setBusy(false)
    }
  }

  return (
    <Panel
      title="Algorand escrow"
      subtitle="One smart-contract app per order. Demo amount is often 10 000 microAlgos (0.01 ALGO)."
    >
      {!order ? (
        <p className="text-sm text-slate-500">Create and approve an order first.</p>
      ) : order.algorandAppId ? (
        <div className="space-y-2 text-sm">
          <p className="flex items-center gap-2 text-teal-400">
            <Link2 className="h-4 w-4" />
            Escrow already linked to this order.
          </p>
          <p>
            App ID{' '}
            <code className="text-slate-200">{order.algorandAppId}</code> · Address{' '}
            <code className="break-all text-slate-300">{order.algorandAppAddress}</code>
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <Field label="Buyer address" hint="Funds the escrow">
              <TextInput value={buyer} onChange={(e) => setBuyer(e.target.value.trim())} />
            </Field>
            <Field label="Seller address" hint="Receives release">
              <TextInput value={seller} onChange={(e) => setSeller(e.target.value.trim())} />
            </Field>
            <Field label="Agent address" hint="Selects supplier on-chain">
              <TextInput value={agent} onChange={(e) => setAgent(e.target.value.trim())} />
            </Field>
            <Field label="Approver address" hint="Approves escrow & releases">
              <TextInput value={approver} onChange={(e) => setApprover(e.target.value.trim())} />
            </Field>
            <Field label="Amount (microAlgos)">
              <TextInput value={microAlgos} onChange={(e) => setMicroAlgos(e.target.value)} />
            </Field>
            <Field label="Quote valid until (round)">
              <TextInput value={quoteUntil} onChange={(e) => setQuoteUntil(e.target.value)} />
            </Field>
          </div>
          <Button
            disabled={!canSubmit || busy}
            onClick={() => void createEscrow()}
          >
            Deploy escrow for {order.id}
          </Button>
          {order.status !== 'approved' ? (
            <p className="text-xs text-amber-400/90">Order must be approved before escrow.</p>
          ) : null}
        </div>
      )}
    </Panel>
  )
}
