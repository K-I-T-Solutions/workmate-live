import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types/auth'

interface AuthStore {
  token: string | null
  user: User | null
  isAuthenticated: boolean

  // Actions
  setAuth: (token: string, user: User) => void
  clearAuth: () => void
  setUser: (user: User) => void
}

export const useAuthStore = create<AuthStore>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      isAuthenticated: false,

      setAuth: (token, user) => set({ token, user, isAuthenticated: true }),

      clearAuth: () => set({ token: null, user: null, isAuthenticated: false }),

      setUser: (user) => set({ user }),
    }),
    {
      name: 'workmate-auth', // localStorage key
    }
  )
)
