import { useConnectionStatus } from '@/hooks/useConnectionStatus'
import { ConnectionBadge } from '@/components/shared/ConnectionBadge'
import { GlowCard } from '@/components/shared/GlowCard'
import { SectionHeader } from '@/components/shared/SectionHeader'
import { Wifi } from 'lucide-react'

export function SettingsConnectionStatus() {
  const { connections, connectedCount } = useConnectionStatus()

  return (
    <GlowCard>
      <SectionHeader title="Verbindungen" icon={Wifi} iconColor="text-secondary">
        <span className="text-xs text-muted-foreground">
          {connectedCount}/{connections.length} verbunden
        </span>
      </SectionHeader>
      <div className="grid sm:grid-cols-2 gap-3">
        {connections.map((conn) => (
          <div
            key={conn.label}
            className="flex items-center justify-between p-3 rounded-lg bg-muted/30 border border-border/30"
          >
            <div>
              <p className="text-sm font-medium">{conn.label}</p>
              {conn.detail && (
                <p className="text-xs text-muted-foreground mt-0.5">{conn.detail}</p>
              )}
            </div>
            <ConnectionBadge
              label={conn.connected ? 'Online' : 'Offline'}
              connected={conn.connected}
            />
          </div>
        ))}
      </div>
    </GlowCard>
  )
}
