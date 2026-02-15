import { GlowCard } from '@/components/shared/GlowCard'
import { SectionHeader } from '@/components/shared/SectionHeader'
import { Info } from 'lucide-react'

export function SettingsAbout() {
  return (
    <GlowCard>
      <SectionHeader title="Über" icon={Info} />
      <div className="space-y-2">
        <div className="flex items-center justify-between p-2 rounded-md bg-muted/30 border border-border/30">
          <span className="text-xs text-muted-foreground">Portal</span>
          <span className="text-xs font-medium">Workmate Live Portal</span>
        </div>
        <div className="flex items-center justify-between p-2 rounded-md bg-muted/30 border border-border/30">
          <span className="text-xs text-muted-foreground">Version</span>
          <span className="text-xs font-mono">0.1.0</span>
        </div>
        <div className="flex items-center justify-between p-2 rounded-md bg-muted/30 border border-border/30">
          <span className="text-xs text-muted-foreground">Stack</span>
          <span className="text-xs font-medium">React 19 + Vite 7 + Tailwind 4</span>
        </div>
      </div>
    </GlowCard>
  )
}
