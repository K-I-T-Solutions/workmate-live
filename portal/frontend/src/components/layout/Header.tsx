import { useAuthStore } from '@/store/authStore'
import { useConnectionStatus } from '@/hooks/useConnectionStatus'
import { authAPI } from '@/services/auth'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { LogOut } from 'lucide-react'
import { cn } from '@/lib/utils'

export function Header() {
  const { token, user, clearAuth } = useAuthStore()
  const { connections } = useConnectionStatus()

  const handleLogout = async () => {
    if (token) {
      try {
        await authAPI.logout(token)
      } catch (err) {
        console.error('Logout error:', err)
      }
    }
    clearAuth()
  }

  return (
    <header className="flex items-center justify-between h-12 px-4 border-b border-border/50 bg-card/50 backdrop-blur-sm">
      {/* Status Dots */}
      <div className="flex items-center gap-3">
        {connections.map((conn) => (
          <Tooltip key={conn.label} delayDuration={0}>
            <TooltipTrigger asChild>
              <div className="flex items-center gap-1.5 cursor-default">
                <div
                  className={cn(
                    'h-2 w-2 rounded-full transition-colors',
                    conn.connected
                      ? 'bg-success shadow-[0_0_6px_rgba(16,185,129,0.6)]'
                      : 'bg-destructive shadow-[0_0_6px_rgba(239,68,68,0.6)]'
                  )}
                />
                <span className="text-xs text-muted-foreground hidden sm:inline">
                  {conn.label}
                </span>
              </div>
            </TooltipTrigger>
            <TooltipContent>
              {conn.label}: {conn.connected ? 'Connected' : 'Disconnected'}
              {conn.detail && ` — ${conn.detail}`}
            </TooltipContent>
          </Tooltip>
        ))}
      </div>

      {/* User */}
      <div className="flex items-center gap-3">
        <span className="text-xs text-muted-foreground">
          {user?.username}
        </span>
        <Button variant="ghost" size="icon" onClick={handleLogout} className="h-7 w-7 text-muted-foreground hover:text-destructive">
          <LogOut className="h-3.5 w-3.5" />
        </Button>
      </div>
    </header>
  )
}
