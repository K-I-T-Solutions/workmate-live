import { useEffect, useRef } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useTwitchStore } from '@/store/twitchStore'
import { MessageCircle, Trash2, Shield, Star } from 'lucide-react'

export function TwitchChat() {
  const { chatMessages, clearChat } = useTwitchStore()
  const chatEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [chatMessages])

  return (
    <Card className="flex flex-col h-[500px]">
      <CardHeader className="flex-none flex flex-row items-center justify-between pb-3">
        <CardTitle className="flex items-center gap-2">
          <MessageCircle className="w-5 h-5 text-primary" />
          Live Chat
          {chatMessages.length > 0 && (
            <span className="text-xs font-normal text-muted-foreground">
              ({chatMessages.length})
            </span>
          )}
        </CardTitle>
        <Button
          variant="ghost"
          size="icon"
          onClick={clearChat}
          title="Clear chat"
          className="h-8 w-8 hover:bg-destructive/10 hover:text-destructive"
        >
          <Trash2 className="w-4 h-4" />
        </Button>
      </CardHeader>
      <CardContent className="flex-1 overflow-y-auto space-y-2 p-4">
        {chatMessages.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center space-y-2">
            <MessageCircle className="w-12 h-12 text-muted-foreground/30" />
            <p className="text-sm text-muted-foreground">No messages yet...</p>
            <p className="text-xs text-muted-foreground/60">Chat messages will appear here when viewers send them</p>
          </div>
        ) : (
          chatMessages.map((msg, idx) => (
            <div key={idx} className="group hover:bg-muted/50 -mx-2 px-2 py-1.5 rounded transition-colors">
              <div className="flex items-start gap-2">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-1.5 flex-wrap">
                    <span
                      className="font-bold text-sm truncate"
                      style={{ color: msg.color || '#9147ff' }}
                    >
                      {msg.display_name || msg.username}
                    </span>
                    {msg.is_moderator && (
                      <span className="inline-flex items-center gap-1 px-1.5 py-0.5 text-xs font-semibold bg-success text-success-foreground rounded">
                        <Shield className="w-3 h-3" />
                        MOD
                      </span>
                    )}
                    {msg.is_subscriber && (
                      <span className="inline-flex items-center gap-1 px-1.5 py-0.5 text-xs font-semibold bg-primary text-primary-foreground rounded">
                        <Star className="w-3 h-3" />
                        SUB
                      </span>
                    )}
                    {msg.badges.length > 0 && msg.badges.map((badge, bidx) => (
                      <span
                        key={bidx}
                        className="px-1.5 py-0.5 text-xs font-medium bg-muted text-muted-foreground rounded"
                      >
                        {badge}
                      </span>
                    ))}
                  </div>
                  <p className="text-sm mt-0.5 break-words leading-relaxed">{msg.message}</p>
                </div>
              </div>
            </div>
          ))
        )}
        <div ref={chatEndRef} />
      </CardContent>
    </Card>
  )
}
