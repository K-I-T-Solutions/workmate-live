import type { AgentStatus } from '@/types/agent'
import type { OBSEvent } from '@/types/obs'
import type { ChatMessage, TwitchEvent } from '@/types/twitch'
import { useAgentStore } from '@/store/agentStore'
import { useOBSStore } from '@/store/obsStore'
import { useTwitchStore } from '@/store/twitchStore'

export interface WebSocketMessage {
  type: string
  data: unknown
}

class WebSocketService {
  private ws: WebSocket | null = null
  private reconnectTimeout: number | null = null
  private reconnectDelay = 3000

  connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const hostname = window.location.hostname
    const wsUrl = `${protocol}//${hostname}:8080/ws`

    this.ws = new WebSocket(wsUrl)

    this.ws.onopen = () => {
      console.log('WebSocket connected')
      useAgentStore.getState().setConnected(true)

      if (this.reconnectTimeout) {
        clearTimeout(this.reconnectTimeout)
        this.reconnectTimeout = null
      }
    }

    this.ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data)
        this.handleMessage(message)
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error)
      }
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }

    this.ws.onclose = () => {
      console.log('WebSocket disconnected')
      useAgentStore.getState().setConnected(false)
      this.scheduleReconnect()
    }
  }

  private handleMessage(message: WebSocketMessage) {
    switch (message.type) {
      case 'agent_status':
        useAgentStore.getState().setStatus(message.data as AgentStatus)
        break
      case 'obs_event':
        this.handleOBSEvent(message.data as OBSEvent)
        break
      case 'twitch_chat':
        useTwitchStore.getState().addChatMessage(message.data as ChatMessage)
        break
      case 'twitch_event':
        useTwitchStore.getState().addEvent(message.data as TwitchEvent)
        break
      case 'youtube_chat':
        // Handle YouTube chat (Phase 5)
        console.log('YouTube chat:', message.data)
        break
      default:
        console.log('Unknown message type:', message.type)
    }
  }

  private handleOBSEvent(event: OBSEvent) {
    switch (event.type) {
      case 'scene_changed':
        if (event.scene_name) {
          useOBSStore.getState().updateScene(event.scene_name)
        }
        break
      case 'stream_state_changed':
      case 'record_state_changed':
        // Trigger a status refresh
        console.log('OBS state changed:', event)
        break
      case 'source_visibility_changed':
        console.log('Source visibility changed:', event)
        break
      default:
        console.log('Unknown OBS event:', event)
    }
  }

  private scheduleReconnect() {
    if (this.reconnectTimeout) {
      return
    }

    this.reconnectTimeout = window.setTimeout(() => {
      console.log('Attempting to reconnect WebSocket...')
      this.connect()
    }, this.reconnectDelay)
  }

  disconnect() {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout)
      this.reconnectTimeout = null
    }

    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  send(message: WebSocketMessage) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    }
  }
}

export const wsService = new WebSocketService()
