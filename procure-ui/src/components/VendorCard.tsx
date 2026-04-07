import { MapPin, Package } from 'lucide-react'
import type { RankedVendor } from '../api/types'
import { Button } from './Ui'

export function VendorCard({
  ranked,
  selected,
  onSelect,
}: {
  ranked: RankedVendor
  selected: boolean
  onSelect: () => void
}) {
  const { vendor, scoreBreakdown, estimatedTotal, reason, rank } = ranked
  const pct = Math.min(100, Math.round(scoreBreakdown.final * 100))

  return (
    <article
      className={`relative overflow-hidden rounded-2xl border p-5 transition ${
        selected
          ? 'border-teal-400/70 bg-teal-950/25 shadow-lg shadow-teal-500/10'
          : 'border-slate-700/50 bg-slate-900/30 hover:border-slate-600'
      }`}
    >
      <div className="absolute right-3 top-3 flex h-8 w-8 items-center justify-center rounded-full bg-slate-800 text-xs font-bold text-teal-300">
        {rank}
      </div>
      <h3 className="font-[family-name:var(--font-display)] pr-10 text-base font-semibold text-white">
        {vendor.name}
      </h3>
      <div className="mt-2 flex flex-wrap gap-3 text-xs text-slate-400">
        <span className="inline-flex items-center gap-1">
          <MapPin className="h-3.5 w-3.5" />
          {vendor.location || 'n/a'}
        </span>
        <span className="inline-flex items-center gap-1">
          <Package className="h-3.5 w-3.5" />
          Stock {vendor.stock}
        </span>
      </div>
      <div className="mt-4">
        <div className="mb-1 flex justify-between text-xs">
          <span className="text-slate-500">Agent score</span>
          <span className="font-mono text-teal-300">{pct}%</span>
        </div>
        <div className="h-1.5 overflow-hidden rounded-full bg-slate-800">
          <div
            className="h-full rounded-full bg-gradient-to-r from-teal-600 to-teal-300 transition-[width] duration-500"
            style={{ width: `${pct}%` }}
          />
        </div>
      </div>
      <p className="mt-3 text-sm leading-relaxed text-slate-400">{reason}</p>
      <div className="mt-4 flex items-center justify-between gap-3 border-t border-slate-800/80 pt-4">
        <div>
          <p className="text-xs text-slate-500">Est. total</p>
          <p className="font-mono text-lg font-semibold text-white">
            ₹{estimatedTotal.toLocaleString(undefined, { maximumFractionDigits: 0 })}
          </p>
        </div>
        <Button variant={selected ? 'secondary' : 'primary'} onClick={onSelect}>
          {selected ? 'Selected' : 'Choose supplier'}
        </Button>
      </div>
    </article>
  )
}
