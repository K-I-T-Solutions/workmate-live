import { cn } from '@/lib/utils'

interface StatusIndicatorProps {
  connected: boolean
  size?: 'sm' | 'md' | 'lg'
  pulse?: boolean
  className?: string
}

const sizeMap = {
  sm: 'h-1.5 w-1.5',
  md: 'h-2 w-2',
  lg: 'h-3 w-3',
}

export function StatusIndicator({ connected, size = 'md', pulse = true, className }: StatusIndicatorProps) {
  return (
    <span className={cn('relative inline-flex', className)}>
      {pulse && connected && (
        <span
          className={cn(
            'absolute inline-flex h-full w-full rounded-full opacity-75 animate-ping',
            'bg-success'
          )}
        />
      )}
      <span
        className={cn(
          'relative inline-flex rounded-full',
          sizeMap[size],
          connected
            ? 'bg-success shadow-[0_0_6px_rgba(16,185,129,0.6)]'
            : 'bg-destructive shadow-[0_0_6px_rgba(239,68,68,0.6)]'
        )}
      />
    </span>
  )
}
