import type { TwitchStatus, StreamStats, UpdateStreamRequest } from '@/types/twitch'

const API_BASE = '/api/twitch'

export const twitchAPI = {
  async getStatus(): Promise<TwitchStatus> {
    const response = await fetch(`${API_BASE}/status`)
    if (!response.ok) throw new Error('Failed to fetch Twitch status')
    return response.json()
  },

  async getStats(): Promise<StreamStats> {
    const response = await fetch(`${API_BASE}/stats`)
    if (!response.ok) throw new Error('Failed to fetch stream stats')
    return response.json()
  },

  async updateStream(request: UpdateStreamRequest): Promise<void> {
    const response = await fetch(`${API_BASE}/stream`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    })
    if (!response.ok) throw new Error('Failed to update stream')
  },
}
