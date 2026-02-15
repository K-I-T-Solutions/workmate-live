import { cn } from '@/lib/utils'
import type { LucideIcon } from 'lucide-react'

interface MetricCardProps {
  label: string
  value: string | number
  icon: LucideIcon
  iconColor?: string
  subtext?: string
  className?: string
}

export function MetricCard({ label, value, icon: Icon, iconColor = 'text-primary', subtext, className }: MetricCardProps) {
  return (
    <div
      className={cn(
        'flex items-center gap-3 rounded-lg border border-border/50 bg-card p-3 transition-all duration-200 hover:border-border animate-card-enter',
        className
      )}
    >
      <div className={cn('flex items-center justify-center h-9 w-9 rounded-lg bg-muted', iconColor)}>
        <Icon className="h-4 w-4" />
      </div>
      <div className="min-w-0 flex-1">
        <p className="text-xs text-muted-foreground truncate">{label}</p>
        <p className="text-lg font-bold leading-tight">{value}</p>
        {subtext && <p className="text-xs text-muted-foreground truncate">{subtext}</p>}
      </div>
    </div>
  )
}
