import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useTwitchStore } from '@/store/twitchStore'
import { twitchAPI } from '@/services/twitch'
import { Edit3, Gamepad2, CheckCircle2, AlertCircle, Loader2 } from 'lucide-react'

export function StreamMetadataEditor() {
  const { stats } = useTwitchStore()
  const [title, setTitle] = useState('')
  const [gameName, setGameName] = useState('')
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleUpdate = async () => {
    if (!title && !gameName) return

    setLoading(true)
    setSuccess(false)
    setError(null)

    try {
      await twitchAPI.updateStream({
        title: title || undefined,
        game_name: gameName || undefined,
      })

      setTitle('')
      setGameName('')
      setSuccess(true)
      setTimeout(() => setSuccess(false), 3000)
    } catch (err) {
      console.error('Failed to update stream:', err)
      setError('Failed to update stream metadata')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Edit3 className="w-5 h-5 text-purple-600" />
          Stream Metadata
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {stats && (
          <div className="p-4 bg-gradient-to-br from-purple-50 to-blue-50 dark:from-purple-950/20 dark:to-blue-950/20 rounded-lg border border-purple-200/50 dark:border-purple-800/30 space-y-2">
            <div className="flex items-start gap-2">
              <Edit3 className="w-4 h-4 text-purple-600 dark:text-purple-400 mt-0.5 flex-shrink-0" />
              <div className="space-y-1 flex-1 min-w-0">
                <p className="text-xs font-medium text-purple-900 dark:text-purple-300">Current Title</p>
                <p className="text-sm font-semibold line-clamp-2">
                  {stats.title || <span className="text-muted-foreground italic">No title set</span>}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Gamepad2 className="w-4 h-4 text-blue-600 dark:text-blue-400 flex-shrink-0" />
              <div className="flex-1">
                <p className="text-xs font-medium text-blue-900 dark:text-blue-300">Current Game</p>
                <p className="text-sm font-semibold">
                  {stats.game_name || <span className="text-muted-foreground italic">No game set</span>}
                </p>
              </div>
            </div>
          </div>
        )}

        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="title" className="flex items-center gap-2 text-sm font-medium">
              <Edit3 className="w-4 h-4 text-muted-foreground" />
              New Title
            </Label>
            <Input
              id="title"
              placeholder="Update your stream title..."
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              disabled={loading}
              className="transition-all focus:ring-2 focus:ring-purple-500/20"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="game" className="flex items-center gap-2 text-sm font-medium">
              <Gamepad2 className="w-4 h-4 text-muted-foreground" />
              Game/Category
            </Label>
            <Input
              id="game"
              placeholder="Enter game or category name..."
              value={gameName}
              onChange={(e) => setGameName(e.target.value)}
              disabled={loading}
              className="transition-all focus:ring-2 focus:ring-purple-500/20"
            />
          </div>
        </div>

        {error && (
          <div className="flex items-center gap-2 p-3 bg-red-50 dark:bg-red-950/20 text-red-700 dark:text-red-400 rounded-lg border border-red-200 dark:border-red-800/30">
            <AlertCircle className="w-4 h-4 flex-shrink-0" />
            <p className="text-sm font-medium">{error}</p>
          </div>
        )}

        {success && (
          <div className="flex items-center gap-2 p-3 bg-green-50 dark:bg-green-950/20 text-green-700 dark:text-green-400 rounded-lg border border-green-200 dark:border-green-800/30 animate-in fade-in slide-in-from-top-2 duration-300">
            <CheckCircle2 className="w-4 h-4 flex-shrink-0" />
            <p className="text-sm font-medium">Stream metadata updated successfully!</p>
          </div>
        )}

        <Button
          className="w-full bg-gradient-to-r from-purple-600 to-pink-600 hover:from-purple-700 hover:to-pink-700 text-white shadow-lg shadow-purple-500/25 transition-all hover:shadow-xl hover:shadow-purple-500/30"
          onClick={handleUpdate}
          disabled={loading || (!title && !gameName)}
          size="lg"
        >
          {loading ? (
            <>
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              Updating...
            </>
          ) : (
            <>
              <Edit3 className="w-4 h-4 mr-2" />
              Update Stream
            </>
          )}
        </Button>
      </CardContent>
    </Card>
  )
}
