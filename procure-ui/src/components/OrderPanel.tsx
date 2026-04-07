import { CheckCircle2, FileCheck } from 'lucide-react'
import { useState } from 'react'
import { apiPost } from '../api/client'
import type { Order, ProcurementRecommendationResponse } from '../api/types'
import { Button, Field, Panel, TextInput } from './Ui'

export function OrderPanel({
  recommendation,
  selectedVendor,
  orderQuantity,
  order,
  onOrderCreated,
  onOrderApproved,
  notify,
}: {
  recommendation: ProcurementRecommendationResponse | null
  selectedVendor: string
  orderQuantity: number
  order: Order | null
  onOrderCreated: (o: Order) => void
  onOrderApproved: (o: Order) => void
  notify: (message: string, ok?: boolean) => void
}) {
  const [busy, setBusy] = useState(false)

  const canCreate = Boolean(
    recommendation?.recommendationId && selectedVendor && orderQuantity > 0,
  )

  async function createOrder() {
    if (!recommendation?.recommendationId) return
    setBusy(true)
    try {
      const o = await apiPost<Order>('/create-order', {
        recommendationId: recommendation.recommendationId,
        vendor: selectedVendor,
        quantity: orderQuantity,
      })
      onOrderCreated(o)
      notify(`Order ${o.id} created - pending approval.`, true)
    } catch (e) {
      notify(e instanceof Error ? e.message : 'Create order failed')
    } finally {
      setBusy(false)
    }
  }

  async function approveOrder() {
    if (!order?.id) return
    setBusy(true)
    try {
      const res = await apiPost<{ order?: Order; message: string }>('/approve-order', {
        orderId: order.id,
      })
      if (res.order) onOrderApproved(res.order)
      notify(res.message || 'Order approved', true)
    } catch (e) {
      notify(e instanceof Error ? e.message : 'Approve failed')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="space-y-6">
      <Panel
        title="Raise purchase order"
        subtitle="Quantity must match the recommendation session (backend validates this)."
      >
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Recommendation ID">
            <TextInput
              readOnly
              value={recommendation?.recommendationId ?? ''}
              placeholder="Run Discover first"
            />
          </Field>
          <Field label="Supplier">
            <TextInput readOnly value={selectedVendor} placeholder="Select a supplier card" />
          </Field>
          <Field label="Quantity (fixed)">
            <TextInput readOnly value={String(orderQuantity || '')} />
          </Field>
          <div className="flex items-end">
            <Button disabled={!canCreate || busy || Boolean(order)} onClick={() => void createOrder()}>
              Create order
            </Button>
          </div>
        </div>
      </Panel>

      <Panel
        title="Approver sign-off"
        subtitle="Moves the order to approved so you can open on-chain escrow."
      >
        <div className="flex flex-wrap items-center gap-4">
          {order ? (
            <div className="flex items-center gap-2 text-sm">
              <span className="text-slate-500">Order</span>
              <code className="rounded-lg bg-slate-900 px-2 py-1 text-teal-300">{order.id}</code>
              <span className="rounded-full bg-slate-800 px-2 py-0.5 text-xs text-slate-400">
                {order.status}
              </span>
            </div>
          ) : (
            <p className="text-sm text-slate-500">No order yet.</p>
          )}
          <Button
            variant="secondary"
            disabled={!order || order.status !== 'pending_approval' || busy}
            onClick={() => void approveOrder()}
          >
            <FileCheck className="h-4 w-4" />
            Approve order
          </Button>
        </div>
        {order?.status === 'approved' ? (
          <p className="mt-4 flex items-center gap-2 text-sm text-teal-400">
            <CheckCircle2 className="h-4 w-4 shrink-0" />
            Approved - you can configure escrow on the next tab.
          </p>
        ) : null}
      </Panel>
    </div>
  )
}
