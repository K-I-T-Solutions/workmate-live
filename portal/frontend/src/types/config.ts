/**
 * direct: Das Portal verbindet sich selbst zum OBS-WebSocket.
 * agent:  Die Steuerung läuft über einen Agent, der sich beim Portal meldet.
 */
export type OBSMode = 'direct' | 'agent'

export interface OBSConfig {
  mode: OBSMode
  host: string
  port: number
  password: string
  auto_reconnect: boolean
  reconnect_delay: number
}

export interface TwitchConfig {
  enabled: boolean
  client_id: string
  client_secret: string
  channel: string
  oauth_token: string
}

export interface YouTubeConfig {
  enabled: boolean
  api_key: string
  channel_id: string
  client_id: string
  client_secret: string
}

export interface AgentConfig {
  url: string
  polling_interval: number
  timeout: number
  /** Gemeinsamer Schlüssel für /ws/agent. Leer deaktiviert den Endpunkt. */
  api_key: string
  /** Agent, der OBS steuert. Leer: erster verbundener Agent mit OBS. */
  primary_id: string
  /** Wartezeit auf die Antwort eines Agents (ns). */
  command_timeout: number
}

/** Ein über den Link verbundener Agent. */
export interface ConnectedAgent {
  agent_id: string
  hostname?: string
  version?: string
  commit?: string
  obs_local: boolean
  obs_active: boolean
  connected: boolean
  since: string
}

export interface ConnectedAgentsResponse {
  agents: ConnectedAgent[]
  link_enabled: boolean
}

export interface DefaultUserConfig {
  username: string
  password: string
}

export interface AuthConfig {
  jwt_secret: string
  token_duration: number
  default_user: DefaultUserConfig
}

export interface PortalConfig {
  obs: OBSConfig
  twitch: TwitchConfig
  youtube: YouTubeConfig
  agent: AgentConfig
  auth: AuthConfig
}
