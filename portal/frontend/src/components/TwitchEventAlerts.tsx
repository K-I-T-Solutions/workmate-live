import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useTwitchStore } from '@/store/twitchStore'
import { Trash2, Heart, Gift, Users, Bell, Sparkles } from 'lucide-react'
import type { FollowEvent, SubscribeEvent, RaidEvent } from '@/types/twitch'

export function TwitchEventAlerts() {
  const { events, clearEvents } = useTwitchStore()

  const renderEvent = (event: typeof events[0], idx: number) => {
    const timestamp = new Date(event.timestamp).toLocaleTimeString()

    switch (event.type) {
      case 'follow': {
        const data = event.data as FollowEvent
        return (
          <div
            key={idx}
            className="group relative flex items-start gap-3 p-4 bg-gradient-to-br from-blue-50 to-cyan-50 dark:from-blue-950/30 dark:to-cyan-950/30 rounded-lg border border-blue-200/50 dark:border-blue-800/30 transition-all hover:shadow-md hover:scale-[1.02]"
          >
            <div className="flex items-center justify-center w-10 h-10 bg-blue-600 rounded-full flex-shrink-0">
              <Heart className="w-5 h-5 text-white" fill="currentColor" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-bold text-blue-900 dark:text-blue-300 truncate">
                {data.user_name} followed!
              </p>
              <p className="text-xs text-blue-700/70 dark:text-blue-400/70 mt-0.5">{timestamp}</p>
            </div>
            <Sparkles className="w-4 h-4 text-blue-400 opacity-0 group-hover:opacity-100 transition-opacity" />
          </div>
        )
      }
      case 'subscribe': {
        const data = event.data as SubscribeEvent
        return (
          <div
            key={idx}
            className="group relative flex items-start gap-3 p-4 bg-gradient-to-br from-purple-50 to-pink-50 dark:from-purple-950/30 dark:to-pink-950/30 rounded-lg border border-purple-200/50 dark:border-purple-800/30 transition-all hover:shadow-md hover:scale-[1.02]"
          >
            <div className="flex items-center justify-center w-10 h-10 bg-purple-600 rounded-full flex-shrink-0">
              <Gift className="w-5 h-5 text-white" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-bold text-purple-900 dark:text-purple-300 truncate">
                {data.user_name} subscribed!
              </p>
              <div className="flex items-center gap-2 mt-1">
                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-purple-600 text-white">
                  Tier {data.tier}
                </span>
                {data.is_gift && (
                  <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-pink-600 text-white">
                    Gift
                  </span>
                )}
                <span className="text-xs text-purple-700/70 dark:text-purple-400/70">{timestamp}</span>
              </div>
            </div>
            <Sparkles className="w-4 h-4 text-purple-400 opacity-0 group-hover:opacity-100 transition-opacity" />
          </div>
        )
      }
      case 'raid': {
        const data = event.data as RaidEvent
        return (
          <div
            key={idx}
            className="group relative flex items-start gap-3 p-4 bg-gradient-to-br from-orange-50 to-yellow-50 dark:from-orange-950/30 dark:to-yellow-950/30 rounded-lg border border-orange-200/50 dark:border-orange-800/30 transition-all hover:shadow-md hover:scale-[1.02]"
          >
            <div className="flex items-center justify-center w-10 h-10 bg-orange-600 rounded-full flex-shrink-0">
              <Users className="w-5 h-5 text-white" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-bold text-orange-900 dark:text-orange-300 truncate">
                {data.from_user_name} raided!
              </p>
              <div className="flex items-center gap-2 mt-1">
                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-orange-600 text-white">
                  {data.viewers} viewers
                </span>
                <span className="text-xs text-orange-700/70 dark:text-orange-400/70">{timestamp}</span>
              </div>
            </div>
            <Sparkles className="w-4 h-4 text-orange-400 opacity-0 group-hover:opacity-100 transition-opacity" />
          </div>
        )
      }
    }
  }

  return (
    <Card className="flex flex-col h-[400px]">
      <CardHeader className="flex-none flex flex-row items-center justify-between pb-3">
        <CardTitle className="flex items-center gap-2">
          <Bell className="w-5 h-5 text-purple-600" />
          Event Alerts
          {events.length > 0 && (
            <span className="text-xs font-normal text-muted-foreground">
              ({events.length})
            </span>
          )}
        </CardTitle>
        <Button
          variant="ghost"
          size="icon"
          onClick={clearEvents}
          title="Clear events"
          className="h-8 w-8 hover:bg-red-100 hover:text-red-600 dark:hover:bg-red-950/30"
        >
          <Trash2 className="w-4 h-4" />
        </Button>
      </CardHeader>
      <CardContent className="flex-1 overflow-y-auto space-y-3 p-4">
        {events.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center space-y-2">
            <Bell className="w-12 h-12 text-muted-foreground/30" />
            <p className="text-sm text-muted-foreground">No events yet...</p>
            <p className="text-xs text-muted-foreground/60">Follows, subs, and raids will appear here</p>
          </div>
        ) : (
          events.map((event, idx) => renderEvent(event, idx))
        )}
      </CardContent>
    </Card>
  )
}
