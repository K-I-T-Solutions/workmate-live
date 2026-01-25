import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useTwitchStore } from '@/store/twitchStore'
import { twitchAPI } from '@/services/twitch'
import { Eye, Users, Clock, RefreshCw, Radio, Globe, CalendarClock } from 'lucide-react'

export function TwitchStats() {
  const { stats, setStats } = useTwitchStore()
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const loadStats = async () => {
    try {
      setRefreshing(true)
      const streamStats = await twitchAPI.getStats()
      setStats(streamStats)
      setError(null)
    } catch (err) {
      console.error('Failed to load Twitch stats:', err)
      setError('Failed to load stats')
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => {
    loadStats()
    const interval = setInterval(loadStats, 30000) // Refresh every 30s
    return () => clearInterval(interval)
  }, [setStats])

  const formatUptime = (seconds: number) => {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    return `${hours}h ${minutes}m`
  }

  const getLanguageDisplay = (lang?: string) => {
    const languages: Record<string, { name: string; flag: string }> = {
      en: { name: 'English', flag: '🇬🇧' },
      de: { name: 'Deutsch', flag: '🇩🇪' },
      fr: { name: 'Français', flag: '🇫🇷' },
      es: { name: 'Español', flag: '🇪🇸' },
      pt: { name: 'Português', flag: '🇵🇹' },
      ru: { name: 'Русский', flag: '🇷🇺' },
      ja: { name: '日本語', flag: '🇯🇵' },
      ko: { name: '한국어', flag: '🇰🇷' },
      zh: { name: '中文', flag: '🇨🇳' },
    }
    return lang && languages[lang] ? languages[lang] : { name: lang || 'Unknown', flag: '🌐' }
  }

  const getThumbnailUrl = (url?: string) => {
    if (!url) return null
    // Replace {width}x{height} placeholders with actual dimensions
    return url.replace('{width}', '440').replace('{height}', '248')
  }

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Radio className="w-5 h-5 text-purple-600" />
            Twitch Stats
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="animate-pulse space-y-3">
            <div className="h-4 bg-muted rounded w-3/4"></div>
            <div className="h-4 bg-muted rounded w-1/2"></div>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Radio className="w-5 h-5 text-purple-600" />
            Twitch Stats
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-red-600">{error}</p>
          <Button size="sm" variant="outline" onClick={loadStats} className="mt-3">
            <RefreshCw className="w-4 h-4 mr-2" />
            Retry
          </Button>
        </CardContent>
      </Card>
    )
  }

  if (!stats) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Radio className="w-5 h-5 text-purple-600" />
            Twitch Stats
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">No data available</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Radio className="w-5 h-5 text-purple-600" />
            Twitch Stats
          </CardTitle>
          <Button
            size="icon"
            variant="ghost"
            onClick={loadStats}
            disabled={refreshing}
            className="h-8 w-8"
          >
            <RefreshCw className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`} />
          </Button>
        </div>
        {stats.is_live ? (
          <div className="flex items-center gap-2 mt-2">
            <span className="relative flex h-3 w-3">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-3 w-3 bg-red-500"></span>
            </span>
            <span className="text-sm font-bold text-red-600 dark:text-red-500">LIVE NOW</span>
          </div>
        ) : (
          <span className="inline-flex items-center gap-1.5 mt-2 text-sm text-muted-foreground">
            <div className="w-2 h-2 rounded-full bg-muted-foreground/40"></div>
            Offline
          </span>
        )}
      </CardHeader>
      <CardContent className="space-y-3">
        {stats.is_live && getThumbnailUrl(stats.thumbnail_url) && (
          <div className="relative overflow-hidden rounded-lg border border-purple-200/50 dark:border-purple-800/30">
            <img
              src={getThumbnailUrl(stats.thumbnail_url)!}
              alt="Stream Preview"
              className="w-full h-auto"
              onError={(e) => {
                e.currentTarget.style.display = 'none'
              }}
            />
            <div className="absolute top-2 right-2">
              <Badge variant="destructive" className="bg-red-600 text-white font-bold">
                ● LIVE
              </Badge>
            </div>
          </div>
        )}
        {stats.is_live && (
          <>
            <div className="flex items-center gap-3 p-3 bg-gradient-to-r from-purple-50 to-pink-50 dark:from-purple-950/20 dark:to-pink-950/20 rounded-lg">
              <Eye className="w-5 h-5 text-purple-600 dark:text-purple-400" />
              <span className="text-sm font-medium">Viewers</span>
              <span className="ml-auto text-lg font-bold text-purple-600 dark:text-purple-400">
                {stats.viewer_count.toLocaleString()}
              </span>
            </div>
            <div className="flex items-center gap-3 p-3 bg-muted/50 rounded-lg">
              <Clock className="w-5 h-5 text-blue-600 dark:text-blue-400" />
              <span className="text-sm font-medium">Uptime</span>
              <span className="ml-auto font-mono text-sm font-semibold text-blue-600 dark:text-blue-400">
                {formatUptime(stats.uptime)}
              </span>
            </div>
            {stats.started_at && (
              <div className="flex items-center gap-3 p-3 bg-muted/50 rounded-lg">
                <CalendarClock className="w-5 h-5 text-orange-600 dark:text-orange-400" />
                <span className="text-sm font-medium">Started</span>
                <span className="ml-auto text-xs font-medium text-orange-600 dark:text-orange-400">
                  {new Date(stats.started_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </span>
              </div>
            )}
            {stats.language && (
              <div className="flex items-center gap-3 p-3 bg-muted/50 rounded-lg">
                <Globe className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
                <span className="text-sm font-medium">Language</span>
                <span className="ml-auto text-sm font-semibold flex items-center gap-1.5">
                  <span>{getLanguageDisplay(stats.language).flag}</span>
                  <span className="text-indigo-600 dark:text-indigo-400">{getLanguageDisplay(stats.language).name}</span>
                </span>
              </div>
            )}
          </>
        )}
        <div className="flex items-center gap-3 p-3 bg-muted/50 rounded-lg">
          <Users className="w-5 h-5 text-green-600 dark:text-green-400" />
          <span className="text-sm font-medium">Followers</span>
          <span className="ml-auto text-lg font-bold text-green-600 dark:text-green-400">
            {stats.follower_count.toLocaleString()}
          </span>
        </div>
        {stats.is_live && stats.title && (
          <div className="pt-3 mt-3 border-t space-y-2">
            <p className="text-sm font-semibold line-clamp-2 leading-relaxed">{stats.title}</p>
            {stats.game_name && (
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300">
                {stats.game_name}
              </span>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
