import { useEffect } from "react"
import { Routes, Route } from "react-router-dom"
import { TooltipProvider } from "@/components/ui/tooltip"
import { useAuthStore } from "@/store/authStore"
import { useOBSStore } from "@/store/obsStore"
import { useTwitchStore } from "@/store/twitchStore"
import { useYouTubeStore } from "@/store/youtubeStore"
import { wsService } from "@/services/websocket"
import { obsAPI } from "@/services/obs"
import { twitchAPI } from "@/services/twitch"
import { youtubeAPI } from "@/services/youtube"
import { Login } from "@/components/pages/LoginPage"
import { AppShell } from "@/components/layout/AppShell"
import { DashboardPage } from "@/components/pages/DashboardPage"
import { OBSPage } from "@/components/pages/OBSPage"
import { TwitchPage } from "@/components/pages/TwitchPage"
import { YouTubePage } from "@/components/pages/YouTubePage"
import { SettingsPage } from "@/components/pages/SettingsPage"
import { CommandsPage } from "@/components/pages/CommandsPage"
import { AutomationPage } from "@/components/pages/AutomationPage"

function App() {
  const { isAuthenticated } = useAuthStore()

  useEffect(() => {
    if (isAuthenticated) {
      wsService.connect()

      // Fetch initial connection status for all services
      obsAPI.getStatus().then(useOBSStore.getState().setStatus).catch(() => {})
      twitchAPI.getStatus().then(useTwitchStore.getState().setStatus).catch(() => {})
      youtubeAPI.getStatus().then(useYouTubeStore.getState().setStatus).catch(() => {})
    }
    return () => {
      wsService.disconnect()
    }
  }, [isAuthenticated])

  if (!isAuthenticated) {
    return <Login />
  }

  return (
    <TooltipProvider>
      <AppShell>
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/obs" element={<OBSPage />} />
          <Route path="/twitch" element={<TwitchPage />} />
          <Route path="/commands" element={<CommandsPage />} />
          <Route path="/automation" element={<AutomationPage />} />
          <Route path="/youtube" element={<YouTubePage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Routes>
      </AppShell>
    </TooltipProvider>
  )
}

export default App
