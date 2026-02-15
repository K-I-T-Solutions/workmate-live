import { useAgentStore } from '@/store/agentStore'
import { useOBSStore } from '@/store/obsStore'
import { useTwitchStore } from '@/store/twitchStore'
import { useYouTubeStore } from '@/store/youtubeStore'

export interface ConnectionInfo {
  label: string
  connected: boolean
  detail?: string
}

export function useConnectionStatus() {
  const agentConnected = useAgentStore((s) => s.connected)
  const obsStatus = useOBSStore((s) => s.status)
  const twitchStatus = useTwitchStore((s) => s.status)
  const youtubeStatus = useYouTubeStore((s) => s.status)

  const connections: ConnectionInfo[] = [
    {
      label: 'Agent',
      connected: agentConnected,
    },
    {
      label: 'OBS',
      connected: obsStatus?.connected ?? false,
      detail: obsStatus?.version,
    },
    {
      label: 'Twitch',
      connected: twitchStatus?.connected ?? false,
      detail: twitchStatus?.channel,
    },
    {
      label: 'YouTube',
      connected: youtubeStatus?.connected ?? false,
      detail: youtubeStatus?.channel_id,
    },
  ]

  const allConnected = connections.every((c) => c.connected)
  const connectedCount = connections.filter((c) => c.connected).length

  return { connections, allConnected, connectedCount }
}
