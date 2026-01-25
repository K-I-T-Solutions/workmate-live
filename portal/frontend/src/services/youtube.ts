import type { YouTubeStatus, StreamStats, UpdateStreamRequest } from '@/types/youtube'
import { authFetch } from '@/lib/api'

const API_BASE = '/api/youtube'

export const youtubeAPI = {
  async getStatus(): Promise<YouTubeStatus> {
    const res = await authFetch(`${API_BASE}/status`)
    if (!res.ok) throw new Error('Failed to fetch YouTube status')
    return res.json()
  },

  async getStats(): Promise<StreamStats> {
    const res = await authFetch(`${API_BASE}/stats`)
    if (!res.ok) throw new Error('Failed to fetch YouTube stats')
    return res.json()
  },

  async updateStream(data: UpdateStreamRequest): Promise<void> {
    const res = await authFetch(`${API_BASE}/stream`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    })
    if (!res.ok) throw new Error('Failed to update stream')
  },
}
