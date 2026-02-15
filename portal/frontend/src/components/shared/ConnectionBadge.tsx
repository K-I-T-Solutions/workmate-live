import { cn } from '@/lib/utils'
import { StatusIndicator } from './StatusIndicator'

interface ConnectionBadgeProps {
  label: string
  connected: boolean
  detail?: string
  className?: string
}

export function ConnectionBadge({ label, connected, detail, className }: ConnectionBadgeProps) {
  return (
    <div
      className={cn(
        'inline-flex items-center gap-2 rounded-full border px-3 py-1 text-xs font-medium transition-colors',
        connected
          ? 'border-success/30 bg-success/10 text-success'
          : 'border-destructive/30 bg-destructive/10 text-destructive',
        className
      )}
    >
      <StatusIndicator connected={connected} size="sm" pulse={false} />
      <span>{label}</span>
      {detail && <span className="text-muted-foreground">({detail})</span>}
    </div>
  )
}
