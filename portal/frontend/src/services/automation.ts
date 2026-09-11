import { authFetch } from '@/lib/api'
import type { Rule } from '@/types/automation'

export const automationAPI = {
  async listRules(): Promise<Rule[]> {
    const res = await authFetch('/api/automation/rules')
    if (!res.ok) throw new Error('Failed to load rules')
    return res.json()
  },

  async createRule(rule: Rule): Promise<void> {
    const res = await authFetch('/api/automation/rules', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(rule),
    })
    if (!res.ok) {
      const text = await res.text()
      throw new Error(text || 'Failed to create rule')
    }
  },

  async updateRule(name: string, rule: Rule): Promise<void> {
    const res = await authFetch(`/api/automation/rules/${encodeURIComponent(name)}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(rule),
    })
    if (!res.ok) {
      const text = await res.text()
      throw new Error(text || 'Failed to update rule')
    }
  },

  async deleteRule(name: string): Promise<void> {
    const res = await authFetch(`/api/automation/rules/${encodeURIComponent(name)}`, {
      method: 'DELETE',
    })
    if (!res.ok) throw new Error('Failed to delete rule')
  },

  async testRule(name: string): Promise<void> {
    const res = await authFetch(`/api/automation/rules/${encodeURIComponent(name)}/test`, {
      method: 'POST',
    })
    if (!res.ok) throw new Error('Failed to test rule')
  },
}
