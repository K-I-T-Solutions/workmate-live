import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { useYouTubeStore } from '@/store/youtubeStore'
import { youtubeAPI } from '@/services/youtube'
import { Eye, Users, Video, Clock } from 'lucide-react'

export function YouTubeStats() {
  const { stats, setStats } = useYouTubeStore()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const data = await youtubeAPI.getStats()
        setStats(data)
        setError('')
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch stats')
      } finally {
        setLoading(false)
      }
    }

    fetchStats()
    const interval = setInterval(fetchStats, 30000) // Refresh every 30s

    return () => clearInterval(interval)
  }, [setStats])

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Video className="w-5 h-5 text-destructive" />
            YouTube Stats
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-sm text-muted-foreground">Loading...</div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Video className="w-5 h-5 text-destructive" />
            YouTube Stats
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-sm text-destructive">{error}</div>
        </CardContent>
      </Card>
    )
  }

  if (!stats) return null

  const formatUptime = (startTime?: string) => {
    if (!startTime) return 'N/A'
    const start = new Date(startTime)
    const now = new Date()
    const diffMs = now.getTime() - start.getTime()
    const hours = Math.floor(diffMs / (1000 * 60 * 60))
    const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60))
    return `${hours}h ${minutes}m`
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Video className="w-5 h-5 text-destructive" />
            YouTube Stats
          </CardTitle>
          <Badge variant={stats.is_live ? 'destructive' : 'secondary'}>
            {stats.is_live ? '● LIVE' : 'Offline'}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {stats.is_live && (
          <>
            <div className="flex items-center justify-between">
              <span className="text-sm flex items-center gap-2">
                <Eye className="w-4 h-4" />
                Viewers
              </span>
              <span className="text-sm font-medium">{stats.viewer_count.toLocaleString()}</span>
            </div>

            <div className="flex items-center justify-between">
              <span className="text-sm flex items-center gap-2">
                <Clock className="w-4 h-4" />
                Uptime
              </span>
              <span className="text-sm font-medium">{formatUptime(stats.actual_start_time)}</span>
            </div>
          </>
        )}

        <div className="flex items-center justify-between">
          <span className="text-sm flex items-center gap-2">
            <Users className="w-4 h-4" />
            Subscribers
          </span>
          <span className="text-sm font-medium">{stats.subscriber_count.toLocaleString()}</span>
        </div>

        <div className="flex items-center justify-between">
          <span className="text-sm flex items-center gap-2">
            <Video className="w-4 h-4" />
            Total Videos
          </span>
          <span className="text-sm font-medium">{stats.video_count.toLocaleString()}</span>
        </div>

        {stats.is_live && stats.title && (
          <div className="pt-2 border-t">
            <div className="text-xs text-muted-foreground mb-1">Stream Title</div>
            <div className="text-sm font-medium line-clamp-2">{stats.title}</div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
