import { PageContainer } from '@/components/shared/PageContainer'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { SettingsConnectionStatus } from '@/components/settings/SettingsConnectionStatus'
import { SettingsSystemInfo } from '@/components/settings/SettingsSystemInfo'
import { SettingsUserManagement } from '@/components/settings/SettingsUserManagement'
import { SettingsAbout } from '@/components/settings/SettingsAbout'
import { SettingsConfig } from '@/components/settings/SettingsConfig'

export function SettingsPage() {
  return (
    <PageContainer title="Settings" subtitle="Konfiguration und System-Informationen">
      <Tabs defaultValue="config">
        <TabsList>
          <TabsTrigger value="config">Konfiguration</TabsTrigger>
          <TabsTrigger value="connections">Verbindungen</TabsTrigger>
          <TabsTrigger value="system">System</TabsTrigger>
          <TabsTrigger value="account">Konto</TabsTrigger>
          <TabsTrigger value="about">Über</TabsTrigger>
        </TabsList>

        <TabsContent value="config">
          <SettingsConfig />
        </TabsContent>

        <TabsContent value="connections">
          <SettingsConnectionStatus />
        </TabsContent>

        <TabsContent value="system">
          <SettingsSystemInfo />
        </TabsContent>

        <TabsContent value="account">
          <SettingsUserManagement />
        </TabsContent>

        <TabsContent value="about">
          <SettingsAbout />
        </TabsContent>
      </Tabs>
    </PageContainer>
  )
}
