import { PageContainer } from '@/components/shared/PageContainer'
import { OBSSceneSwitcher } from '@/components/obs/OBSSceneSwitcher'
import { OBSStreamControl } from '@/components/obs/OBSStreamControl'
import { OBSRecordControl } from '@/components/obs/OBSRecordControl'
import { OBSSourceList } from '@/components/obs/OBSSourceList'

export function OBSPage() {
  return (
    <PageContainer title="OBS Studio" subtitle="Szenen, Streaming und Recording kontrollieren">
      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-3">
        <OBSSceneSwitcher />
        <OBSStreamControl />
        <OBSRecordControl />
      </div>
      <OBSSourceList />
    </PageContainer>
  )
}
