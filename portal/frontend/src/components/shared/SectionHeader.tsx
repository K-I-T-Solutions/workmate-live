import { cn } from '@/lib/utils'
import type { LucideIcon } from 'lucide-react'

interface SectionHeaderProps {
  title: string
  icon?: LucideIcon
  iconColor?: string
  className?: string
  children?: React.ReactNode
}

export function SectionHeader({ title, icon: Icon, iconColor = 'text-primary', className, children }: SectionHeaderProps) {
  return (
    <div className={cn('flex items-center justify-between mb-3', className)}>
      <h2 className="flex items-center gap-2 text-sm font-semibold text-foreground">
        {Icon && <Icon className={cn('h-4 w-4', iconColor)} />}
        {title}
      </h2>
      {children}
    </div>
  )
}
