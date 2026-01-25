import type { LoginRequest, LoginResponse, VerifyResponse } from '@/types/auth'

const API_BASE = 'http://localhost:8080'

export const authAPI = {
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const res = await fetch(`${API_BASE}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(credentials),
    })

    if (!res.ok) {
      const error = await res.text()
      throw new Error(error || 'Login failed')
    }

    return res.json()
  },

  async logout(token: string): Promise<void> {
    const res = await fetch(`${API_BASE}/api/auth/logout`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    })

    if (!res.ok) {
      throw new Error('Logout failed')
    }
  },

  async verify(token: string): Promise<VerifyResponse> {
    const res = await fetch(`${API_BASE}/api/auth/verify`, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    })

    if (!res.ok) {
      throw new Error('Token verification failed')
    }

    return res.json()
  },
}
