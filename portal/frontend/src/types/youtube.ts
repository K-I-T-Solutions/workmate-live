export interface YouTubeStatus {
  connected: boolean
  live_chat_ready: boolean
  channel_id?: string
  broadcast_id?: string
}

export interface StreamStats {
  is_live: boolean
  viewer_count: number
  subscriber_count: number
  video_count: number
  title?: string
  description?: string
  scheduled_start_time?: string
  actual_start_time?: string
}

export interface ChatMessage {
  id: string
  author_name: string
  author_channel_id: string
  message: string
  timestamp: string
  is_moderator: boolean
  is_sponsor: boolean
  is_owner: boolean
}

export interface UpdateStreamRequest {
  title?: string
  description?: string
  category_id?: string
}
