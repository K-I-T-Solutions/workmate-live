import { PageContainer } from '@/components/shared/PageContainer'
import { MetricCard } from '@/components/shared/MetricCard'
import { GlowCard } from '@/components/shared/GlowCard'
import { SectionHeader } from '@/components/shared/SectionHeader'
import { StatusIndicator } from '@/components/shared/StatusIndicator'
import { useAgentStore } from '@/store/agentStore'
import { useOBSStore } from '@/store/obsStore'
import { useTwitchStore } from '@/store/twitchStore'
import { useYouTubeStore } from '@/store/youtubeStore'
import { Activity, Video, Radio, Youtube, Eye, Clock, MessageCircle, Heart, Gift, Users } from 'lucide-react'
import type { FollowEvent, SubscribeEvent, RaidEvent } from '@/types/twitch'

export function DashboardPage() {
  const agentConnected = useAgentStore((s) => s.connected)
  const obsStatus = useOBSStore((s) => s.status)
  const twitchStats = useTwitchStore((s) => s.stats)
  const twitchEvents = useTwitchStore((s) => s.events)
  const twitchChat = useTwitchStore((s) => s.chatMessages)
  const youtubeStats = useYouTubeStore((s) => s.stats)
  const youtubeChat = useYouTubeStore((s) => s.chatMessages)

  const formatDuration = (seconds: number) => {
    const h = Math.floor(seconds / 3600)
    const m = Math.floor((seconds % 3600) / 60)
    const s = seconds % 60
    return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  }

  const recentEvents = twitchEvents.slice(0, 5)
  const recentChat = [...twitchChat.slice(-5).map(m => ({ ...m, platform: 'twitch' as const })), ...youtubeChat.slice(-5).map(m => ({ ...m, platform: 'youtube' as const, display_name: ('author_name' in m ? m.author_name : ''), color: undefined }))]
    .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
    .slice(0, 10)

  return (
    <PageContainer title="Dashboard" subtitle="Live-Übersicht aller Services">
      {/* Row 1: Metric Cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <MetricCard
          label="Agent"
          value={agentConnected ? 'Online' : 'Offline'}
          icon={Activity}
          iconColor={agentConnected ? 'text-success' : 'text-destructive'}
          subtext={agentConnected ? 'Connected' : 'Disconnected'}
        />
        <MetricCard
          label="OBS Studio"
          value={obsStatus?.connected ? 'Connected' : 'Offline'}
          icon={Video}
          iconColor={obsStatus?.connected ? 'text-success' : 'text-muted-foreground'}
          subtext={obsStatus?.current_scene || 'No scene'}
        />
        <MetricCard
          label="Twitch Viewers"
          value={twitchStats?.is_live ? twitchStats.viewer_count.toLocaleString() : '—'}
          icon={Radio}
          iconColor="text-primary"
          subtext={twitchStats?.is_live ? 'Live' : 'Offline'}
        />
        <MetricCard
          label="YouTube Viewers"
          value={youtubeStats?.is_live ? youtubeStats.viewer_count.toLocaleString() : '—'}
          icon={Youtube}
          iconColor="text-destructive"
          subtext={youtubeStats?.is_live ? 'Live' : 'Offline'}
        />
      </div>

      {/* Row 2: Live Status */}
      <GlowCard>
        <SectionHeader title="Live Status" icon={Eye} />
        <div className="grid md:grid-cols-2 gap-4">
          {/* OBS Status */}
          <div className="space-y-2">
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <Video className="h-3.5 w-3.5" />
              OBS
            </div>
            {obsStatus?.streaming?.active ? (
              <div className="flex items-center gap-3 p-2 rounded-md bg-success/10 border border-success/20">
                <StatusIndicator connected={true} size="sm" />
                <div className="flex-1">
                  <span className="text-xs font-medium text-success">Streaming</span>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                    <Clock className="h-3 w-3" />
                    {formatDuration(obsStatus.streaming.duration)}
                  </div>
                </div>
              </div>
            ) : (
              <div className="flex items-center gap-3 p-2 rounded-md bg-muted/50">
                <StatusIndicator connected={false} size="sm" pulse={false} />
                <span className="text-xs text-muted-foreground">Not streaming</span>
              </div>
            )}
            {obsStatus?.recording?.active ? (
              <div className="flex items-center gap-3 p-2 rounded-md bg-destructive/10 border border-destructive/20">
                <div className="h-1.5 w-1.5 rounded-full bg-destructive animate-status-pulse" />
                <div className="flex-1">
                  <span className="text-xs font-medium text-destructive">Recording</span>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                    <Clock className="h-3 w-3" />
                    {formatDuration(obsStatus.recording.duration)}
                  </div>
                </div>
              </div>
            ) : null}
          </div>

          {/* Platform Status */}
          <div className="space-y-2">
            {twitchStats?.is_live && (
              <div className="p-2 rounded-md bg-primary/10 border border-primary/20">
                <div className="flex items-center gap-2">
                  <Radio className="h-3.5 w-3.5 text-primary" />
                  <span className="text-xs font-medium text-primary">Twitch Live</span>
                  <span className="ml-auto text-xs font-bold text-primary">{twitchStats.viewer_count} viewers</span>
                </div>
                {twitchStats.title && (
                  <p className="text-xs text-muted-foreground mt-1 truncate">{twitchStats.title}</p>
                )}
                {twitchStats.game_name && (
                  <span className="inline-block mt-1 px-1.5 py-0.5 rounded text-[10px] bg-primary/20 text-primary">{twitchStats.game_name}</span>
                )}
              </div>
            )}
            {youtubeStats?.is_live && (
              <div className="p-2 rounded-md bg-destructive/10 border border-destructive/20">
                <div className="flex items-center gap-2">
                  <Youtube className="h-3.5 w-3.5 text-destructive" />
                  <span className="text-xs font-medium text-destructive">YouTube Live</span>
                  <span className="ml-auto text-xs font-bold text-destructive">{youtubeStats.viewer_count} viewers</span>
                </div>
                {youtubeStats.title && (
                  <p className="text-xs text-muted-foreground mt-1 truncate">{youtubeStats.title}</p>
                )}
              </div>
            )}
            {!twitchStats?.is_live && !youtubeStats?.is_live && (
              <div className="flex items-center gap-3 p-2 rounded-md bg-muted/50">
                <StatusIndicator connected={false} size="sm" pulse={false} />
                <span className="text-xs text-muted-foreground">No platforms live</span>
              </div>
            )}
          </div>
        </div>
      </GlowCard>

      {/* Row 3: Activity Feed */}
      <div className="grid md:grid-cols-2 gap-3">
        {/* Recent Events */}
        <GlowCard glowColor="secondary">
          <SectionHeader title="Recent Events" icon={Heart} iconColor="text-secondary" />
          <div className="space-y-2">
            {recentEvents.length === 0 ? (
              <p className="text-xs text-muted-foreground text-center py-4">No recent events</p>
            ) : (
              recentEvents.map((event, idx) => {
                const time = new Date(event.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
                switch (event.type) {
                  case 'follow': {
                    const d = event.data as FollowEvent
                    return (
                      <div key={idx} className="flex items-center gap-2 p-2 rounded-md bg-muted/30 text-xs">
                        <Heart className="h-3.5 w-3.5 text-secondary shrink-0" />
                        <span className="font-medium truncate">{d.user_name}</span>
                        <span className="text-muted-foreground">followed</span>
                        <span className="ml-auto text-muted-foreground shrink-0">{time}</span>
                      </div>
                    )
                  }
                  case 'subscribe': {
                    const d = event.data as SubscribeEvent
                    return (
                      <div key={idx} className="flex items-center gap-2 p-2 rounded-md bg-muted/30 text-xs">
                        <Gift className="h-3.5 w-3.5 text-primary shrink-0" />
                        <span className="font-medium truncate">{d.user_name}</span>
                        <span className="text-muted-foreground">subscribed (T{d.tier})</span>
                        <span className="ml-auto text-muted-foreground shrink-0">{time}</span>
                      </div>
                    )
                  }
                  case 'raid': {
                    const d = event.data as RaidEvent
                    return (
                      <div key={idx} className="flex items-center gap-2 p-2 rounded-md bg-muted/30 text-xs">
                        <Users className="h-3.5 w-3.5 text-warning shrink-0" />
                        <span className="font-medium truncate">{d.from_user_name}</span>
                        <span className="text-muted-foreground">raided ({d.viewers})</span>
                        <span className="ml-auto text-muted-foreground shrink-0">{time}</span>
                      </div>
                    )
                  }
                }
              })
            )}
          </div>
        </GlowCard>

        {/* Recent Chat */}
        <GlowCard>
          <SectionHeader title="Recent Chat" icon={MessageCircle} />
          <div className="space-y-1.5">
            {recentChat.length === 0 ? (
              <p className="text-xs text-muted-foreground text-center py-4">No chat messages</p>
            ) : (
              recentChat.map((msg, idx) => (
                <div key={idx} className="flex items-start gap-2 p-1.5 rounded text-xs">
                  <span className={msg.platform === 'twitch' ? 'text-primary' : 'text-destructive'}>
                    {msg.platform === 'twitch' ? '●' : '▶'}
                  </span>
                  <span className="font-medium shrink-0" style={{ color: ('color' in msg && msg.color) || undefined }}>
                    {msg.display_name || ('author_name' in msg ? msg.author_name : '')}
                  </span>
                  <span className="text-muted-foreground truncate">
                    {'message' in msg ? msg.message : ''}
                  </span>
                </div>
              ))
            )}
          </div>
        </GlowCard>
      </div>
    </PageContainer>
  )
}
