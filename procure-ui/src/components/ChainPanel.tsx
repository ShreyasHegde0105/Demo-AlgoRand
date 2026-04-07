import { useEffect, useState } from 'react'
import { apiPost } from '../api/client'
import type { Order, PrepareEscrowActionResponse } from '../api/types'
import { CopyBlock } from './CopyBlock'
import { Button, Field, Panel, TextInput } from './Ui'

function mergeOrder(prev: Order | null, next?: Order): Order | null {
  if (!next) return prev
  if (!prev) return next
  return { ...prev, ...next }
}

export function ChainPanel({
  order,
  selectedSupplierName,
  onOrderUpdate,
  notify,
}: {
  order: Order | null
  selectedSupplierName: string
  onOrderUpdate: (o: Order) => void
  notify: (message: string, ok?: boolean) => void
}) {
  const [busy, setBusy] = useState(false)
  const [quoteId, setQuoteId] = useState('QUOTE-001')
  const [fundTxId, setFundTxId] = useState('')
  const [releaseTxId, setReleaseTxId] = useState('')
  const [lastPrepare, setLastPrepare] = useState<PrepareEscrowActionResponse | null>(null)
  const [supplierOnChain, setSupplierOnChain] = useState(selectedSupplierName)

  useEffect(() => {
    setSupplierOnChain(selectedSupplierName)
  }, [selectedSupplierName])

  const oid = order?.id
  const canPrepareSelection =
    Boolean(oid) &&
    order?.status === 'approved' &&
    supplierOnChain.trim().length > 0 &&
    quoteId.trim().length > 0
  const canPrepareApproval =
    Boolean(oid) && order?.status === 'approved' && Boolean(order?.selectedSupplier)
  const canPrepareFunding =
    Boolean(oid) && order?.status === 'approved' && Boolean(order?.escrowApproved)
  const canConfirmDelivery = Boolean(oid) && order?.status === 'funds_locked'
  const canPrepareRelease =
    Boolean(oid) && (order?.status === 'funds_locked' || order?.status === 'delivered')

  async function wrap(
    fn: () => Promise<{ order?: Order } & Record<string, unknown>>,
    okMsg: string,
  ) {
    setBusy(true)
    try {
      const res = await fn()
      const merged = mergeOrder(order, res.order)
      if (merged) onOrderUpdate(merged)
      notify(okMsg, true)
    } catch (e) {
      notify(e instanceof Error ? e.message : 'Request failed')
    } finally {
      setBusy(false)
    }
  }

  if (!order?.algorandAppId) {
    return (
      <Panel
        title="On-chain steps"
        subtitle="Create an escrow for this order first. Then walk through prepare → sign → confirm."
      >
        <p className="text-sm text-slate-500">No escrow app linked yet.</p>
      </Panel>
    )
  }

  return (
    <div className="space-y-6">
      <Panel
        title="Escrow state"
        subtitle="The on-chain contract now follows supplier selection -> approval -> funding -> delivery -> release."
      >
        <div className="flex flex-wrap items-center gap-3 text-sm text-slate-400">
          <span className="rounded-full bg-slate-800 px-3 py-1 capitalize">{order.status}</span>
          {order.selectedSupplier ? (
            <span>
              Supplier <code className="text-teal-300">{order.selectedSupplier}</code>
            </span>
          ) : (
            <span>No on-chain supplier confirmed yet.</span>
          )}
          {order.escrowApproved ? <span>Approver confirmed</span> : <span>Awaiting approval</span>}
        </div>
      </Panel>

      <Panel
        title="1 · Agent: select supplier"
        subtitle="Select the supplier on-chain first. The stricter contract flow requires this before approval and funding."
      >
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Selected supplier (string)">
            <TextInput
              value={supplierOnChain}
              onChange={(e) => setSupplierOnChain(e.target.value)}
            />
          </Field>
          <Field label="Quote ID">
            <TextInput value={quoteId} onChange={(e) => setQuoteId(e.target.value)} />
          </Field>
        </div>
        <div className="mt-4 flex flex-wrap gap-2">
          <Button
            disabled={busy || !canPrepareSelection}
            onClick={() => {
              void (async () => {
                setBusy(true)
                try {
                  const res = await apiPost<PrepareEscrowActionResponse>(
                    '/blockchain/prepare-select-supplier',
                    {
                      orderId: oid,
                      selectedSupplier: supplierOnChain,
                      quoteId,
                    },
                  )
                  setLastPrepare(res)
                  notify('Prepared supplier-selection transactions.', true)
                } catch (e) {
                  notify(e instanceof Error ? e.message : 'Prepare failed')
                } finally {
                  setBusy(false)
                }
              })()
            }}
          >
            Prepare
          </Button>
          <Button
            variant="secondary"
            disabled={busy || !oid}
            onClick={() =>
              void wrap(
                () => apiPost<{ order?: Order }>('/blockchain/confirm-select-supplier', { orderId: oid! }),
                'Supplier selection confirmed on-chain.',
              )
            }
          >
            Confirm
          </Button>
        </div>
        {!order?.selectedSupplier ? (
          <p className="mt-4 text-xs text-slate-500">
            Confirm this step after the agent wallet submits the selection transaction.
          </p>
        ) : null}
      </Panel>

      <Panel title="2 · Approver: approve escrow" subtitle="Approval now happens after supplier selection.">
        <div className="flex flex-wrap gap-2">
          <Button
            disabled={busy || !canPrepareApproval}
            onClick={() => {
              void (async () => {
                setBusy(true)
                try {
                  const res = await apiPost<PrepareEscrowActionResponse>(
                    '/blockchain/prepare-approve',
                    { orderId: oid! },
                  )
                  setLastPrepare(res)
                  notify('Prepared approval transactions.', true)
                } catch (e) {
                  notify(e instanceof Error ? e.message : 'Prepare failed')
                } finally {
                  setBusy(false)
                }
              })()
            }}
          >
            Prepare
          </Button>
          <Button
            variant="secondary"
            disabled={busy || !oid}
            onClick={() =>
              void wrap(
                () => apiPost<{ order?: Order }>('/blockchain/confirm-approve', { orderId: oid! }),
                'Escrow approval confirmed.',
              )
            }
          >
            Confirm
          </Button>
        </div>
        {!order?.selectedSupplier ? (
          <p className="mt-4 text-xs text-amber-400/90">Select and confirm a supplier first.</p>
        ) : null}
      </Panel>

      <Panel title="3 · Buyer: fund escrow" subtitle="Funding is only valid after supplier selection and approval.">
        <div className="flex flex-wrap gap-2">
          <Button
            disabled={busy || !canPrepareFunding}
            onClick={() => {
              void (async () => {
                setBusy(true)
                try {
                  const res = await apiPost<PrepareEscrowActionResponse>('/blockchain/prepare-fund', {
                    orderId: oid!,
                  })
                  setLastPrepare(res)
                  notify('Prepared funding transactions.', true)
                } catch (e) {
                  notify(e instanceof Error ? e.message : 'Prepare failed')
                } finally {
                  setBusy(false)
                }
              })()
            }}
          >
            Prepare
          </Button>
        </div>
        <div className="mt-4 grid gap-3 sm:grid-cols-2 sm:items-end">
          <Field label="Submitted transaction / group ID" hint="After signing & broadcast">
            <TextInput value={fundTxId} onChange={(e) => setFundTxId(e.target.value.trim())} />
          </Field>
          <Button
            variant="secondary"
            disabled={busy || !oid || !fundTxId}
            onClick={() =>
              void wrap(
                () =>
                  apiPost<{ order?: Order }>('/blockchain/confirm-fund', {
                    orderId: oid!,
                    txID: fundTxId,
                  }),
                'Funding confirmed.',
              )
            }
          >
            Confirm fund
          </Button>
        </div>
      </Panel>

      <Panel title="4 · Buyer: confirm delivery" subtitle="Use this only after escrow funding has been confirmed.">
        <Button
          disabled={busy || !canConfirmDelivery}
          onClick={() =>
            void wrap(
              () => apiPost<{ order?: Order }>('/confirm-delivery', { orderId: oid! }),
              'Delivery marked on the order.',
            )
          }
        >
          Confirm delivery
        </Button>
      </Panel>

      <Panel title="5 · Approver: release payment" subtitle="Prepare release, sign, then confirm with tx ID after delivery is marked.">
        <div className="flex flex-wrap gap-2">
          <Button
            disabled={busy || !canPrepareRelease}
            onClick={() => {
              void (async () => {
                setBusy(true)
                try {
                  const res = await apiPost<PrepareEscrowActionResponse>(
                    '/blockchain/prepare-release',
                    { orderId: oid! },
                  )
                  setLastPrepare(res)
                  notify('Prepared release transactions.', true)
                } catch (e) {
                  notify(e instanceof Error ? e.message : 'Prepare failed')
                } finally {
                  setBusy(false)
                }
              })()
            }}
          >
            Prepare
          </Button>
        </div>
        <div className="mt-4 grid gap-3 sm:grid-cols-2 sm:items-end">
          <Field label="Release transaction ID">
            <TextInput value={releaseTxId} onChange={(e) => setReleaseTxId(e.target.value.trim())} />
          </Field>
          <Button
            variant="secondary"
            disabled={busy || !oid || !releaseTxId}
            onClick={() =>
              void wrap(
                () =>
                  apiPost<{ order?: Order }>('/blockchain/confirm-release', {
                    orderId: oid!,
                    txID: releaseTxId,
                  }),
                'Release confirmed - settlement saved.',
              )
            }
          >
            Confirm release
          </Button>
        </div>
      </Panel>

      {lastPrepare?.transactions?.length ? (
        <Panel
          title="Latest prepared transactions"
          subtitle={`Action: ${lastPrepare.action} · Network: ${lastPrepare.algorandNetwork}`}
        >
          <p className="mb-4 text-xs text-slate-500">
            Copy each payload into your Algorand wallet or CLI, sign with the correct account, broadcast,
            then run the matching Confirm step.
          </p>
          <div className="space-y-3">
            {lastPrepare.transactions.map((t) => (
              <CopyBlock
                key={t.index}
                label={`#${t.index} - ${t.description}`}
                value={t.txnBase64}
              />
            ))}
          </div>
        </Panel>
      ) : null}
    </div>
  )
}
