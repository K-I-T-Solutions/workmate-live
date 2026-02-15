import { cn } from '@/lib/utils'

interface PageContainerProps {
  title: string
  subtitle?: string
  children: React.ReactNode
  className?: string
}

export function PageContainer({ title, subtitle, children, className }: PageContainerProps) {
  return (
    <div className={cn('space-y-4 max-w-7xl', className)}>
      <div>
        <h1 className="text-lg font-bold">{title}</h1>
        {subtitle && <p className="text-xs text-muted-foreground mt-0.5">{subtitle}</p>}
      </div>
      {children}
    </div>
  )
}
