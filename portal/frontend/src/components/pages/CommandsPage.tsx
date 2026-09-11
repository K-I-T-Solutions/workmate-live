import { useEffect, useState, useRef } from 'react'
import { PageContainer } from '@/components/shared/PageContainer'
import { GlowCard } from '@/components/shared/GlowCard'
import { SectionHeader } from '@/components/shared/SectionHeader'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { twitchAPI } from '@/services/twitch'
import type { CommandInfo, CommandResult, CreateCommandRequest } from '@/types/twitch'
import {
  Terminal, Play, Send, Shield, Clock, List, MessageSquare,
  Loader2, Plus, Pencil, Trash2, Save, X, Hash,
} from 'lucide-react'
import { cn } from '@/lib/utils'

interface LogEntry {
  id: number
  timestamp: string
  command: string
  response: string
  success: boolean
  source: 'command' | 'message'
}

let logIdCounter = 0

const emptyForm: CreateCommandRequest = {
  name: '',
  description: '',
  response: '',
  mod_only: false,
  cooldown: 0,
  enabled: true,
}

export function CommandsPage() {
  const [commands, setCommands] = useState<CommandInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [executingCmd, setExecutingCmd] = useState<string | null>(null)
  const [chatMessage, setChatMessage] = useState('')
  const [sendingMessage, setSendingMessage] = useState(false)
  const [log, setLog] = useState<LogEntry[]>([])
  const logEndRef = useRef<HTMLDivElement>(null)

  // CRUD state
  const [editMode, setEditMode] = useState<'create' | 'edit' | null>(null)
  const [editingName, setEditingName] = useState<string | null>(null)
  const [form, setForm] = useState<CreateCommandRequest>({ ...emptyForm })
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState<string | null>(null)

  const loadCommands = () => {
    twitchAPI.listCommands()
      .then(setCommands)
      .catch(() => {})
      .finally(() => setLoading(false))
  }

  useEffect(() => { loadCommands() }, [])

  useEffect(() => {
    logEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [log])

  const addLog = (entry: Omit<LogEntry, 'id' | 'timestamp'>) => {
    setLog(prev => [...prev, {
      ...entry,
      id: ++logIdCounter,
      timestamp: new Date().toLocaleTimeString('de-DE'),
    }])
  }

  const executeCommand = async (name: string) => {
    setExecutingCmd(name)
    try {
      const result: CommandResult = await twitchAPI.executeCommand(name)
      addLog({
        command: `!${result.command}`,
        response: result.response,
        success: result.success,
        source: 'command',
      })
    } catch {
      addLog({
        command: `!${name}`,
        response: 'Fehler beim Ausführen',
        success: false,
        source: 'command',
      })
    } finally {
      setExecutingCmd(null)
    }
  }

  const sendMessage = async (e: React.FormEvent) => {
    e.preventDefault()
    const msg = chatMessage.trim()
    if (!msg) return

    setSendingMessage(true)
    try {
      await twitchAPI.sendMessage(msg)
      addLog({
        command: msg,
        response: 'Nachricht gesendet',
        success: true,
        source: 'message',
      })
      setChatMessage('')
    } catch {
      addLog({
        command: msg,
        response: 'Fehler beim Senden',
        success: false,
        source: 'message',
      })
    } finally {
      setSendingMessage(false)
    }
  }

  const startCreate = () => {
    setEditMode('create')
    setEditingName(null)
    setForm({ ...emptyForm })
  }

  const startEdit = (cmd: CommandInfo) => {
    setEditMode('edit')
    setEditingName(cmd.name)
    setForm({
      name: cmd.name,
      description: cmd.description,
      response: cmd.response,
      mod_only: cmd.mod_only,
      cooldown: cmd.cooldown,
      enabled: cmd.enabled,
    })
  }

  const cancelEdit = () => {
    setEditMode(null)
    setEditingName(null)
    setForm({ ...emptyForm })
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      if (editMode === 'create') {
        await twitchAPI.createCommand(form)
        addLog({ command: `!${form.name}`, response: 'Command erstellt', success: true, source: 'command' })
      } else if (editMode === 'edit' && editingName) {
        await twitchAPI.updateCommand(editingName, form)
        addLog({ command: `!${editingName}`, response: 'Command aktualisiert', success: true, source: 'command' })
      }
      cancelEdit()
      loadCommands()
    } catch (err) {
      addLog({
        command: `!${form.name}`,
        response: err instanceof Error ? err.message : 'Fehler',
        success: false,
        source: 'command',
      })
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (name: string) => {
    if (!confirm(`Command "!${name}" wirklich löschen?`)) return
    setDeleting(name)
    try {
      await twitchAPI.deleteCommand(name)
      addLog({ command: `!${name}`, response: 'Command gelöscht', success: true, source: 'command' })
      if (editingName === name) cancelEdit()
      loadCommands()
    } catch (err) {
      addLog({
        command: `!${name}`,
        response: err instanceof Error ? err.message : 'Fehler beim Löschen',
        success: false,
        source: 'command',
      })
    } finally {
      setDeleting(null)
    }
  }

  const handleToggle = async (cmd: CommandInfo) => {
    try {
      await twitchAPI.updateCommand(cmd.name, {
        name: cmd.name,
        description: cmd.description,
        response: cmd.response,
        mod_only: cmd.mod_only,
        cooldown: cmd.cooldown,
        enabled: !cmd.enabled,
      })
      loadCommands()
    } catch {
      // Silently fail
    }
  }

  const builtinCommands = commands.filter(c => c.builtin)
  const customCommands = commands.filter(c => !c.builtin)

  return (
    <PageContainer title="Chat Commands" subtitle="Bot-Commands verwalten und Chat-Nachrichten senden">
      <div className="grid lg:grid-cols-2 gap-3">
        {/* Left Column - Commands + Send Message */}
        <div className="space-y-3">
          {/* Command List */}
          <GlowCard glowColor="primary">
            <SectionHeader title="Commands" icon={List} iconColor="text-primary">
              <Button size="sm" variant="ghost" onClick={startCreate} className="text-xs hover:bg-primary/10 hover:text-primary">
                <Plus className="w-3.5 h-3.5 mr-1" />
                Neuer Command
              </Button>
            </SectionHeader>

            {loading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
              </div>
            ) : commands.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-8 text-center space-y-2">
                <Terminal className="w-10 h-10 text-muted-foreground/30" />
                <p className="text-sm text-muted-foreground">Keine Commands registriert</p>
              </div>
            ) : (
              <div className="space-y-1.5">
                {/* Built-in Commands */}
                {builtinCommands.length > 0 && (
                  <>
                    <p className="text-xs text-muted-foreground/60 uppercase tracking-wider px-1">Built-in</p>
                    {builtinCommands.map((cmd) => (
                      <CommandRow
                        key={cmd.name}
                        cmd={cmd}
                        onExecute={executeCommand}
                        executing={executingCmd}
                      />
                    ))}
                  </>
                )}

                {/* Custom Commands */}
                {customCommands.length > 0 && (
                  <>
                    <p className="text-xs text-muted-foreground/60 uppercase tracking-wider px-1 mt-3">Custom</p>
                    {customCommands.map((cmd) => (
                      <CommandRow
                        key={cmd.name}
                        cmd={cmd}
                        onExecute={executeCommand}
                        onEdit={startEdit}
                        onDelete={handleDelete}
                        onToggle={handleToggle}
                        executing={executingCmd}
                        deleting={deleting}
                      />
                    ))}
                  </>
                )}
              </div>
            )}
          </GlowCard>

          {/* Send Chat Message */}
          <GlowCard glowColor="secondary">
            <SectionHeader title="Nachricht senden" icon={MessageSquare} iconColor="text-secondary" />
            <form onSubmit={sendMessage} className="flex gap-2">
              <Input
                value={chatMessage}
                onChange={(e) => setChatMessage(e.target.value)}
                placeholder="Nachricht in den Chat senden..."
                className="flex-1 bg-muted/30 border-border/50"
                disabled={sendingMessage}
              />
              <Button
                type="submit"
                size="sm"
                disabled={sendingMessage || !chatMessage.trim()}
                className="shrink-0"
              >
                {sendingMessage ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <Send className="w-4 h-4" />
                )}
              </Button>
            </form>
          </GlowCard>
        </div>

        {/* Right Column - Form + Log */}
        <div className="space-y-3">
          {/* Create/Edit Form */}
          {editMode && (
            <GlowCard glowColor="warning">
              <SectionHeader
                title={editMode === 'create' ? 'Neuer Command' : `Command bearbeiten: !${editingName}`}
                icon={editMode === 'create' ? Plus : Pencil}
                iconColor="text-warning"
              >
                <Button size="sm" variant="ghost" onClick={cancelEdit} className="text-xs text-muted-foreground hover:text-foreground">
                  <X className="w-3.5 h-3.5" />
                </Button>
              </SectionHeader>

              <form onSubmit={handleSubmit} className="space-y-3">
                {editMode === 'create' && (
                  <div className="space-y-1.5">
                    <Label className="text-xs">Name</Label>
                    <Input
                      value={form.name}
                      onChange={(e) => setForm(f => ({ ...f, name: e.target.value.toLowerCase().replace(/\s/g, '') }))}
                      placeholder="z.B. discord"
                      className="bg-muted/30 border-border/50"
                      required
                    />
                  </div>
                )}

                <div className="space-y-1.5">
                  <Label className="text-xs">Beschreibung</Label>
                  <Input
                    value={form.description}
                    onChange={(e) => setForm(f => ({ ...f, description: e.target.value }))}
                    placeholder="Was macht der Command?"
                    className="bg-muted/30 border-border/50"
                  />
                </div>

                <div className="space-y-1.5">
                  <Label className="text-xs">Antwort (Template)</Label>
                  <Input
                    value={form.response}
                    onChange={(e) => setForm(f => ({ ...f, response: e.target.value }))}
                    placeholder="Hallo {user}! Willkommen auf {channel}"
                    className="bg-muted/30 border-border/50"
                    required
                  />
                  <p className="text-xs text-muted-foreground/60">
                    Variablen: <code className="text-primary/80">{'{user}'}</code> <code className="text-primary/80">{'{channel}'}</code> <code className="text-primary/80">{'{args}'}</code> <code className="text-primary/80">{'{count}'}</code>
                  </p>
                </div>

                <div className="space-y-1.5">
                  <Label className="text-xs">Cooldown (Sekunden)</Label>
                  <Input
                    type="number"
                    min={0}
                    value={form.cooldown}
                    onChange={(e) => setForm(f => ({ ...f, cooldown: Number(e.target.value) }))}
                    className="bg-muted/30 border-border/50 w-24"
                  />
                </div>

                <div className="flex items-center gap-4">
                  <div className="flex items-center gap-2">
                    <Switch
                      checked={form.mod_only}
                      onCheckedChange={(v) => setForm(f => ({ ...f, mod_only: v }))}
                    />
                    <Label className="text-xs">Nur Mods</Label>
                  </div>
                  <div className="flex items-center gap-2">
                    <Switch
                      checked={form.enabled}
                      onCheckedChange={(v) => setForm(f => ({ ...f, enabled: v }))}
                    />
                    <Label className="text-xs">Aktiviert</Label>
                  </div>
                </div>

                <Button type="submit" size="sm" disabled={saving || !form.name || !form.response} className="w-full">
                  {saving ? (
                    <Loader2 className="w-4 h-4 animate-spin mr-1" />
                  ) : (
                    <Save className="w-4 h-4 mr-1" />
                  )}
                  {editMode === 'create' ? 'Erstellen' : 'Speichern'}
                </Button>
              </form>
            </GlowCard>
          )}

          {/* Activity Log */}
          <GlowCard glowColor="success" className="flex flex-col h-[500px]">
            <SectionHeader title="Aktivitäts-Log" icon={Clock} iconColor="text-success">
              {log.length > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setLog([])}
                  className="text-xs text-muted-foreground hover:text-foreground"
                >
                  Leeren
                </Button>
              )}
            </SectionHeader>

            <div className="flex-1 overflow-y-auto space-y-1.5 -mx-1 px-1">
              {log.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full text-center space-y-2">
                  <Terminal className="w-10 h-10 text-muted-foreground/30" />
                  <p className="text-sm text-muted-foreground">Noch keine Aktivität</p>
                  <p className="text-xs text-muted-foreground/60">
                    Führe einen Command aus oder sende eine Nachricht
                  </p>
                </div>
              ) : (
                log.map((entry) => (
                  <div
                    key={entry.id}
                    className={cn(
                      'p-2.5 rounded-md border transition-colors',
                      entry.success
                        ? 'bg-success/5 border-success/20'
                        : 'bg-destructive/5 border-destructive/20'
                    )}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <div className="flex items-center gap-1.5 min-w-0">
                        {entry.source === 'command' ? (
                          <Terminal className="w-3.5 h-3.5 text-primary shrink-0" />
                        ) : (
                          <Send className="w-3.5 h-3.5 text-secondary shrink-0" />
                        )}
                        <span className="font-mono text-xs font-semibold truncate">
                          {entry.command}
                        </span>
                      </div>
                      <span className="text-xs text-muted-foreground shrink-0">
                        {entry.timestamp}
                      </span>
                    </div>
                    <p className={cn(
                      'text-xs mt-1 pl-5',
                      entry.success ? 'text-muted-foreground' : 'text-destructive'
                    )}>
                      {entry.response}
                    </p>
                  </div>
                ))
              )}
              <div ref={logEndRef} />
            </div>
          </GlowCard>
        </div>
      </div>
    </PageContainer>
  )
}

// CommandRow renders a single command in the list
function CommandRow({
  cmd,
  onExecute,
  onEdit,
  onDelete,
  onToggle,
  executing,
  deleting,
}: {
  cmd: CommandInfo
  onExecute: (name: string) => void
  onEdit?: (cmd: CommandInfo) => void
  onDelete?: (name: string) => void
  onToggle?: (cmd: CommandInfo) => void
  executing: string | null
  deleting?: string | null
}) {
  return (
    <div
      className={cn(
        'flex items-center justify-between p-2.5 rounded-md bg-muted/30 border border-border/30 hover:bg-muted/50 transition-colors',
        !cmd.builtin && !cmd.enabled && 'opacity-50'
      )}
    >
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="font-mono text-sm font-semibold text-primary">
            !{cmd.name}
          </span>
          {cmd.mod_only && (
            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 text-xs font-semibold bg-amber-500/20 text-amber-400 rounded">
              <Shield className="w-3 h-3" />
              MOD
            </span>
          )}
          {cmd.builtin && (
            <span className="inline-flex items-center px-1.5 py-0.5 text-xs bg-muted/50 text-muted-foreground rounded">
              built-in
            </span>
          )}
          {!cmd.builtin && cmd.count > 0 && (
            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 text-xs text-muted-foreground/60 rounded">
              <Hash className="w-3 h-3" />
              {cmd.count}
            </span>
          )}
        </div>
        <p className="text-xs text-muted-foreground mt-0.5 truncate">{cmd.description}</p>
      </div>
      <div className="flex items-center gap-1 shrink-0 ml-2">
        {!cmd.builtin && onToggle && (
          <Switch
            checked={cmd.enabled}
            onCheckedChange={() => onToggle(cmd)}
            className="scale-75"
          />
        )}
        {!cmd.builtin && onEdit && (
          <Button size="sm" variant="ghost" onClick={() => onEdit(cmd)} className="h-7 w-7 p-0 hover:bg-primary/10 hover:text-primary">
            <Pencil className="w-3.5 h-3.5" />
          </Button>
        )}
        {!cmd.builtin && onDelete && (
          <Button
            size="sm"
            variant="ghost"
            onClick={() => onDelete(cmd.name)}
            disabled={deleting === cmd.name}
            className="h-7 w-7 p-0 hover:bg-destructive/10 hover:text-destructive"
          >
            {deleting === cmd.name ? (
              <Loader2 className="w-3.5 h-3.5 animate-spin" />
            ) : (
              <Trash2 className="w-3.5 h-3.5" />
            )}
          </Button>
        )}
        <Button
          size="sm"
          variant="ghost"
          className="h-7 w-7 p-0 hover:bg-primary/10 hover:text-primary"
          onClick={() => onExecute(cmd.name)}
          disabled={executing !== null}
        >
          {executing === cmd.name ? (
            <Loader2 className="w-3.5 h-3.5 animate-spin" />
          ) : (
            <Play className="w-3.5 h-3.5" />
          )}
        </Button>
      </div>
    </div>
  )
}
