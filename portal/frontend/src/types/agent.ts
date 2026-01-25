export interface AgentStatus {
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
  render_nodes?: string[]
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
