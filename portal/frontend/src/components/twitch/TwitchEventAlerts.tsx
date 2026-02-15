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
            className="group relative flex items-start gap-3 p-4 bg-secondary/10 rounded-lg border border-secondary/20 transition-all hover:shadow-md hover:scale-[1.02]"
          >
            <div className="flex items-center justify-center w-10 h-10 bg-secondary rounded-full flex-shrink-0">
              <Heart className="w-5 h-5 text-secondary-foreground" fill="currentColor" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-bold text-secondary truncate">
                {data.user_name} followed!
              </p>
              <p className="text-xs text-muted-foreground mt-0.5">{timestamp}</p>
            </div>
            <Sparkles className="w-4 h-4 text-secondary opacity-0 group-hover:opacity-100 transition-opacity" />
          </div>
        )
      }
      case 'subscribe': {
        const data = event.data as SubscribeEvent
        return (
          <div
            key={idx}
            className="group relative flex items-start gap-3 p-4 bg-primary/10 rounded-lg border border-primary/20 transition-all hover:shadow-md hover:scale-[1.02]"
          >
            <div className="flex items-center justify-center w-10 h-10 bg-primary rounded-full flex-shrink-0">
              <Gift className="w-5 h-5 text-primary-foreground" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-bold text-primary truncate">
                {data.user_name} subscribed!
              </p>
              <div className="flex items-center gap-2 mt-1">
                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-primary text-primary-foreground">
                  Tier {data.tier}
                </span>
                {data.is_gift && (
                  <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-destructive text-destructive-foreground">
                    Gift
                  </span>
                )}
                <span className="text-xs text-muted-foreground">{timestamp}</span>
              </div>
            </div>
            <Sparkles className="w-4 h-4 text-primary opacity-0 group-hover:opacity-100 transition-opacity" />
          </div>
        )
      }
      case 'raid': {
        const data = event.data as RaidEvent
        return (
          <div
            key={idx}
            className="group relative flex items-start gap-3 p-4 bg-warning/10 rounded-lg border border-warning/20 transition-all hover:shadow-md hover:scale-[1.02]"
          >
            <div className="flex items-center justify-center w-10 h-10 bg-warning rounded-full flex-shrink-0">
              <Users className="w-5 h-5 text-warning-foreground" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-bold text-warning truncate">
                {data.from_user_name} raided!
              </p>
              <div className="flex items-center gap-2 mt-1">
                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-warning text-warning-foreground">
                  {data.viewers} viewers
                </span>
                <span className="text-xs text-muted-foreground">{timestamp}</span>
              </div>
            </div>
            <Sparkles className="w-4 h-4 text-warning opacity-0 group-hover:opacity-100 transition-opacity" />
          </div>
        )
      }
    }
  }

  return (
    <Card className="flex flex-col h-[400px]">
      <CardHeader className="flex-none flex flex-row items-center justify-between pb-3">
        <CardTitle className="flex items-center gap-2">
          <Bell className="w-5 h-5 text-primary" />
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
          className="h-8 w-8 hover:bg-destructive/10 hover:text-destructive"
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
