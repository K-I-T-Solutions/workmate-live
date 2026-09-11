export interface Trigger {
  type: string
}

export interface Action {
  type: string
  params: Record<string, unknown>
}

export interface Rule {
  name: string
  enabled: boolean
  trigger: Trigger
  actions: Action[]
}

export interface FiredEvent {
  rule_name: string
  trigger_type: string
  timestamp: string
}

// Supported trigger types
export const TRIGGER_TYPES = [
  { value: 'twitch:follow', label: 'Twitch: Follow' },
  { value: 'twitch:subscribe', label: 'Twitch: Abonnement' },
  { value: 'twitch:raid', label: 'Twitch: Raid' },
  { value: 'obs:scene_changed', label: 'OBS: Szene gewechselt' },
  { value: 'obs:stream_started', label: 'OBS: Stream gestartet' },
  { value: 'obs:stream_stopped', label: 'OBS: Stream gestoppt' },
] as const

// Supported action types with their parameter definitions
export const ACTION_TYPES = [
  {
    value: 'twitch:chat_message',
    label: 'Twitch Chat-Nachricht',
    params: [{ key: 'message', label: 'Nachricht', type: 'text', placeholder: 'Danke für den Follow, {user}!' }],
  },
  {
    value: 'obs:scene_switch',
    label: 'OBS Szene wechseln',
    params: [{ key: 'scene', label: 'Szene', type: 'text', placeholder: 'Gameplay' }],
  },
  {
    value: 'obs:source_toggle',
    label: 'OBS Quelle togglen',
    params: [
      { key: 'source', label: 'Quelle', type: 'text', placeholder: 'Webcam' },
      { key: 'scene', label: 'Szene (leer = aktuelle)', type: 'text', placeholder: '' },
      { key: 'visible', label: 'Sichtbar', type: 'boolean', placeholder: '' },
    ],
  },
  {
    value: 'delay',
    label: 'Pause (Sekunden)',
    params: [{ key: 'seconds', label: 'Sekunden', type: 'number', placeholder: '3' }],
  },
] as const
