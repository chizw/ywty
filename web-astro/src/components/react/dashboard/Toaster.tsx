import { CheckCircle2, AlertCircle, AlertTriangle, Info } from 'lucide-react'
import { useToastStore } from '@/lib/react-store'
import type { ToastKind } from '@/lib/react-store'
import { cn } from '@/lib/utils'

const ICONS: Record<ToastKind, React.ComponentType<{ className?: string }>> = {
  success: CheckCircle2,
  error: AlertCircle,
  warning: AlertTriangle,
  info: Info,
}

const CLASSES: Record<ToastKind, string> = {
  success: 'border-green-500/40 bg-green-500/95 text-white',
  error: 'border-destructive/40 bg-destructive/95 text-destructive-foreground',
  warning: 'border-yellow-500/40 bg-yellow-500/95 text-white',
  info: 'border-border bg-card text-card-foreground shadow-lg',
}

export function Toaster() {
  const toasts = useToastStore((s) => s.toasts)
  return (
    <div className="pointer-events-none fixed right-4 top-4 z-[9999] flex max-w-sm flex-col gap-2">
      {toasts.map((t) => {
        const Icon = ICONS[t.kind]
        return (
          <div
            key={t.id}
            className={cn(
              'pointer-events-auto flex items-center gap-2 rounded-md border px-4 py-3 text-sm shadow-lg animate-in fade-in slide-in-from-right-2 duration-200',
              CLASSES[t.kind]
            )}
          >
            <Icon className="h-4 w-4 shrink-0" />
            <span>{t.text}</span>
          </div>
        )
      })}
    </div>
  )
}
