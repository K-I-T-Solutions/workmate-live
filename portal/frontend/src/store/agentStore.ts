import { create } from 'zustand'
import type { AgentStatus } from '@/types/agent'

interface AgentStore {
  status: AgentStatus | null
  connected: boolean
  setStatus: (status: AgentStatus) => void
  setConnected: (connected: boolean) => void
}

export const useAgentStore = create<AgentStore>((set) => ({
  status: null,
  connected: false,
  setStatus: (status) => set({ status, connected: true }),
  setConnected: (connected) => set({ connected }),
}))
