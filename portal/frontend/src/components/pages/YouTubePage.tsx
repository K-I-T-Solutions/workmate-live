import { PageContainer } from '@/components/shared/PageContainer'
import { YouTubeStats } from '@/components/youtube/YouTubeStats'
import { YouTubeChat } from '@/components/youtube/YouTubeChat'

export function YouTubePage() {
  return (
    <PageContainer title="YouTube" subtitle="Live-Stream Statistiken und Chat">
      <YouTubeStats />
      <YouTubeChat />
    </PageContainer>
  )
}
