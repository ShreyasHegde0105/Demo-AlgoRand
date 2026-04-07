import type { ReactNode } from 'react'

export function Panel({
  title,
  subtitle,
  children,
  className = '',
}: {
  title: string
  subtitle?: string
  children: ReactNode
  className?: string
}) {
  return (
    <section
      className={`rounded-2xl border border-slate-700/50 bg-[var(--color-surface-1)]/90 p-6 shadow-xl shadow-black/25 backdrop-blur-md ${className}`}
    >
      <div className="mb-5">
        <h2 className="font-[family-name:var(--font-display)] text-lg font-semibold tracking-tight text-white">
          {title}
        </h2>
        {subtitle ? <p className="mt-1 text-sm text-slate-400">{subtitle}</p> : null}
      </div>
      {children}
    </section>
  )
}

export function Button({
  children,
  onClick,
  disabled,
  variant = 'primary',
  type = 'button',
  className = '',
}: {
  children: ReactNode
  onClick?: () => void
  disabled?: boolean
  variant?: 'primary' | 'secondary' | 'ghost'
  type?: 'button' | 'submit'
  className?: string
}) {
  const styles =
    variant === 'primary'
      ? 'bg-teal-400 text-slate-950 shadow-lg shadow-teal-500/20 hover:bg-teal-300'
      : variant === 'secondary'
        ? 'border border-slate-600 bg-slate-800/80 text-slate-100 hover:bg-slate-700/80'
        : 'text-slate-300 hover:bg-slate-800/60 hover:text-white'
  return (
    <button
      type={type}
      disabled={disabled}
      onClick={onClick}
      className={`inline-flex items-center justify-center gap-2 rounded-xl px-4 py-2.5 text-sm font-semibold transition disabled:pointer-events-none disabled:opacity-45 ${styles} ${className}`}
    >
      {children}
    </button>
  )
}

export function Field({
  label,
  children,
  hint,
}: {
  label: string
  children: ReactNode
  hint?: string
}) {
  return (
    <label className="block space-y-1.5">
      <span className="text-xs font-medium uppercase tracking-wider text-slate-500">{label}</span>
      {children}
      {hint ? <span className="block text-xs text-slate-500">{hint}</span> : null}
    </label>
  )
}

export function TextInput(props: React.InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      {...props}
      className={`w-full rounded-xl border border-slate-600/80 bg-slate-900/60 px-3 py-2 text-sm text-slate-100 outline-none ring-teal-500/0 transition placeholder:text-slate-600 focus:border-teal-500/50 focus:ring-2 focus:ring-teal-500/30 ${props.className ?? ''}`}
    />
  )
}

export function TextArea(props: React.TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return (
    <textarea
      {...props}
      className={`min-h-[100px] w-full resize-y rounded-xl border border-slate-600/80 bg-slate-900/60 px-3 py-2 text-sm text-slate-100 outline-none ring-teal-500/0 transition placeholder:text-slate-600 focus:border-teal-500/50 focus:ring-2 focus:ring-teal-500/30 ${props.className ?? ''}`}
    />
  )
}
