import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { useOBSStore } from "@/store/obsStore"
import { obsAPI } from "@/services/obs"

export function OBSSceneSwitcher() {
  const { status, scenes, setScenes, setStatus } = useOBSStore()
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    // Load initial OBS status
    const loadStatus = async () => {
      try {
        const obsStatus = await obsAPI.getStatus()
        setStatus(obsStatus)
      } catch (error) {
        console.error("Failed to load OBS status:", error)
      }
    }
    loadStatus()
  }, [])

  useEffect(() => {
    if (status?.connected) {
      loadScenes()
    }
  }, [status?.connected])

  const loadScenes = async () => {
    try {
      const data = await obsAPI.getScenes()
      setScenes(data)
    } catch (error) {
      console.error("Failed to load scenes:", error)
    }
  }

  const handleSwitchScene = async (sceneName: string) => {
    setLoading(true)
    try {
      await obsAPI.switchScene(sceneName)
      await loadScenes() // Refresh scenes
    } catch (error) {
      console.error("Failed to switch scene:", error)
    } finally {
      setLoading(false)
    }
  }

  if (!status?.connected) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Scene Switcher</CardTitle>
          <CardDescription>OBS is not connected</CardDescription>
        </CardHeader>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Scene Switcher</CardTitle>
        <CardDescription>Current: {status.current_scene || "None"}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-2">
        {scenes.length === 0 ? (
          <p className="text-sm text-muted-foreground">No scenes available</p>
        ) : (
          scenes.map((scene) => (
            <Button
              key={scene.name}
              className="w-full"
              variant={scene.active ? "default" : "outline"}
              disabled={loading || scene.active}
              onClick={() => handleSwitchScene(scene.name)}
            >
              {scene.name}
            </Button>
          ))
        )}
      </CardContent>
    </Card>
  )
}
