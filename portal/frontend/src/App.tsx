import { useEffect } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { Badge } from "@/components/ui/badge"
import { useAgentStore } from "@/store/agentStore"
import { wsService } from "@/services/websocket"
import { OBSSceneSwitcher } from "@/components/OBSSceneSwitcher"
import { OBSStreamControl } from "@/components/OBSStreamControl"
import { OBSRecordControl } from "@/components/OBSRecordControl"
import { TwitchStats } from "@/components/TwitchStats"
import { TwitchChat } from "@/components/TwitchChat"
import { StreamMetadataEditor } from "@/components/StreamMetadataEditor"
import { TwitchEventAlerts } from "@/components/TwitchEventAlerts"
import { Activity, Radio, Video, Youtube } from "lucide-react"

function App() {
  const { status, connected } = useAgentStore()

  useEffect(() => {
    // Connect WebSocket on mount
    wsService.connect()

    // Cleanup on unmount
    return () => {
      wsService.disconnect()
    }
  }, [])

  const connectionStatus = connected ? "Connected" : "Disconnected"
  const connectionColor = connected ? "text-green-600" : "text-red-600"

  return (
    <div className="min-h-screen bg-gradient-to-br from-background via-background to-muted/20 p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        {/* Header */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-4xl font-bold tracking-tight bg-gradient-to-r from-purple-600 to-pink-600 bg-clip-text text-transparent">
                Workmate Live Portal
              </h1>
              <p className="text-muted-foreground mt-2">
                Creator IT Dashboard - OBS Control & Stream Integration
              </p>
            </div>
            <Badge variant={connected ? "success" : "destructive"} className="text-sm px-3 py-1">
              {connected ? "● Connected" : "● Disconnected"}
            </Badge>
          </div>
          <Separator />
        </div>

        {/* Agent Status Section */}
        <div className="space-y-4">
          <h2 className="text-xl font-semibold flex items-center gap-2">
            <Activity className="w-5 h-5 text-purple-600" />
            System Status
          </h2>
          <div className="grid gap-6 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Activity className="w-5 h-5 text-purple-600" />
                  Agent Status
                </CardTitle>
                <CardDescription>
                  Real-time system monitoring from workmate-agent
                </CardDescription>
              </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-sm">Connection</span>
                  <span className={`text-sm font-medium ${connectionColor}`}>
                    {connectionStatus}
                  </span>
                </div>
                {status && (
                  <>
                    <div className="flex items-center justify-between">
                      <span className="text-sm">Hostname</span>
                      <span className="text-sm text-muted-foreground">{status.hostname}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm">GPU</span>
                      <span className="text-sm text-muted-foreground">
                        {status.gpu.present ? (
                          status.gpu.vendors?.join(", ") || "Detected"
                        ) : (
                          "Not detected"
                        )}
                      </span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm">Audio</span>
                      <span className="text-sm text-muted-foreground">
                        {status.audio.backend} - {status.audio.ready ? "Ready" : "Not ready"}
                      </span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm">Video Devices</span>
                      <span className="text-sm text-muted-foreground">
                        {status.video.device_count} detected
                      </span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm">OBS</span>
                      <span className={`text-sm font-medium ${status.obs.running ? "text-green-600" : "text-muted-foreground"}`}>
                        {status.obs.running ? "Running" : "Not running"}
                      </span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm">Display</span>
                      <span className="text-sm text-muted-foreground">
                        {status.headless ? "Headless" : "GUI"}
                      </span>
                    </div>
                  </>
                )}
                </div>
              </CardContent>
            </Card>
          </div>
        </div>

        {/* OBS Controls Section */}
        <div className="space-y-4">
          <Separator />
          <h2 className="text-xl font-semibold flex items-center gap-2">
            <Video className="w-5 h-5 text-purple-600" />
            OBS Studio
          </h2>
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            <OBSSceneSwitcher />
            <OBSStreamControl />
            <OBSRecordControl />
          </div>
        </div>

        {/* Twitch Integration Section */}
        <div className="space-y-4">
          <Separator />
          <h2 className="text-xl font-semibold flex items-center gap-2">
            <Radio className="w-5 h-5 text-purple-600" />
            Twitch Integration
          </h2>
          <div className="grid gap-6 md:grid-cols-2">
            <TwitchStats />
            <StreamMetadataEditor />
            <TwitchChat />
            <TwitchEventAlerts />
          </div>
        </div>

        {/* YouTube Placeholder Section */}
        <div className="space-y-4">
          <Separator />
          <h2 className="text-xl font-semibold flex items-center gap-2">
            <Youtube className="w-5 h-5 text-red-600" />
            YouTube Integration
          </h2>
          <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Youtube className="w-5 h-5 text-red-600" />
                  YouTube Integration
                </CardTitle>
                <CardDescription>
                  Stream stats and chat integration
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="flex items-center gap-2">
                  <Badge variant="outline">Coming Soon</Badge>
                </div>
                <p className="text-sm text-muted-foreground">Not configured</p>
                <p className="text-xs text-muted-foreground">
                  YouTube live streaming integration planned for future release
                </p>
              </CardContent>
            </Card>
          </div>

        {status && (
          <Card>
            <CardHeader>
              <CardTitle>System Details</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-muted-foreground">Last Update:</span>
                  <div>{new Date(status.timestamp).toLocaleString()}</div>
                </div>
                {status.gpu.render_nodes && status.gpu.render_nodes.length > 0 && (
                  <div>
                    <span className="text-muted-foreground">Render Nodes:</span>
                    <div className="font-mono text-xs">{status.gpu.render_nodes.join(", ")}</div>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}

export default App
