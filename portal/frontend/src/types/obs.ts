export interface Scene {
  name: string
  active: boolean
  index: number
  scene_uuid?: string
}

export interface Source {
  name: string
  type: string
  visible: boolean
  muted?: boolean
  volume?: number
  input_uuid?: string
}

export interface StreamStatus {
  active: boolean
  reconnecting: boolean
  duration: number
  bytes: number
  frames?: number
  dropped_frames?: number
}

export interface RecordingStatus {
  active: boolean
  paused: boolean
  duration: number
  bytes: number
  path?: string
}

export interface OBSStatus {
  connected: boolean
  version?: string
  current_scene?: string
  streaming?: StreamStatus
  recording?: RecordingStatus
}

export interface OBSEvent {
  type: string
  scene_name?: string
  source_name?: string
  active?: boolean
  visible?: boolean
}
