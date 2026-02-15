import { cn } from '@/lib/utils'

interface GlowCardProps extends React.HTMLAttributes<HTMLDivElement> {
  glowColor?: 'primary' | 'secondary' | 'success' | 'destructive'
}

const glowMap = {
  primary: 'hover:shadow-[0_0_25px_rgba(124,58,237,0.15)]',
  secondary: 'hover:shadow-[0_0_25px_rgba(6,182,212,0.15)]',
  success: 'hover:shadow-[0_0_25px_rgba(16,185,129,0.15)]',
  destructive: 'hover:shadow-[0_0_25px_rgba(239,68,68,0.15)]',
}

export function GlowCard({ glowColor = 'primary', className, children, ...props }: GlowCardProps) {
  return (
    <div
      className={cn(
        'rounded-lg border border-border/50 bg-card p-4 transition-all duration-300 animate-card-enter',
        glowMap[glowColor],
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}
