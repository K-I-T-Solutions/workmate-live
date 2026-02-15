import { useAgentStore } from '@/store/agentStore'
import { GlowCard } from '@/components/shared/GlowCard'
import { SectionHeader } from '@/components/shared/SectionHeader'
import { Server } from 'lucide-react'

export function SettingsSystemInfo() {
  const status = useAgentStore((s) => s.status)

  if (!status) {
    return (
      <GlowCard>
        <SectionHeader title="System" icon={Server} />
        <p className="text-xs text-muted-foreground">Agent nicht verbunden — keine Systemdaten verfügbar.</p>
      </GlowCard>
    )
  }

  const items = [
    { label: 'Hostname', value: status.hostname },
    { label: 'Display', value: status.headless ? 'Headless' : 'GUI' },
    { label: 'GPU', value: status.gpu.present ? (status.gpu.vendors?.join(', ') || 'Detected') : 'Not detected' },
    { label: 'Audio Backend', value: status.audio.backend },
    { label: 'Audio Ready', value: status.audio.ready ? 'Yes' : 'No' },
    { label: 'Video Devices', value: `${status.video.device_count}` },
    { label: 'OBS Running', value: status.obs.running ? 'Yes' : 'No' },
    { label: 'Last Update', value: new Date(status.timestamp).toLocaleString() },
  ]

  if (status.gpu.render_nodes && status.gpu.render_nodes.length > 0) {
    items.push({ label: 'Render Nodes', value: status.gpu.render_nodes.join(', ') })
  }

  return (
    <GlowCard>
      <SectionHeader title="System-Informationen" icon={Server} />
      <div className="grid sm:grid-cols-2 gap-2">
        {items.map((item) => (
          <div key={item.label} className="flex items-center justify-between p-2 rounded-md bg-muted/30 border border-border/30">
            <span className="text-xs text-muted-foreground">{item.label}</span>
            <span className="text-xs font-medium text-right max-w-[60%] truncate">{item.value}</span>
          </div>
        ))}
      </div>
    </GlowCard>
  )
}
