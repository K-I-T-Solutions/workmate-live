import { useEffect, useRef } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useYouTubeStore } from '@/store/youtubeStore'
import { MessageSquare, Trash2, Shield, Star, Crown } from 'lucide-react'

export function YouTubeChat() {
  const { chatMessages, clearChat } = useYouTubeStore()
  const chatEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [chatMessages])

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <MessageSquare className="w-5 h-5 text-red-600" />
            YouTube Live Chat
          </CardTitle>
          <Button variant="outline" size="sm" onClick={clearChat} title="Clear chat">
            <Trash2 className="w-4 h-4" />
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-2 h-[500px] overflow-y-auto">
          {chatMessages.length === 0 ? (
            <div className="text-sm text-muted-foreground text-center py-8">
              No chat messages yet
            </div>
          ) : (
            chatMessages.map((msg) => (
              <div key={msg.id} className="flex flex-col gap-1 p-2 hover:bg-muted/50 rounded">
                <div className="flex items-center gap-2">
                  <span className="font-semibold text-sm">{msg.author_name}</span>
                  {msg.is_owner && (
                    <Badge variant="default" className="h-5 px-1.5 bg-red-600 text-xs">
                      <Crown className="w-3 h-3" />
                    </Badge>
                  )}
                  {msg.is_moderator && (
                    <Badge variant="default" className="h-5 px-1.5 bg-green-600 text-xs">
                      <Shield className="w-3 h-3" />
                    </Badge>
                  )}
                  {msg.is_sponsor && (
                    <Badge variant="default" className="h-5 px-1.5 bg-purple-600 text-xs">
                      <Star className="w-3 h-3" />
                    </Badge>
                  )}
                </div>
                <div className="text-sm">{msg.message}</div>
              </div>
            ))
          )}
          <div ref={chatEndRef} />
        </div>
      </CardContent>
    </Card>
  )
}
