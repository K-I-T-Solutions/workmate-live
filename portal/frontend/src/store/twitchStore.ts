import { create } from 'zustand'
import type { TwitchStatus, StreamStats, ChatMessage, TwitchEvent } from '@/types/twitch'

interface TwitchStore {
  status: TwitchStatus | null
  stats: StreamStats | null
  chatMessages: ChatMessage[]
  events: TwitchEvent[]

  setStatus: (status: TwitchStatus) => void
  setStats: (stats: StreamStats) => void
  addChatMessage: (message: ChatMessage) => void
  addEvent: (event: TwitchEvent) => void
  clearChat: () => void
  clearEvents: () => void
}

const MAX_CHAT_MESSAGES = 100
const MAX_EVENTS = 50

export const useTwitchStore = create<TwitchStore>((set) => ({
  status: null,
  stats: null,
  chatMessages: [],
  events: [],

  setStatus: (status) => set({ status }),
  setStats: (stats) => set({ stats }),

  addChatMessage: (message) =>
    set((state) => ({
      chatMessages: [...state.chatMessages, message].slice(-MAX_CHAT_MESSAGES),
    })),

  addEvent: (event) =>
    set((state) => ({
      events: [event, ...state.events].slice(0, MAX_EVENTS),
    })),

  clearChat: () => set({ chatMessages: [] }),
  clearEvents: () => set({ events: [] }),
}))
