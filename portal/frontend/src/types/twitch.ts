export interface TwitchStatus {
  connected: boolean
  chat_connected: boolean
  events_connected: boolean
  channel: string
  user_id?: string
}

export interface StreamStats {
  is_live: boolean
  viewer_count: number
  follower_count: number
  uptime: number // seconds
  started_at?: string
  title: string
  game_name: string
  language?: string
  thumbnail_url?: string
  stream_id?: string
}

export interface ChatMessage {
  username: string
  display_name: string
  message: string
  color: string
  timestamp: string
  is_moderator: boolean
  is_subscriber: boolean
  badges: string[]
}

export interface TwitchEvent {
  type: 'follow' | 'subscribe' | 'raid'
  data: FollowEvent | SubscribeEvent | RaidEvent
  timestamp: string
}

export interface FollowEvent {
  user_id: string
  user_login: string
  user_name: string
  followed_at: string
}

export interface SubscribeEvent {
  user_id: string
  user_login: string
  user_name: string
  tier: string
  is_gift: boolean
}

export interface RaidEvent {
  from_user_id: string
  from_user_login: string
  from_user_name: string
  viewers: number
}

export interface UpdateStreamRequest {
  title?: string
  game_id?: string
  game_name?: string
}
