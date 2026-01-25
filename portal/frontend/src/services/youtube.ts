import type { YouTubeStatus, StreamStats, UpdateStreamRequest } from '@/types/youtube'

const API_BASE = 'http://localhost:8080'

export const youtubeAPI = {
  async getStatus(): Promise<YouTubeStatus> {
    const res = await fetch(`${API_BASE}/api/youtube/status`)
    if (!res.ok) throw new Error('Failed to fetch YouTube status')
    return res.json()
  },

  async getStats(): Promise<StreamStats> {
    const res = await fetch(`${API_BASE}/api/youtube/stats`)
    if (!res.ok) throw new Error('Failed to fetch YouTube stats')
    return res.json()
  },

  async updateStream(data: UpdateStreamRequest): Promise<void> {
    const res = await fetch(`${API_BASE}/api/youtube/stream`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    })
    if (!res.ok) throw new Error('Failed to update stream')
  },
}
