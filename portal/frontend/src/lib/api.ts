import { useAuthStore } from '@/store/authStore'

/**
 * Get authorization headers with JWT token
 */
export function getAuthHeaders(): HeadersInit {
  const token = useAuthStore.getState().token

  if (!token) {
    throw new Error('No authentication token available')
  }

  return {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  }
}

/**
 * Fetch with automatic authentication
 */
export async function authFetch(url: string, options: RequestInit = {}): Promise<Response> {
  const token = useAuthStore.getState().token

  if (!token) {
    throw new Error('No authentication token available')
  }

  const headers = {
    'Authorization': `Bearer ${token}`,
    ...options.headers,
  }

  const response = await fetch(url, {
    ...options,
    headers,
  })

  // Handle 401 Unauthorized - token expired or invalid
  if (response.status === 401) {
    // Clear auth and redirect to login
    useAuthStore.getState().clearAuth()
    throw new Error('Authentication failed. Please login again.')
  }

  return response
}
