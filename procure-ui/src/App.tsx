import {
  ArrowRight,
  Box,
  ChevronRight,
  Cpu,
  LayoutDashboard,
  Link2,
  QrCode,
  ShoppingCart,
} from 'lucide-react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { getApiBaseUrl } from './api/client'
import type { Order, ProcurementRecommendationResponse } from './api/types'
import { ChainPanel } from './components/ChainPanel'
import { DiscoverPanel } from './components/DiscoverPanel'
import { EscrowPanel } from './components/EscrowPanel'
import { OrderPanel } from './components/OrderPanel'
import { QrPanel } from './components/QrPanel'

type TabId = 'discover' | 'order' | 'escrow' | 'chain' | 'qr'

type SessionV1 = {
  recommendation?: ProcurementRecommendationResponse | null
  order?: Order | null
  selectedVendor?: string
  orderQuantity?: number
}

const STORAGE_KEY = 'procure-ui-v1'

function loadSession(): SessionV1 {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return {}
    return JSON.parse(raw) as SessionV1
  } catch {
    return {}
  }
}

function saveSession(s: SessionV1) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(s))
  } catch {
    /* private mode */
  }
}

function formatOrderStatus(status?: string) {
  if (!status) return 'No status'
  return status.replace(/_/g, ' ')
}

function getProgressStep(
  recommendation: ProcurementRecommendationResponse | null,
  order: Order | null,
) {
  if (!recommendation) return 0
  if (!order) return 1
  if (order.status === 'pending_approval') return 2
  if (order.status === 'approved' && !order.algorandAppId) return 3
  if (order.algorandAppId && order.status !== 'payment_released') return 4
  return 5
}

const tabs: { id: TabId; label: string; icon: typeof LayoutDashboard }[] = [
  { id: 'discover', label: 'Discover', icon: Cpu },
  { id: 'order', label: 'Order', icon: ShoppingCart },
  { id: 'escrow', label: 'Escrow', icon: Link2 },
  { id: 'chain', label: 'On-chain', icon: ArrowRight },
  { id: 'qr', label: 'QR', icon: QrCode },
]

export default function App() {
  const saved = useMemo(() => loadSession(), [])
  const [tab, setTab] = useState<TabId>('discover')
  const [banner, setBanner] = useState<{ ok: boolean; text: string } | null>(null)
  const [recommendation, setRecommendation] = useState<ProcurementRecommendationResponse | null>(
    saved.recommendation ?? null,
  )
  const [order, setOrder] = useState<Order | null>(saved.order ?? null)
  const [selectedVendor, setSelectedVendor] = useState(saved.selectedVendor ?? '')
  const [orderQuantity, setOrderQuantity] = useState(saved.orderQuantity ?? 0)

  useEffect(() => {
    saveSession({
      recommendation,
      order,
      selectedVendor,
      orderQuantity,
    })
  }, [recommendation, order, selectedVendor, orderQuantity])

  const notify = useCallback((message: string, ok = false) => {
    setBanner({ ok, text: message })
    window.setTimeout(() => setBanner(null), ok ? 4500 : 8000)
  }, [])

  const onRecommendation = useCallback(
    (r: ProcurementRecommendationResponse, draft: { quantity: number }) => {
      setRecommendation(r)
      setOrderQuantity(draft.quantity)
    },
    [],
  )

  const progressStep = useMemo(
    () => getProgressStep(recommendation, order),
    [recommendation, order],
  )

  return (
    <div className="min-h-screen">
      <header className="border-b border-slate-800/80 bg-[var(--color-surface-1)]/80 backdrop-blur-md">
        <div className="mx-auto flex max-w-6xl flex-col gap-4 px-4 py-5 sm:flex-row sm:items-center sm:justify-between sm:px-6">
          <div className="flex items-center gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br from-teal-400 to-cyan-600 shadow-lg shadow-teal-500/20">
              <Box className="h-6 w-6 text-slate-950" />
            </div>
            <div>
              <h1 className="font-[family-name:var(--font-display)] text-lg font-semibold tracking-tight text-white">
                Procure AI
              </h1>
              <p className="text-xs text-slate-500">
                Console · API{' '}
                <code className="text-teal-400/90">{getApiBaseUrl()}</code>
              </p>
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-2 text-xs text-slate-500">
            <span className="hidden sm:inline">Flow progress</span>
            <div className="flex h-2 overflow-hidden rounded-full bg-slate-800">
              {[0, 1, 2, 3, 4].map((i) => (
                <div
                  key={i}
                  className={`h-full w-8 transition-colors ${
                    progressStep > i ? 'bg-teal-500' : 'bg-slate-800'
                  }`}
                />
              ))}
            </div>
            <span className="text-slate-600">·</span>
            <LayoutDashboard className="h-3.5 w-3.5" />
            <span>Step {Math.min(5, progressStep + 1)} / 5</span>
          </div>
        </div>
      </header>

      {banner ? (
        <div
          className={`border-b px-4 py-3 text-center text-sm ${
            banner.ok
              ? 'border-teal-500/30 bg-teal-950/40 text-teal-200'
              : 'border-rose-500/30 bg-rose-950/40 text-rose-200'
          }`}
        >
          {banner.text}
        </div>
      ) : null}

      <div className="mx-auto flex max-w-6xl flex-col gap-8 px-4 py-8 lg:flex-row lg:px-6">
        <nav className="lg:w-52 lg:shrink-0">
          <div className="sticky top-6 space-y-1 rounded-2xl border border-slate-800/80 bg-slate-900/40 p-2">
            {tabs.map((t) => {
              const Icon = t.icon
              const active = tab === t.id
              return (
                <button
                  key={t.id}
                  type="button"
                  onClick={() => setTab(t.id)}
                  className={`flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-left text-sm font-medium transition ${
                    active
                      ? 'bg-teal-500/15 text-teal-200 shadow-inner shadow-teal-500/5'
                      : 'text-slate-400 hover:bg-slate-800/60 hover:text-slate-200'
                  }`}
                >
                  <Icon className={`h-4 w-4 shrink-0 ${active ? 'text-teal-400' : ''}`} />
                  {t.label}
                  {active ? <ChevronRight className="ml-auto h-4 w-4 text-teal-500/80" /> : null}
                </button>
              )
            })}
          </div>
          {order ? (
            <div className="mt-4 rounded-2xl border border-slate-800/80 bg-slate-900/30 p-4 text-xs text-slate-500">
              <p className="font-medium text-slate-400">Active order</p>
              <p className="mt-1 font-mono text-teal-300">{order.id}</p>
              <p className="mt-2 capitalize text-slate-400">{formatOrderStatus(order.status)}</p>
              {order.algorandAppId ? (
                <p className="mt-1 font-mono text-[10px] text-slate-600">app {order.algorandAppId}</p>
              ) : null}
            </div>
          ) : null}
        </nav>

        <main className="min-w-0 flex-1">
          {tab === 'discover' && (
            <DiscoverPanel
              recommendation={recommendation}
              selectedVendor={selectedVendor}
              onRecommendation={onRecommendation}
              onSelectVendor={setSelectedVendor}
              notify={notify}
            />
          )}
          {tab === 'order' && (
            <OrderPanel
              recommendation={recommendation}
              selectedVendor={selectedVendor}
              orderQuantity={orderQuantity}
              order={order}
              onOrderCreated={setOrder}
              onOrderApproved={setOrder}
              notify={notify}
            />
          )}
          {tab === 'escrow' && (
            <EscrowPanel order={order} onEscrowCreated={setOrder} notify={notify} />
          )}
          {tab === 'chain' && (
            <ChainPanel
              order={order}
              selectedSupplierName={selectedVendor}
              onOrderUpdate={setOrder}
              notify={notify}
            />
          )}
          {tab === 'qr' && <QrPanel order={order} notify={notify} />}
        </main>
      </div>
    </div>
  )
}
