import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { useOBSStore } from "@/store/obsStore"
import { obsAPI } from "@/services/obs"

export function OBSRecordControl() {
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

  const handleStartRecording = async () => {
    setLoading(true)
    try {
      await obsAPI.startRecording()
    } catch (error) {
      console.error("Failed to start recording:", error)
    } finally {
      setLoading(false)
    }
  }

  const handleStopRecording = async () => {
    setLoading(true)
    try {
      await obsAPI.stopRecording()
    } catch (error) {
      console.error("Failed to stop recording:", error)
    } finally {
      setLoading(false)
    }
  }

  const handlePauseRecording = async () => {
    setLoading(true)
    try {
      await obsAPI.pauseRecording()
    } catch (error) {
      console.error("Failed to pause recording:", error)
    } finally {
      setLoading(false)
    }
  }

  const handleResumeRecording = async () => {
    setLoading(true)
    try {
      await obsAPI.resumeRecording()
    } catch (error) {
      console.error("Failed to resume recording:", error)
    } finally {
      setLoading(false)
    }
  }

  if (!status?.connected) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Recording Control</CardTitle>
          <CardDescription>OBS is not connected</CardDescription>
        </CardHeader>
      </Card>
    )
  }

  const recording = status.recording
  const isActive = recording?.active || false
  const isPaused = recording?.paused || false

  return (
    <Card>
      <CardHeader>
        <CardTitle>Recording Control</CardTitle>
        <CardDescription>
          {isActive ? (isPaused ? "Paused" : "Recording") : "Not recording"}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {isActive && recording && (
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">Duration</span>
              <span className="font-mono">{formatDuration(recording.duration)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">File Size</span>
              <span className="font-mono">{formatBytes(recording.bytes)}</span>
            </div>
          </div>
        )}

        <div className="flex gap-2">
          {isActive ? (
            <>
              {isPaused ? (
                <Button
                  className="flex-1"
                  disabled={loading}
                  onClick={handleResumeRecording}
                >
                  Resume
                </Button>
              ) : (
                <Button
                  className="flex-1"
                  variant="secondary"
                  disabled={loading}
                  onClick={handlePauseRecording}
                >
                  Pause
                </Button>
              )}
              <Button
                className="flex-1"
                variant="destructive"
                disabled={loading}
                onClick={handleStopRecording}
              >
                Stop
              </Button>
            </>
          ) : (
            <Button
              className="w-full"
              disabled={loading}
              onClick={handleStartRecording}
            >
              Start Recording
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
