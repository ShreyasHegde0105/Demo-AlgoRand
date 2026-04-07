import { Sparkles, Wand2 } from 'lucide-react'
import { useState } from 'react'
import { apiPost } from '../api/client'
import type { ProcurementRecommendationResponse } from '../api/types'
import { Button, Field, Panel, TextArea, TextInput } from './Ui'
import { VendorCard } from './VendorCard'

type ProcurementForm = {
  category: string
  quantity: number
  budget: number
  maxDeliveryDays: number
  preferredCities: string
  topN: number
}

const defaultForm: ProcurementForm = {
  category: 'electronics',
  quantity: 10,
  budget: 50000,
  maxDeliveryDays: 7,
  preferredCities: 'Bangalore, Mumbai',
  topN: 3,
}

export function DiscoverPanel({
  recommendation,
  selectedVendor,
  onRecommendation,
  onSelectVendor,
  notify,
}: {
  recommendation: ProcurementRecommendationResponse | null
  selectedVendor: string
  onRecommendation: (r: ProcurementRecommendationResponse, draft: ProcurementForm) => void
  onSelectVendor: (name: string) => void
  notify: (message: string, ok?: boolean) => void
}) {
  const [mode, setMode] = useState<'structured' | 'natural'>('structured')
  const [loading, setLoading] = useState(false)
  const [form, setForm] = useState<ProcurementForm>(defaultForm)
  const [prompt, setPrompt] = useState(
    'Need 25 laptops for our Mumbai office under 40 lakh with delivery within 2 weeks.',
  )

  async function runStructured() {
    setLoading(true)
    try {
      const cities = form.preferredCities
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean)
      const res = await apiPost<ProcurementRecommendationResponse>('/agent/recommend-vendors', {
        category: form.category,
        quantity: form.quantity,
        budget: form.budget,
        maxDeliveryDays: form.maxDeliveryDays,
        preferredCities: cities,
        topN: form.topN,
      })
      onRecommendation(res, form)
      notify('Shortlist ready. Pick a supplier for your order.', true)
    } catch (e) {
      notify(e instanceof Error ? e.message : 'Recommendation failed')
    } finally {
      setLoading(false)
    }
  }

  async function runNatural() {
    setLoading(true)
    try {
      const res = await apiPost<{
        recommendation: ProcurementRecommendationResponse
        request: {
          category: string
          quantity: number
          budget: number
          maxDeliveryDays: number
          preferredCities: string[]
          topN: number
        }
      }>('/agent/parse-and-recommend', { prompt, topN: form.topN })

      if (!res.recommendation) {
        notify('Parser returned no recommendation')
        return
      }
      const draft: ProcurementForm = {
        category: res.request.category,
        quantity: res.request.quantity,
        budget: res.request.budget,
        maxDeliveryDays: res.request.maxDeliveryDays,
        preferredCities: res.request.preferredCities.join(', '),
        topN: res.request.topN || form.topN,
      }
      setForm(draft)
      onRecommendation(res.recommendation, draft)
      notify('Parsed your brief and ranked suppliers.', true)
    } catch (e) {
      notify(
        e instanceof Error
          ? e.message
          : 'Natural-language flow failed (is GEMINI_API_KEY set on the server?)',
      )
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-8">
      <Panel
        title="Procurement brief"
        subtitle="Describe what you need. The agent scores suppliers on price, delivery, trust, and reliability."
      >
        <div className="mb-6 flex flex-wrap gap-2 rounded-xl bg-slate-900/50 p-1">
          <button
            type="button"
            onClick={() => setMode('structured')}
            className={`flex flex-1 items-center justify-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition sm:flex-none ${
              mode === 'structured'
                ? 'bg-slate-800 text-white shadow'
                : 'text-slate-500 hover:text-slate-300'
            }`}
          >
            <Wand2 className="h-4 w-4 text-teal-400" />
            Structured
          </button>
          <button
            type="button"
            onClick={() => setMode('natural')}
            className={`flex flex-1 items-center justify-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition sm:flex-none ${
              mode === 'natural'
                ? 'bg-slate-800 text-white shadow'
                : 'text-slate-500 hover:text-slate-300'
            }`}
          >
            <Sparkles className="h-4 w-4 text-violet-400" />
            Natural language
          </button>
        </div>

        {mode === 'structured' ? (
          <div className="grid gap-4 sm:grid-cols-2">
            <Field label="Category">
              <TextInput
                value={form.category}
                onChange={(e) => setForm((f) => ({ ...f, category: e.target.value }))}
              />
            </Field>
            <Field label="Quantity">
              <TextInput
                type="number"
                min={1}
                value={form.quantity}
                onChange={(e) =>
                  setForm((f) => ({ ...f, quantity: Number.parseInt(e.target.value, 10) || 1 }))
                }
              />
            </Field>
            <Field label="Budget (₹)">
              <TextInput
                type="number"
                min={0}
                value={form.budget}
                onChange={(e) =>
                  setForm((f) => ({ ...f, budget: Number.parseFloat(e.target.value) || 0 }))
                }
              />
            </Field>
            <Field label="Max delivery (days)">
              <TextInput
                type="number"
                min={0}
                value={form.maxDeliveryDays}
                onChange={(e) =>
                  setForm((f) => ({
                    ...f,
                    maxDeliveryDays: Number.parseInt(e.target.value, 10) || 0,
                  }))
                }
              />
            </Field>
            <Field label="Preferred cities" hint="Comma-separated">
              <TextInput
                value={form.preferredCities}
                onChange={(e) => setForm((f) => ({ ...f, preferredCities: e.target.value }))}
              />
            </Field>
            <Field label="Shortlist size">
              <TextInput
                type="number"
                min={1}
                max={10}
                value={form.topN}
                onChange={(e) =>
                  setForm((f) => ({ ...f, topN: Number.parseInt(e.target.value, 10) || 3 }))
                }
              />
            </Field>
            <div className="sm:col-span-2">
              <Button disabled={loading} onClick={() => void runStructured()}>
                {loading ? 'Scoring vendors…' : 'Run agent & build shortlist'}
              </Button>
            </div>
          </div>
        ) : (
          <div className="space-y-4">
            <Field label="Describe the purchase" hint="Requires GEMINI_API_KEY on the backend.">
              <TextArea value={prompt} onChange={(e) => setPrompt(e.target.value)} />
            </Field>
            <Field label="Shortlist size">
              <TextInput
                type="number"
                min={1}
                max={10}
                value={form.topN}
                onChange={(e) =>
                  setForm((f) => ({ ...f, topN: Number.parseInt(e.target.value, 10) || 3 }))
                }
              />
            </Field>
            <Button disabled={loading} onClick={() => void runNatural()}>
              {loading ? 'Parsing…' : 'Parse with Gemini & recommend'}
            </Button>
          </div>
        )}
      </Panel>

      {recommendation ? (
        <Panel title="Ranked shortlist" subtitle={recommendation.summary}>
          <div className="mb-4 flex flex-wrap gap-3 text-xs text-slate-500">
            <span>
              Session ID:{' '}
              <code className="text-teal-300">{recommendation.recommendationId ?? 'n/a'}</code>
            </span>
          </div>
          <div className="grid gap-4 lg:grid-cols-3">
            {recommendation.topVendors.map((rv) => (
              <VendorCard
                key={rv.vendor.id}
                ranked={rv}
                selected={selectedVendor === rv.vendor.name}
                onSelect={() => onSelectVendor(rv.vendor.name)}
              />
            ))}
          </div>
        </Panel>
      ) : null}
    </div>
  )
}
