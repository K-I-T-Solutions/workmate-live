import { authFetch } from '@/lib/api'
import type { PortalConfig, ConnectedAgentsResponse } from '@/types/config'

const API_BASE = '/api/config'

export const configAPI = {
  async getConfig(): Promise<PortalConfig> {
    const res = await authFetch(API_BASE)

    if (!res.ok) {
      throw new Error('Failed to load config')
    }

    return res.json()
  },

  /** Agents, die aktuell über den Link verbunden sind. */
  async getConnectedAgents(): Promise<ConnectedAgentsResponse> {
    const res = await authFetch('/api/agents')

    if (!res.ok) {
      throw new Error('Failed to load connected agents')
    }

    return res.json()
  },

  async updateConfig(section: string, data: object): Promise<void> {
    const res = await authFetch(API_BASE, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ section, data }),
    })

    if (!res.ok) {
      const error = await res.text()
      throw new Error(error || 'Failed to update config')
    }
  },
}
