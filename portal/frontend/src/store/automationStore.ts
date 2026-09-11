import { create } from 'zustand'
import type { Rule, FiredEvent } from '@/types/automation'

const MAX_FIRED_EVENTS = 50

interface AutomationStore {
  rules: Rule[]
  firedEvents: FiredEvent[]
  setRules: (rules: Rule[]) => void
  addFiredEvent: (event: FiredEvent) => void
  clearFiredEvents: () => void
}

export const useAutomationStore = create<AutomationStore>((set) => ({
  rules: [],
  firedEvents: [],

  setRules: (rules) => set({ rules }),

  addFiredEvent: (event) =>
    set((state) => ({
      firedEvents: [...state.firedEvents, event].slice(-MAX_FIRED_EVENTS),
    })),

  clearFiredEvents: () => set({ firedEvents: [] }),
}))
