import { PageContainer } from '@/components/shared/PageContainer'
import { TwitchStats } from '@/components/twitch/TwitchStats'
import { TwitchChat } from '@/components/twitch/TwitchChat'
import { TwitchEventAlerts } from '@/components/twitch/TwitchEventAlerts'
import { StreamMetadataEditor } from '@/components/twitch/StreamMetadataEditor'

export function TwitchPage() {
  return (
    <PageContainer title="Twitch" subtitle="Stream-Statistiken, Chat und Events">
      <div className="grid lg:grid-cols-2 gap-3">
        <div className="space-y-3">
          <TwitchStats />
          <StreamMetadataEditor />
        </div>
        <div className="space-y-3">
          <TwitchChat />
          <TwitchEventAlerts />
        </div>
      </div>
    </PageContainer>
  )
}
