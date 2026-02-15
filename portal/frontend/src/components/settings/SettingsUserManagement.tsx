import { useAuthStore } from '@/store/authStore'
import { GlowCard } from '@/components/shared/GlowCard'
import { SectionHeader } from '@/components/shared/SectionHeader'
import { User } from 'lucide-react'

export function SettingsUserManagement() {
  const user = useAuthStore((s) => s.user)

  return (
    <GlowCard>
      <SectionHeader title="Konto" icon={User} />
      <div className="space-y-3">
        <div className="flex items-center justify-between p-3 rounded-md bg-muted/30 border border-border/30">
          <span className="text-xs text-muted-foreground">Benutzername</span>
          <span className="text-sm font-medium">{user?.username || '—'}</span>
        </div>
        {user?.created_at && (
          <div className="flex items-center justify-between p-3 rounded-md bg-muted/30 border border-border/30">
            <span className="text-xs text-muted-foreground">Erstellt am</span>
            <span className="text-xs font-medium">{new Date(user.created_at).toLocaleDateString()}</span>
          </div>
        )}
        <div className="p-3 rounded-md bg-muted/20 border border-dashed border-border/50">
          <p className="text-xs text-muted-foreground">
            Passwort-Änderung wird in einer zukünftigen Version verfügbar sein.
          </p>
        </div>
      </div>
    </GlowCard>
  )
}
