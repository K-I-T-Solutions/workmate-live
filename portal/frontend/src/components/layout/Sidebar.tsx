import { useEffect } from 'react'
import { useUIStore } from '@/store/uiStore'
import { SidebarItem } from './SidebarItem'
import { Separator } from '@/components/ui/separator'
import { LayoutDashboard, Video, Radio, Youtube, Settings, PanelLeftClose, PanelLeft } from 'lucide-react'
import { cn } from '@/lib/utils'

const navItems = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/obs', icon: Video, label: 'OBS Studio' },
  { to: '/twitch', icon: Radio, label: 'Twitch' },
  { to: '/youtube', icon: Youtube, label: 'YouTube' },
]

const bottomItems = [
  { to: '/settings', icon: Settings, label: 'Settings' },
]

export function Sidebar() {
  const { sidebarCollapsed, toggleSidebar, setSidebarCollapsed } = useUIStore()

  // Auto-collapse on small screens
  useEffect(() => {
    const mql = window.matchMedia('(max-width: 1024px)')
    const handleChange = (e: MediaQueryListEvent | MediaQueryList) => {
      if (e.matches) setSidebarCollapsed(true)
    }
    handleChange(mql)
    mql.addEventListener('change', handleChange)
    return () => mql.removeEventListener('change', handleChange)
  }, [setSidebarCollapsed])

  return (
    <aside
      className={cn(
        'flex h-screen flex-col border-r border-border/50 bg-sidebar transition-all duration-300',
        sidebarCollapsed ? 'w-16' : 'w-60'
      )}
    >
      {/* Logo / Brand */}
      <div className={cn('flex items-center h-12 px-3 border-b border-border/50', sidebarCollapsed ? 'justify-center' : 'gap-2')}>
        {!sidebarCollapsed && (
          <span className="text-sm font-bold text-primary text-glow-primary truncate">
            Workmate Live
          </span>
        )}
        <button
          onClick={toggleSidebar}
          className={cn(
            'p-1.5 rounded-md text-muted-foreground hover:text-foreground hover:bg-accent transition-colors',
            sidebarCollapsed ? '' : 'ml-auto'
          )}
        >
          {sidebarCollapsed ? (
            <PanelLeft className="h-4 w-4" />
          ) : (
            <PanelLeftClose className="h-4 w-4" />
          )}
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 flex flex-col gap-1 p-2">
        {navItems.map((item) => (
          <SidebarItem key={item.to} {...item} collapsed={sidebarCollapsed} />
        ))}
      </nav>

      {/* Bottom */}
      <div className="p-2">
        <Separator className="mb-2" />
        {bottomItems.map((item) => (
          <SidebarItem key={item.to} {...item} collapsed={sidebarCollapsed} />
        ))}
      </div>
    </aside>
  )
}
