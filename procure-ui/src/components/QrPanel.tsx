import { QrCode, ShieldCheck } from 'lucide-react'
import { useState } from 'react'
import { apiPost } from '../api/client'
import type { Order } from '../api/types'
import { Button, Field, Panel, TextInput } from './Ui'

export function QrPanel({
  order,
  notify,
}: {
  order: Order | null
  notify: (message: string, ok?: boolean) => void
}) {
  const [busy, setBusy] = useState(false)
  const [qrImage, setQrImage] = useState<string | null>(null)
  const [qrPayload, setQrPayload] = useState('')
  const [verifyCode, setVerifyCode] = useState('')
  const [verifyOk, setVerifyOk] = useState<boolean | null>(null)

  async function generate() {
    if (!order?.id) return
    setBusy(true)
    try {
      const res = await apiPost<{
        orderId: string
        qrImageBase64: string
        qrCode: string
        message: string
      }>('/generate-qr', { orderId: order.id })
      setQrImage(res.qrImageBase64)
      setQrPayload(res.qrCode)
      notify(res.message || 'QR generated', true)
    } catch (e) {
      notify(e instanceof Error ? e.message : 'Generate failed')
    } finally {
      setBusy(false)
    }
  }

  async function verify() {
    if (!order?.id) return
    setBusy(true)
    setVerifyOk(null)
    try {
      const res = await apiPost<{ valid: boolean; message: string }>('/verify-qr', {
        orderId: order.id,
        qrCode: verifyCode,
      })
      setVerifyOk(res.valid)
      notify(res.message, res.valid)
    } catch (e) {
      notify(e instanceof Error ? e.message : 'Verify failed')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="space-y-6">
      <Panel
        title="Delivery QR"
        subtitle="Off-chain handoff: generate a code for couriers or warehouse scanning."
      >
        {!order ? (
          <p className="text-sm text-slate-500">Need an order context.</p>
        ) : (
          <div className="flex flex-wrap items-start gap-6">
            <div>
              <Button disabled={busy} onClick={() => void generate()}>
                <QrCode className="h-4 w-4" />
                Generate for {order.id}
              </Button>
              {qrImage ? (
                <div className="mt-4 rounded-2xl border border-slate-700 bg-white p-3 shadow-inner">
                  <img
                    src={`data:image/png;base64,${qrImage}`}
                    alt="Order QR"
                    className="h-48 w-48 object-contain"
                  />
                </div>
              ) : null}
            </div>
            {qrPayload ? (
              <div className="min-w-0 flex-1">
                <Field label="QR payload">
                  <TextInput readOnly value={qrPayload} className="font-mono text-xs" />
                </Field>
              </div>
            ) : null}
          </div>
        )}
      </Panel>

      <Panel title="Verify QR" subtitle="Paste a scanned token to validate against the order.">
        <div className="grid gap-4 sm:grid-cols-2 sm:items-end">
          <Field label="Scanned QR content">
            <TextInput value={verifyCode} onChange={(e) => setVerifyCode(e.target.value)} />
          </Field>
          <Button disabled={busy || !order?.id || !verifyCode} onClick={() => void verify()}>
            <ShieldCheck className="h-4 w-4" />
            Verify
          </Button>
        </div>
        {verifyOk !== null ? (
          <p
            className={`mt-3 text-sm font-medium ${verifyOk ? 'text-teal-400' : 'text-rose-400'}`}
          >
            {verifyOk ? 'Token matches this order.' : 'Verification failed.'}
          </p>
        ) : null}
      </Panel>
    </div>
  )
}
