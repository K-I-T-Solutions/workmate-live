import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { useOBSStore } from "@/store/obsStore"
import { obsAPI } from "@/services/obs"

export function OBSStreamControl() {
  const { status, setStatus } = useOBSStore()
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

  const formatDuration = (seconds: number) => {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const secs = seconds % 60
    return `${hours.toString().padStart(2, "0")}:${minutes.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`
  }

  const formatBytes = (bytes: number) => {
    const mb = bytes / (1024 * 1024)
    if (mb < 1024) return `${mb.toFixed(2)} MB`
    return `${(mb / 1024).toFixed(2)} GB`
  }

  const handleStartStreaming = async () => {
    setLoading(true)
    try {
      await obsAPI.startStreaming()
    } catch (error) {
      console.error("Failed to start streaming:", error)
    } finally {
      setLoading(false)
    }
  }

  const handleStopStreaming = async () => {
    setLoading(true)
    try {
      await obsAPI.stopStreaming()
    } catch (error) {
      console.error("Failed to stop streaming:", error)
    } finally {
      setLoading(false)
    }
  }

  if (!status?.connected) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Stream Control</CardTitle>
          <CardDescription>OBS is not connected</CardDescription>
        </CardHeader>
      </Card>
    )
  }

  const streaming = status.streaming
  const isActive = streaming?.active || false

  return (
    <Card>
      <CardHeader>
        <CardTitle>Stream Control</CardTitle>
        <CardDescription>
          {isActive ? "Streaming" : "Not streaming"}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {isActive && streaming && (
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">Duration</span>
              <span className="font-mono">{formatDuration(streaming.duration)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Data Sent</span>
              <span className="font-mono">{formatBytes(streaming.bytes)}</span>
            </div>
            {streaming.reconnecting && (
              <p className="text-sm text-warning">Reconnecting...</p>
            )}
          </div>
        )}

        <div className="flex gap-2">
          {isActive ? (
            <Button
              className="flex-1"
              variant="destructive"
              disabled={loading}
              onClick={handleStopStreaming}
            >
              Stop Streaming
            </Button>
          ) : (
            <Button
              className="flex-1"
              disabled={loading}
              onClick={handleStartStreaming}
            >
              Start Streaming
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
