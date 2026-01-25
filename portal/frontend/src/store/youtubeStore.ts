import { create } from 'zustand'
import type { YouTubeStatus, StreamStats, ChatMessage } from '@/types/youtube'

const MAX_CHAT_MESSAGES = 100

interface YouTubeStore {
  status: YouTubeStatus | null
  stats: StreamStats | null
  chatMessages: ChatMessage[]

  // Actions
  setStatus: (status: YouTubeStatus) => void
  setStats: (stats: StreamStats) => void
  addChatMessage: (message: ChatMessage) => void
  clearChat: () => void
}

export const useYouTubeStore = create<YouTubeStore>((set) => ({
  status: null,
  stats: null,
  chatMessages: [],

  setStatus: (status) => set({ status }),

  setStats: (stats) => set({ stats }),

  addChatMessage: (message) =>
    set((state) => ({
      chatMessages: [...state.chatMessages, message].slice(-MAX_CHAT_MESSAGES),
    })),

  clearChat: () => set({ chatMessages: [] }),
}))
