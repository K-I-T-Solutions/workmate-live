import type { TwitchStatus, StreamStats, UpdateStreamRequest, CommandInfo, CommandResult, CreateCommandRequest } from '@/types/twitch'
import { authFetch } from '@/lib/api'

const API_BASE = '/api/twitch'

export const twitchAPI = {
  async getStatus(): Promise<TwitchStatus> {
    const response = await authFetch(`${API_BASE}/status`)
    if (!response.ok) throw new Error('Failed to fetch Twitch status')
    return response.json()
  },

  async getStats(): Promise<StreamStats> {
    const response = await authFetch(`${API_BASE}/stats`)
    if (!response.ok) throw new Error('Failed to fetch stream stats')
    return response.json()
  },

  async updateStream(request: UpdateStreamRequest): Promise<void> {
    const response = await authFetch(`${API_BASE}/stream`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    })
    if (!response.ok) throw new Error('Failed to update stream')
  },

  async listCommands(): Promise<CommandInfo[]> {
    const response = await authFetch(`${API_BASE}/commands`)
    if (!response.ok) throw new Error('Failed to fetch commands')
    return response.json()
  },

  async executeCommand(command: string): Promise<CommandResult> {
    const response = await authFetch(`${API_BASE}/commands/exec`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command }),
    })
    if (!response.ok) throw new Error('Failed to execute command')
    return response.json()
  },

  async sendMessage(message: string): Promise<void> {
    const response = await authFetch(`${API_BASE}/chat/send`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message }),
    })
    if (!response.ok) throw new Error('Failed to send message')
  },

  async createCommand(req: CreateCommandRequest): Promise<void> {
    const response = await authFetch(`${API_BASE}/commands`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    })
    if (!response.ok) {
      const text = await response.text()
      throw new Error(text || 'Failed to create command')
    }
  },

  async updateCommand(name: string, req: CreateCommandRequest): Promise<void> {
    const response = await authFetch(`${API_BASE}/commands/${encodeURIComponent(name)}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    })
    if (!response.ok) {
      const text = await response.text()
      throw new Error(text || 'Failed to update command')
    }
  },

  async deleteCommand(name: string): Promise<void> {
    const response = await authFetch(`${API_BASE}/commands/${encodeURIComponent(name)}`, {
      method: 'DELETE',
    })
    if (!response.ok) {
      const text = await response.text()
      throw new Error(text || 'Failed to delete command')
    }
  },
}
