export interface User {
  id: number
  username: string
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
  expires_in: string
}

export interface VerifyResponse {
  valid: boolean
  user: User
}
