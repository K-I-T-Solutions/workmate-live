import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { useOBSStore } from '@/store/obsStore'
import { obsAPI } from '@/services/obs'
import { Layers } from 'lucide-react'

export function OBSSourceList() {
  const { status, sources, setSources } = useOBSStore()
  const [loading, setLoading] = useState<string | null>(null)

  useEffect(() => {
    if (status?.connected && status.current_scene) {
      loadSources()
    }
  }, [status?.connected, status?.current_scene])

  const loadSources = async () => {
    if (!status?.current_scene) return
    try {
      const data = await obsAPI.getSources(status.current_scene)
      setSources(data)
    } catch (error) {
      console.error('Failed to load sources:', error)
    }
  }

  const handleToggle = async (sourceName: string, visible: boolean) => {
    if (!status?.current_scene) return
    setLoading(sourceName)
    try {
      await obsAPI.toggleSource(status.current_scene, sourceName, visible)
      await loadSources()
    } catch (error) {
      console.error('Failed to toggle source:', error)
    } finally {
      setLoading(null)
    }
  }

  if (!status?.connected) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Layers className="h-4 w-4 text-primary" />
            Sources
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-xs text-muted-foreground">OBS is not connected</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Layers className="h-4 w-4 text-primary" />
          Sources — {status.current_scene}
        </CardTitle>
      </CardHeader>
      <CardContent>
        {sources.length === 0 ? (
          <p className="text-xs text-muted-foreground">No sources in current scene</p>
        ) : (
          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-2">
            {sources.map((source) => (
              <div
                key={source.name}
                className="flex items-center justify-between gap-2 p-2 rounded-md bg-muted/30 border border-border/30"
              >
                <div className="min-w-0">
                  <p className="text-xs font-medium truncate">{source.name}</p>
                  <p className="text-[10px] text-muted-foreground">{source.type}</p>
                </div>
                <Switch
                  checked={source.visible}
                  onCheckedChange={(checked) => handleToggle(source.name, checked)}
                  disabled={loading === source.name}
                />
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
