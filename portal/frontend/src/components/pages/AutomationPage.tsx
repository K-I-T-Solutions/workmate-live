import { useEffect, useState } from 'react'
import { PageContainer } from '@/components/shared/PageContainer'
import { GlowCard } from '@/components/shared/GlowCard'
import { SectionHeader } from '@/components/shared/SectionHeader'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { automationAPI } from '@/services/automation'
import { useAutomationStore } from '@/store/automationStore'
import {
  TRIGGER_TYPES, ACTION_TYPES,
  type Rule, type Action,
} from '@/types/automation'
import {
  Zap, Plus, Trash2, Save, X, Pencil, Play,
  Loader2, CheckCircle, XCircle, ChevronDown, ChevronUp,
} from 'lucide-react'
import { cn } from '@/lib/utils'

const emptyRule = (): Rule => ({
  name: '',
  enabled: true,
  trigger: { type: TRIGGER_TYPES[0].value },
  actions: [],
})

const emptyAction = (): Action => ({
  type: ACTION_TYPES[0].value,
  params: {},
})

function ActionEditor({
  action,
  index,
  onChange,
  onRemove,
  onMoveUp,
  onMoveDown,
  isFirst,
  isLast,
}: {
  action: Action
  index: number
  onChange: (action: Action) => void
  onRemove: () => void
  onMoveUp: () => void
  onMoveDown: () => void
  isFirst: boolean
  isLast: boolean
}) {
  const actionDef = ACTION_TYPES.find(a => a.value === action.type)

  const handleTypeChange = (type: string) => {
    onChange({ type, params: {} })
  }

  const handleParamChange = (key: string, value: unknown) => {
    onChange({ ...action, params: { ...action.params, [key]: value } })
  }

  return (
    <div className="border border-border/50 rounded-md p-3 space-y-2 bg-card/30">
      <div className="flex items-center gap-2">
        <span className="text-xs text-muted-foreground w-5 shrink-0">{index + 1}.</span>
        <select
          value={action.type}
          onChange={e => handleTypeChange(e.target.value)}
          className="flex-1 h-8 rounded-md border border-border/50 bg-background px-2 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
        >
          {ACTION_TYPES.map(a => (
            <option key={a.value} value={a.value}>{a.label}</option>
          ))}
        </select>
        <button onClick={onMoveUp} disabled={isFirst} className="p-1 text-muted-foreground hover:text-foreground disabled:opacity-30">
          <ChevronUp className="h-3 w-3" />
        </button>
        <button onClick={onMoveDown} disabled={isLast} className="p-1 text-muted-foreground hover:text-foreground disabled:opacity-30">
          <ChevronDown className="h-3 w-3" />
        </button>
        <button onClick={onRemove} className="p-1 text-muted-foreground hover:text-destructive">
          <Trash2 className="h-3 w-3" />
        </button>
      </div>

      {actionDef?.params.map(param => (
        <div key={param.key} className="flex items-center gap-2 pl-5">
          <Label className="text-xs text-muted-foreground w-28 shrink-0">{param.label}</Label>
          {param.type === 'boolean' ? (
            <Switch
              checked={!!action.params[param.key]}
              onCheckedChange={v => handleParamChange(param.key, v)}
            />
          ) : (
            <Input
              type={param.type === 'number' ? 'number' : 'text'}
              value={String(action.params[param.key] ?? '')}
              placeholder={param.placeholder}
              onChange={e => {
                const val = param.type === 'number' ? Number(e.target.value) : e.target.value
                handleParamChange(param.key, val)
              }}
              className="h-7 text-xs"
            />
          )}
        </div>
      ))}
    </div>
  )
}

export function AutomationPage() {
  const { firedEvents, clearFiredEvents } = useAutomationStore()
  const [rules, setRules] = useState<Rule[]>([])
  const [loading, setLoading] = useState(true)
  const [editMode, setEditMode] = useState<'create' | 'edit' | null>(null)
  const [editingName, setEditingName] = useState<string | null>(null)
  const [form, setForm] = useState<Rule>(emptyRule())
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState<string | null>(null)
  const [deleting, setDeleting] = useState<string | null>(null)
  const [toggling, setToggling] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const loadRules = () => {
    automationAPI.listRules()
      .then(setRules)
      .catch(() => {})
      .finally(() => setLoading(false))
  }

  useEffect(() => { loadRules() }, [])

  const startCreate = () => {
    setForm(emptyRule())
    setEditingName(null)
    setEditMode('create')
    setError(null)
  }

  const startEdit = (rule: Rule) => {
    setForm({ ...rule, actions: rule.actions.map(a => ({ ...a, params: { ...a.params } })) })
    setEditingName(rule.name)
    setEditMode('edit')
    setError(null)
  }

  const cancelEdit = () => {
    setEditMode(null)
    setEditingName(null)
    setError(null)
  }

  const handleSubmit = async () => {
    if (!form.name.trim()) { setError('Name ist erforderlich'); return }
    if (!form.trigger.type) { setError('Trigger-Typ ist erforderlich'); return }

    setSaving(true)
    setError(null)
    try {
      if (editMode === 'create') {
        await automationAPI.createRule(form)
      } else if (editingName) {
        await automationAPI.updateRule(editingName, form)
      }
      cancelEdit()
      loadRules()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Fehler beim Speichern')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (name: string) => {
    if (!confirm(`Rule "${name}" wirklich löschen?`)) return
    setDeleting(name)
    try {
      await automationAPI.deleteRule(name)
      loadRules()
    } catch {
    } finally {
      setDeleting(null)
    }
  }

  const handleTest = async (name: string) => {
    setTesting(name)
    try {
      await automationAPI.testRule(name)
    } catch {
    } finally {
      setTesting(null)
    }
  }

  const handleToggle = async (rule: Rule) => {
    setToggling(rule.name)
    try {
      await automationAPI.updateRule(rule.name, { ...rule, enabled: !rule.enabled })
      loadRules()
    } catch {
    } finally {
      setToggling(null)
    }
  }

  const addAction = () => {
    setForm(f => ({ ...f, actions: [...f.actions, emptyAction()] }))
  }

  const updateAction = (index: number, action: Action) => {
    setForm(f => {
      const actions = [...f.actions]
      actions[index] = action
      return { ...f, actions }
    })
  }

  const removeAction = (index: number) => {
    setForm(f => ({ ...f, actions: f.actions.filter((_, i) => i !== index) }))
  }

  const moveAction = (index: number, dir: -1 | 1) => {
    setForm(f => {
      const actions = [...f.actions]
      const [item] = actions.splice(index, 1)
      actions.splice(index + dir, 0, item)
      return { ...f, actions }
    })
  }

  const triggerLabel = (type: string) =>
    TRIGGER_TYPES.find(t => t.value === type)?.label ?? type

  return (
    <PageContainer title="Automations" subtitle="Wenn Trigger X eintritt → führe Actions Y, Z aus">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">

        {/* Left: Rule list */}
        <div className="space-y-4">
          <GlowCard>
            <div className="flex items-center justify-between mb-4">
              <SectionHeader icon={Zap} title="Regeln" />
              <Button size="sm" onClick={startCreate} disabled={editMode !== null}>
                <Plus className="h-4 w-4 mr-1" /> Neue Regel
              </Button>
            </div>

            {loading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
              </div>
            ) : rules.length === 0 ? (
              <p className="text-sm text-muted-foreground text-center py-8">
                Noch keine Regeln. Erstelle deine erste Automation.
              </p>
            ) : (
              <div className="space-y-2">
                {rules.map(rule => (
                  <div
                    key={rule.name}
                    className={cn(
                      'flex items-center gap-3 p-3 rounded-lg border border-border/50 transition-colors',
                      editingName === rule.name ? 'bg-primary/10 border-primary/40' : 'hover:bg-accent/30'
                    )}
                  >
                    <Switch
                      checked={rule.enabled}
                      onCheckedChange={() => handleToggle(rule)}
                      disabled={toggling === rule.name}
                    />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">{rule.name}</p>
                      <Badge variant="outline" className="text-xs mt-0.5">
                        {triggerLabel(rule.trigger.type)}
                      </Badge>
                      <span className="text-xs text-muted-foreground ml-2">
                        {rule.actions.length} Action{rule.actions.length !== 1 ? 's' : ''}
                      </span>
                    </div>
                    <div className="flex items-center gap-1 shrink-0">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleTest(rule.name)}
                        disabled={testing === rule.name}
                        title="Manuell auslösen"
                      >
                        {testing === rule.name
                          ? <Loader2 className="h-3.5 w-3.5 animate-spin" />
                          : <Play className="h-3.5 w-3.5" />}
                      </Button>
                      <Button size="sm" variant="ghost" onClick={() => startEdit(rule)}>
                        <Pencil className="h-3.5 w-3.5" />
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleDelete(rule.name)}
                        disabled={deleting === rule.name}
                        className="text-muted-foreground hover:text-destructive"
                      >
                        {deleting === rule.name
                          ? <Loader2 className="h-3.5 w-3.5 animate-spin" />
                          : <Trash2 className="h-3.5 w-3.5" />}
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </GlowCard>

          {/* Activity Log */}
          <GlowCard>
            <div className="flex items-center justify-between mb-3">
              <SectionHeader icon={Zap} title="Letzte Auslösungen" />
              {firedEvents.length > 0 && (
                <Button size="sm" variant="ghost" onClick={clearFiredEvents}>
                  <X className="h-3 w-3 mr-1" /> Leeren
                </Button>
              )}
            </div>
            <div className="space-y-1.5 max-h-52 overflow-y-auto">
              {firedEvents.length === 0 ? (
                <p className="text-xs text-muted-foreground text-center py-4">
                  Noch keine Auslösungen in dieser Sitzung.
                </p>
              ) : (
                [...firedEvents].reverse().map((e, i) => (
                  <div key={i} className="flex items-center gap-2 text-xs">
                    <CheckCircle className="h-3 w-3 text-green-500 shrink-0" />
                    <span className="font-medium truncate">{e.rule_name}</span>
                    <span className="text-muted-foreground shrink-0">
                      {new Date(e.timestamp).toLocaleTimeString('de-DE')}
                    </span>
                  </div>
                ))
              )}
            </div>
          </GlowCard>
        </div>

        {/* Right: Rule editor */}
        {editMode && (
          <div>
            <GlowCard>
              <div className="flex items-center justify-between mb-4">
                <SectionHeader
                  icon={editMode === 'create' ? Plus : Pencil}
                  title={editMode === 'create' ? 'Neue Regel' : `"${editingName}" bearbeiten`}
                />
                <Button size="sm" variant="ghost" onClick={cancelEdit}>
                  <X className="h-4 w-4" />
                </Button>
              </div>

              <div className="space-y-4">
                {/* Name */}
                <div className="space-y-1.5">
                  <Label className="text-sm">Name</Label>
                  <Input
                    value={form.name}
                    onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                    placeholder="Follow Begrüßung"
                    disabled={editMode === 'edit'}
                  />
                </div>

                {/* Trigger */}
                <div className="space-y-1.5">
                  <Label className="text-sm">Trigger</Label>
                  <select
                    value={form.trigger.type}
                    onChange={e => setForm(f => ({ ...f, trigger: { type: e.target.value } }))}
                    className="w-full h-9 rounded-md border border-border/50 bg-background px-3 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                  >
                    {TRIGGER_TYPES.map(t => (
                      <option key={t.value} value={t.value}>{t.label}</option>
                    ))}
                  </select>
                </div>

                {/* Enabled toggle */}
                <div className="flex items-center gap-3">
                  <Switch
                    checked={form.enabled}
                    onCheckedChange={v => setForm(f => ({ ...f, enabled: v }))}
                  />
                  <Label className="text-sm">Aktiv</Label>
                </div>

                {/* Actions */}
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label className="text-sm">Actions</Label>
                    <Button size="sm" variant="outline" onClick={addAction}>
                      <Plus className="h-3 w-3 mr-1" /> Action hinzufügen
                    </Button>
                  </div>

                  {form.actions.length === 0 && (
                    <p className="text-xs text-muted-foreground text-center py-3 border border-dashed border-border/50 rounded-md">
                      Noch keine Actions. Füge mindestens eine hinzu.
                    </p>
                  )}

                  {form.actions.map((action, i) => (
                    <ActionEditor
                      key={i}
                      action={action}
                      index={i}
                      onChange={a => updateAction(i, a)}
                      onRemove={() => removeAction(i)}
                      onMoveUp={() => moveAction(i, -1)}
                      onMoveDown={() => moveAction(i, 1)}
                      isFirst={i === 0}
                      isLast={i === form.actions.length - 1}
                    />
                  ))}
                </div>

                {/* Template-Variablen Hinweis */}
                <div className="text-xs text-muted-foreground bg-muted/20 rounded-md p-2.5 space-y-0.5">
                  <p className="font-medium mb-1">Template-Variablen:</p>
                  <p><code className="text-primary">{'{user}'}</code> — Twitch Username (Follow, Sub)</p>
                  <p><code className="text-primary">{'{raider}'}</code> / <code className="text-primary">{'{viewers}'}</code> — Raid</p>
                  <p><code className="text-primary">{'{tier}'}</code> — Sub Tier · <code className="text-primary">{'{scene}'}</code> — Szene</p>
                </div>

                {error && (
                  <div className="flex items-center gap-2 text-sm text-destructive">
                    <XCircle className="h-4 w-4 shrink-0" />
                    {error}
                  </div>
                )}

                <div className="flex gap-2 pt-1">
                  <Button onClick={handleSubmit} disabled={saving} className="flex-1">
                    {saving ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <Save className="h-4 w-4 mr-2" />}
                    Speichern
                  </Button>
                  <Button variant="outline" onClick={cancelEdit}>
                    Abbrechen
                  </Button>
                </div>
              </div>
            </GlowCard>
          </div>
        )}
      </div>
    </PageContainer>
  )
}
