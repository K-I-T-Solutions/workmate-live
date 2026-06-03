import { NavLink } from 'react-router-dom'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
import type { LucideIcon } from 'lucide-react'

interface SidebarItemProps {
  to: string
  icon: LucideIcon
  label: string
  collapsed: boolean
}

export function SidebarItem({ to, icon: Icon, label, collapsed }: SidebarItemProps) {
  const linkContent = (
    <NavLink
      to={to}
      className={({ isActive }) =>
        cn(
          'flex items-center rounded-lg py-2 text-sm font-medium transition-all duration-200',
          collapsed ? 'justify-center px-0' : 'gap-3 px-3',
          isActive
            ? 'bg-primary/10 text-primary shadow-[0_0_15px_rgba(204,0,255,0.15)]'
            : 'text-muted-foreground hover:bg-accent hover:text-foreground'
        )
      }
    >
      <Icon className="h-5 w-5 shrink-0" />
      {!collapsed && <span className="animate-sidebar-slide">{label}</span>}
    </NavLink>
  )

  if (collapsed) {
    return (
      <Tooltip delayDuration={0}>
        <TooltipTrigger asChild>{linkContent}</TooltipTrigger>
        <TooltipContent side="right">{label}</TooltipContent>
      </Tooltip>
    )
  }

  return linkContent
}
