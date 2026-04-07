import { Check, Copy } from 'lucide-react'
import { useState } from 'react'
import { Button } from './Ui'

export function CopyBlock({
  label,
  value,
  truncate = true,
}: {
  label: string
  value: string
  truncate?: boolean
}) {
  const [copied, setCopied] = useState(false)

  async function copy() {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      /* ignore */
    }
  }

  return (
    <div className="rounded-xl border border-slate-700/60 bg-slate-950/40 p-3">
      <div className="mb-2 flex items-center justify-between gap-2">
        <span className="text-xs font-medium text-slate-500">{label}</span>
        <Button variant="ghost" className="!px-2 !py-1 text-xs" onClick={copy}>
          {copied ? (
            <>
              <Check className="h-3.5 w-3.5 text-teal-400" /> Copied
            </>
          ) : (
            <>
              <Copy className="h-3.5 w-3.5" /> Copy
            </>
          )}
        </Button>
      </div>
      <pre
        className={`max-h-40 overflow-auto whitespace-pre-wrap break-all font-mono text-[11px] leading-relaxed text-slate-300 ${truncate ? 'line-clamp-6' : ''}`}
      >
        {value || 'n/a'}
      </pre>
    </div>
  )
}
