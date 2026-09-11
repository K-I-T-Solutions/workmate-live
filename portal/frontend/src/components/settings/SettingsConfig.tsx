import { useState, useEffect, useCallback } from 'react'
import { configAPI } from '@/services/config'
import { authFetch } from '@/lib/api'
import { GlowCard } from '@/components/shared/GlowCard'
import { SectionHeader } from '@/components/shared/SectionHeader'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { AlertTriangle, Save, Eye, EyeOff, Settings2, Tv, MessageCircle, Youtube, Bot, RotateCcw } from 'lucide-react'
import type { PortalConfig, OBSConfig, TwitchConfig, YouTubeConfig, AgentConfig, DefaultUserConfig, ConnectedAgent } from '@/types/config'

type Status = 'idle' | 'saving' | 'success' | 'error'

function StatusMessage({ status, error }: { status: Status; error: string }) {
  if (status === 'saving') return <span className="text-xs text-muted-foreground">Speichern...</span>
  if (status === 'success') return <span className="text-xs text-emerald-400">Gespeichert</span>
  if (status === 'error') return <span className="text-xs text-red-400">{error}</span>
  return null
}

function PasswordInput({ value, onChange, id }: { value: string; onChange: (v: string) => void; id: string }) {
  const [visible, setVisible] = useState(false)
  return (
    <div className="relative">
      <input
        id={id}
        type={visible ? 'text' : 'password'}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full rounded-md border border-border/50 bg-muted/30 px-3 py-2 pr-10 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
      />
      <button
        type="button"
        onClick={() => setVisible(!visible)}
        className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
      >
        {visible ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
      </button>
    </div>
  )
}

function TextInput({ value, onChange, id, type = 'text' }: { value: string; onChange: (v: string) => void; id: string; type?: string }) {
  return (
    <input
      id={id}
      type={type}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="w-full rounded-md border border-border/50 bg-muted/30 px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
    />
  )
}

function NumberInput({ value, onChange, id }: { value: number; onChange: (v: number) => void; id: string }) {
  return (
    <input
      id={id}
      type="number"
      value={value}
      onChange={(e) => onChange(Number(e.target.value))}
      className="w-full rounded-md border border-border/50 bg-muted/30 px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
    />
  )
}

function SaveButton({ status, onClick }: { status: Status; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      disabled={status === 'saving'}
      className="flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
    >
      <Save className="h-4 w-4" />
      Speichern
    </button>
  )
}

function Field({ label, htmlFor, children }: { label: string; htmlFor: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <Label htmlFor={htmlFor} className="text-xs text-muted-foreground">{label}</Label>
      {children}
    </div>
  )
}

function useSection<T extends object>(initial: T) {
  const [data, setData] = useState<T>(initial)
  const [status, setStatus] = useState<Status>('idle')
  const [error, setError] = useState('')

  const save = async (section: string) => {
    setStatus('saving')
    setError('')
    try {
      await configAPI.updateConfig(section, data)
      setStatus('success')
      setTimeout(() => setStatus('idle'), 2000)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Fehler beim Speichern')
      setStatus('error')
    }
  }

  return { data, setData, status, error, save }
}

// --- Section Components ---

function ModeOption({
  active,
  title,
  description,
  onClick,
}: {
  active: boolean
  title: string
  description: string
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={`rounded-lg border p-3 text-left transition ${
        active
          ? 'border-primary bg-primary/10'
          : 'border-border/50 bg-muted/20 hover:border-border'
      }`}
    >
      <div className="text-sm font-medium text-foreground">{title}</div>
      <div className="mt-1 text-xs text-muted-foreground">{description}</div>
    </button>
  )
}

function OBSSection({ initial }: { initial: OBSConfig }) {
  const { data, setData, status, error, save } = useSection(initial)
  const update = (patch: Partial<OBSConfig>) => setData({ ...data, ...patch })

  // Im Agent-Modus verbindet nicht das Portal zu OBS, sondern der Agent —
  // Host, Port und Passwort gehören dann in dessen Konfiguration.
  const viaAgent = data.mode === 'agent'

  return (
    <GlowCard>
      <div className="flex items-center justify-between mb-4">
        <SectionHeader title="OBS WebSocket" icon={Tv} iconColor="text-secondary" className="mb-0" />
        <div className="flex items-center gap-3">
          <StatusMessage status={status} error={error} />
          <SaveButton status={status} onClick={() => save('obs')} />
        </div>
      </div>

      <div className="mb-4 grid sm:grid-cols-2 gap-3">
        <ModeOption
          active={!viaAgent}
          title="Direkt"
          description="Das Portal verbindet sich selbst zum OBS-WebSocket. OBS muss vom Portal aus erreichbar sein."
          onClick={() => update({ mode: 'direct' })}
        />
        <ModeOption
          active={viaAgent}
          title="Über Agent"
          description="Ein Agent auf dem Streaming-Rechner meldet sich beim Portal und steuert das lokale OBS. Kein offener Port nötig."
          onClick={() => update({ mode: 'agent' })}
        />
      </div>

      {viaAgent ? (
        <p className="rounded-md border border-border/50 bg-muted/20 p-3 text-xs text-muted-foreground">
          OBS-Adresse und Passwort werden in der Konfiguration des Agents gesetzt
          (Abschnitt <code className="text-foreground">obs</code>). Der Agent-Schlüssel
          steht im Tab „Agent“.
        </p>
      ) : (
        <div className="grid sm:grid-cols-2 gap-4">
          <Field label="Host" htmlFor="obs-host">
            <TextInput id="obs-host" value={data.host} onChange={(v) => update({ host: v })} />
          </Field>
          <Field label="Port" htmlFor="obs-port">
            <NumberInput id="obs-port" value={data.port} onChange={(v) => update({ port: v })} />
          </Field>
          <Field label="Passwort" htmlFor="obs-password">
            <PasswordInput id="obs-password" value={data.password} onChange={(v) => update({ password: v })} />
          </Field>
          <div className="flex items-center gap-3 pt-5">
            <Switch
              id="obs-reconnect"
              checked={data.auto_reconnect}
              onCheckedChange={(v) => update({ auto_reconnect: v })}
            />
            <Label htmlFor="obs-reconnect" className="text-sm">Auto-Reconnect</Label>
          </div>
        </div>
      )}
    </GlowCard>
  )
}

function TwitchSection({ initial }: { initial: TwitchConfig }) {
  const { data, setData, status, error, save } = useSection(initial)
  const update = (patch: Partial<TwitchConfig>) => setData({ ...data, ...patch })

  return (
    <GlowCard>
      <div className="flex items-center justify-between mb-4">
        <SectionHeader title="Twitch" icon={MessageCircle} iconColor="text-purple-400" className="mb-0" />
        <div className="flex items-center gap-3">
          <StatusMessage status={status} error={error} />
          <SaveButton status={status} onClick={() => save('twitch')} />
        </div>
      </div>
      <div className="flex items-center gap-3 mb-4">
        <Switch
          id="twitch-enabled"
          checked={data.enabled}
          onCheckedChange={(v) => update({ enabled: v })}
        />
        <Label htmlFor="twitch-enabled" className="text-sm">Aktiviert</Label>
      </div>
      <div className="grid sm:grid-cols-2 gap-4">
        <Field label="Client ID" htmlFor="twitch-client-id">
          <TextInput id="twitch-client-id" value={data.client_id} onChange={(v) => update({ client_id: v })} />
        </Field>
        <Field label="Client Secret" htmlFor="twitch-client-secret">
          <PasswordInput id="twitch-client-secret" value={data.client_secret} onChange={(v) => update({ client_secret: v })} />
        </Field>
        <Field label="Channel" htmlFor="twitch-channel">
          <TextInput id="twitch-channel" value={data.channel} onChange={(v) => update({ channel: v })} />
        </Field>
        <Field label="OAuth Token" htmlFor="twitch-oauth">
          <PasswordInput id="twitch-oauth" value={data.oauth_token} onChange={(v) => update({ oauth_token: v })} />
        </Field>
      </div>
    </GlowCard>
  )
}

function YouTubeSection({ initial }: { initial: YouTubeConfig }) {
  const { data, setData, status, error, save } = useSection(initial)
  const update = (patch: Partial<YouTubeConfig>) => setData({ ...data, ...patch })

  return (
    <GlowCard>
      <div className="flex items-center justify-between mb-4">
        <SectionHeader title="YouTube" icon={Youtube} iconColor="text-red-400" className="mb-0" />
        <div className="flex items-center gap-3">
          <StatusMessage status={status} error={error} />
          <SaveButton status={status} onClick={() => save('youtube')} />
        </div>
      </div>
      <div className="flex items-center gap-3 mb-4">
        <Switch
          id="youtube-enabled"
          checked={data.enabled}
          onCheckedChange={(v) => update({ enabled: v })}
        />
        <Label htmlFor="youtube-enabled" className="text-sm">Aktiviert</Label>
      </div>
      <div className="grid sm:grid-cols-2 gap-4">
        <Field label="API Key" htmlFor="youtube-api-key">
          <PasswordInput id="youtube-api-key" value={data.api_key} onChange={(v) => update({ api_key: v })} />
        </Field>
        <Field label="Channel ID" htmlFor="youtube-channel-id">
          <TextInput id="youtube-channel-id" value={data.channel_id} onChange={(v) => update({ channel_id: v })} />
        </Field>
        <Field label="Client ID" htmlFor="youtube-client-id">
          <TextInput id="youtube-client-id" value={data.client_id} onChange={(v) => update({ client_id: v })} />
        </Field>
        <Field label="Client Secret" htmlFor="youtube-client-secret">
          <PasswordInput id="youtube-client-secret" value={data.client_secret} onChange={(v) => update({ client_secret: v })} />
        </Field>
      </div>
    </GlowCard>
  )
}

/** Zeigt die Agents, die gerade über den Link verbunden sind. */
function ConnectedAgents() {
  const [agents, setAgents] = useState<ConnectedAgent[]>([])
  const [linkEnabled, setLinkEnabled] = useState(false)
  const [loading, setLoading] = useState(true)

  const load = useCallback(async () => {
    try {
      const res = await configAPI.getConnectedAgents()
      setAgents(res.agents)
      setLinkEnabled(res.link_enabled)
    } catch {
      setAgents([])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
    const timer = setInterval(load, 5000)
    return () => clearInterval(timer)
  }, [load])

  if (loading) {
    return <p className="text-xs text-muted-foreground">Agents werden geladen...</p>
  }

  if (!linkEnabled) {
    return (
      <p className="text-xs text-muted-foreground">
        Agent-Link deaktiviert. Setze einen Agent-Schlüssel, damit sich Agents verbinden können.
      </p>
    )
  }

  if (agents.length === 0) {
    return <p className="text-xs text-muted-foreground">Kein Agent verbunden.</p>
  }

  return (
    <ul className="space-y-2">
      {agents.map((agent) => (
        <li
          key={agent.agent_id}
          className="flex items-center justify-between rounded-md border border-border/50 bg-muted/20 px-3 py-2"
        >
          <div>
            <div className="text-sm text-foreground">
              {agent.agent_id}
              {agent.hostname && agent.hostname !== agent.agent_id && (
                <span className="ml-2 text-xs text-muted-foreground">{agent.hostname}</span>
              )}
            </div>
            <div className="text-xs text-muted-foreground">
              {agent.version ? `v${agent.version}` : 'Version unbekannt'}
            </div>
          </div>
          <span
            className={`text-xs ${
              agent.obs_active
                ? 'text-emerald-400'
                : agent.obs_local
                  ? 'text-amber-400'
                  : 'text-muted-foreground'
            }`}
          >
            {agent.obs_active ? 'OBS verbunden' : agent.obs_local ? 'OBS getrennt' : 'ohne OBS'}
          </span>
        </li>
      ))}
    </ul>
  )
}

function AgentSection({ initial }: { initial: AgentConfig }) {
  const { data, setData, status, error, save } = useSection(initial)
  const update = (patch: Partial<AgentConfig>) => setData({ ...data, ...patch })

  return (
    <div className="space-y-4">
      <GlowCard>
        <div className="flex items-center justify-between mb-4">
          <SectionHeader title="Agent" icon={Bot} iconColor="text-emerald-400" className="mb-0" />
          <div className="flex items-center gap-3">
            <StatusMessage status={status} error={error} />
            <SaveButton status={status} onClick={() => save('agent')} />
          </div>
        </div>
        <div className="grid sm:grid-cols-2 gap-4">
          <Field label="URL (leer lassen, wenn der Agent sich selbst verbindet)" htmlFor="agent-url">
            <TextInput id="agent-url" value={data.url} onChange={(v) => update({ url: v })} />
          </Field>
          <Field label="Agent-Schlüssel" htmlFor="agent-key">
            <PasswordInput id="agent-key" value={data.api_key} onChange={(v) => update({ api_key: v })} />
          </Field>
          <Field label="Primärer Agent (leer = erster mit OBS)" htmlFor="agent-primary">
            <TextInput id="agent-primary" value={data.primary_id} onChange={(v) => update({ primary_id: v })} />
          </Field>
          <Field label="Kommando-Timeout (ns)" htmlFor="agent-command-timeout">
            <NumberInput
              id="agent-command-timeout"
              value={data.command_timeout}
              onChange={(v) => update({ command_timeout: v })}
            />
          </Field>
          <Field label="Polling Interval (ns)" htmlFor="agent-polling">
            <NumberInput id="agent-polling" value={data.polling_interval} onChange={(v) => update({ polling_interval: v })} />
          </Field>
          <Field label="Timeout (ns)" htmlFor="agent-timeout">
            <NumberInput id="agent-timeout" value={data.timeout} onChange={(v) => update({ timeout: v })} />
          </Field>
        </div>
      </GlowCard>

      <GlowCard>
        <SectionHeader title="Verbundene Agents" icon={Bot} iconColor="text-emerald-400" />
        <ConnectedAgents />
      </GlowCard>
    </div>
  )
}

function AuthSection({ initial }: { initial: DefaultUserConfig }) {
  const { data, setData, status, error, save } = useSection({ default_user: initial })
  const update = (patch: Partial<DefaultUserConfig>) =>
    setData({ default_user: { ...data.default_user, ...patch } })

  return (
    <GlowCard>
      <div className="flex items-center justify-between mb-4">
        <SectionHeader title="Standard-Benutzer" icon={Settings2} iconColor="text-amber-400" className="mb-0" />
        <div className="flex items-center gap-3">
          <StatusMessage status={status} error={error} />
          <SaveButton status={status} onClick={() => save('auth')} />
        </div>
      </div>
      <div className="grid sm:grid-cols-2 gap-4">
        <Field label="Benutzername" htmlFor="auth-username">
          <TextInput id="auth-username" value={data.default_user.username} onChange={(v) => update({ username: v })} />
        </Field>
        <Field label="Passwort" htmlFor="auth-password">
          <PasswordInput id="auth-password" value={data.default_user.password} onChange={(v) => update({ password: v })} />
        </Field>
      </div>
    </GlowCard>
  )
}

// --- Restart Banner ---

function RestartBanner() {
  const [restartStatus, setRestartStatus] = useState<'idle' | 'restarting' | 'waiting' | 'error'>('idle')

  const pollHealth = useCallback(async () => {
    const maxAttempts = 30
    for (let i = 0; i < maxAttempts; i++) {
      await new Promise((resolve) => setTimeout(resolve, 2000))
      try {
        const res = await fetch('/health')
        if (res.ok) {
          window.location.reload()
          return
        }
      } catch {
        // Backend noch nicht bereit
      }
    }
    setRestartStatus('error')
  }, [])

  const handleRestart = async () => {
    setRestartStatus('restarting')
    try {
      await authFetch('/api/restart', { method: 'POST' })
      setRestartStatus('waiting')
      pollHealth()
    } catch {
      setRestartStatus('error')
    }
  }

  return (
    <GlowCard glowColor="destructive">
      <div className="flex items-start gap-3">
        <AlertTriangle className="h-5 w-5 text-amber-400 shrink-0 mt-0.5" />
        <div className="flex-1">
          <p className="text-sm font-medium text-amber-400">Neustart erforderlich</p>
          <p className="text-xs text-muted-foreground mt-1">
            Aenderungen an Service-Konfigurationen (OBS, Twitch, YouTube) werden erst nach einem Neustart des Backends wirksam.
          </p>
        </div>
        {restartStatus === 'idle' && (
          <button
            onClick={handleRestart}
            className="flex items-center gap-2 rounded-md bg-amber-500/20 border border-amber-500/30 px-3 py-1.5 text-sm font-medium text-amber-400 hover:bg-amber-500/30 shrink-0"
          >
            <RotateCcw className="h-4 w-4" />
            Neustart
          </button>
        )}
        {restartStatus === 'restarting' && (
          <span className="text-xs text-amber-400 shrink-0">Wird neu gestartet...</span>
        )}
        {restartStatus === 'waiting' && (
          <span className="text-xs text-amber-400 shrink-0 animate-pulse">Warte auf Backend...</span>
        )}
        {restartStatus === 'error' && (
          <button
            onClick={handleRestart}
            className="flex items-center gap-2 rounded-md bg-red-500/20 border border-red-500/30 px-3 py-1.5 text-sm font-medium text-red-400 hover:bg-red-500/30 shrink-0"
          >
            <RotateCcw className="h-4 w-4" />
            Erneut versuchen
          </button>
        )}
      </div>
    </GlowCard>
  )
}

// --- Main Component ---

export function SettingsConfig() {
  const [config, setConfig] = useState<PortalConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    configAPI
      .getConfig()
      .then(setConfig)
      .catch((e) => setError(e instanceof Error ? e.message : 'Fehler beim Laden'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <GlowCard>
        <p className="text-sm text-muted-foreground">Konfiguration wird geladen...</p>
      </GlowCard>
    )
  }

  if (error || !config) {
    return (
      <GlowCard glowColor="destructive">
        <p className="text-sm text-red-400">{error || 'Konfiguration konnte nicht geladen werden'}</p>
      </GlowCard>
    )
  }

  return (
    <div className="space-y-4">
      <RestartBanner />

      <Tabs defaultValue="obs">
        <TabsList>
          <TabsTrigger value="obs">OBS</TabsTrigger>
          <TabsTrigger value="twitch">Twitch</TabsTrigger>
          <TabsTrigger value="youtube">YouTube</TabsTrigger>
          <TabsTrigger value="agent">Agent</TabsTrigger>
          <TabsTrigger value="auth">Auth</TabsTrigger>
        </TabsList>

        <TabsContent value="obs">
          <OBSSection initial={config.obs} />
        </TabsContent>
        <TabsContent value="twitch">
          <TwitchSection initial={config.twitch} />
        </TabsContent>
        <TabsContent value="youtube">
          <YouTubeSection initial={config.youtube} />
        </TabsContent>
        <TabsContent value="agent">
          <AgentSection initial={config.agent} />
        </TabsContent>
        <TabsContent value="auth">
          <AuthSection initial={config.auth.default_user} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
