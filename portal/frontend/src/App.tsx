import { useEffect } from "react"
import { Routes, Route } from "react-router-dom"
import { TooltipProvider } from "@/components/ui/tooltip"
import { useAuthStore } from "@/store/authStore"
import { wsService } from "@/services/websocket"
import { Login } from "@/components/pages/LoginPage"
import { AppShell } from "@/components/layout/AppShell"
import { DashboardPage } from "@/components/pages/DashboardPage"
import { OBSPage } from "@/components/pages/OBSPage"
import { TwitchPage } from "@/components/pages/TwitchPage"
import { YouTubePage } from "@/components/pages/YouTubePage"
import { SettingsPage } from "@/components/pages/SettingsPage"

function App() {
  const { isAuthenticated } = useAuthStore()

  useEffect(() => {
    if (isAuthenticated) {
      wsService.connect()
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
          <Route path="/youtube" element={<YouTubePage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Routes>
      </AppShell>
    </TooltipProvider>
  )
}

export default App
