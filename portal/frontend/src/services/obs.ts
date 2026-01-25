import type { OBSStatus, Scene, Source } from '@/types/obs'

const API_BASE = '/api/obs'

export const obsAPI = {
  // Get OBS status
  async getStatus(): Promise<OBSStatus> {
    const response = await fetch(`${API_BASE}/status`)
    if (!response.ok) throw new Error('Failed to fetch OBS status')
    return response.json()
  },

  // Get all scenes
  async getScenes(): Promise<Scene[]> {
    const response = await fetch(`${API_BASE}/scenes`)
    if (!response.ok) throw new Error('Failed to fetch scenes')
    const data = await response.json()
    return data.scenes
  },

  // Switch to a scene
  async switchScene(sceneName: string): Promise<void> {
    const response = await fetch(`${API_BASE}/scenes/switch`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ scene_name: sceneName }),
    })
    if (!response.ok) throw new Error('Failed to switch scene')
  },

  // Get sources for a scene
  async getSources(sceneName?: string): Promise<Source[]> {
    const url = sceneName
      ? `${API_BASE}/sources?scene=${encodeURIComponent(sceneName)}`
      : `${API_BASE}/sources`
    const response = await fetch(url)
    if (!response.ok) throw new Error('Failed to fetch sources')
    const data = await response.json()
    return data.sources
  },

  // Toggle source visibility
  async toggleSource(sceneName: string, sourceName: string, visible: boolean): Promise<void> {
    const response = await fetch(`${API_BASE}/sources/toggle`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ scene_name: sceneName, source_name: sourceName, visible }),
    })
    if (!response.ok) throw new Error('Failed to toggle source')
  },

  // Streaming controls
  async startStreaming(): Promise<void> {
    const response = await fetch(`${API_BASE}/streaming/start`, { method: 'POST' })
    if (!response.ok) throw new Error('Failed to start streaming')
  },

  async stopStreaming(): Promise<void> {
    const response = await fetch(`${API_BASE}/streaming/stop`, { method: 'POST' })
    if (!response.ok) throw new Error('Failed to stop streaming')
  },

  // Recording controls
  async startRecording(): Promise<void> {
    const response = await fetch(`${API_BASE}/recording/start`, { method: 'POST' })
    if (!response.ok) throw new Error('Failed to start recording')
  },

  async stopRecording(): Promise<void> {
    const response = await fetch(`${API_BASE}/recording/stop`, { method: 'POST' })
    if (!response.ok) throw new Error('Failed to stop recording')
  },

  async pauseRecording(): Promise<void> {
    const response = await fetch(`${API_BASE}/recording/pause`, { method: 'POST' })
    if (!response.ok) throw new Error('Failed to pause recording')
  },

  async resumeRecording(): Promise<void> {
    const response = await fetch(`${API_BASE}/recording/resume`, { method: 'POST' })
    if (!response.ok) throw new Error('Failed to resume recording')
  },
}
