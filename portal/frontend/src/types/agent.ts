export interface AgentStatus {
  /**
   * Kennung des meldenden Agents. Ohne sie lassen sich mehrere Agents
   * nicht auseinanderhalten — ihre Meldungen überschreiben sich sonst.
   */
  agent_id?: string
  timestamp: string
  hostname: string
  headless: boolean
  video: VideoStatus
  audio: AudioStatus
  obs: OBSStatus
  gpu: GPUStatus
}

export interface VideoStatus {
  device_count: number
  devices: string[]
}

export interface AudioStatus {
  backend: string
  ready: boolean
}

export interface OBSStatus {
  running: boolean
}

export interface GPUStatus {
  present: boolean
  vendors?: string[]
  /** DRM-Render-Nodes — nur unter Linux gesetzt. */
  render_nodes?: string[]
  /** Namen der Grafikadapter — unter Windows und macOS gesetzt. */
  adapters?: string[]
}

export interface Capabilities {
  can_video: boolean
  can_audio: boolean
  can_stream: boolean
}

export interface AgentInfo {
  name: string
  version: string
  commit: string
  build_time: string
  specs: SpecsInfo
}

export interface SpecsInfo {
  os: string
  arch: string
  kernel: string
  cpu: CPUInfo
  memory: MemoryInfo
}

export interface CPUInfo {
  model: string
  cores: number
  threads: number
}

export interface MemoryInfo {
  total_mb: number
}
