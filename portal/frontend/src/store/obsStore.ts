import { create } from 'zustand'
import type { OBSStatus, Scene, Source } from '@/types/obs'

interface OBSStore {
  status: OBSStatus | null
  scenes: Scene[]
  sources: Source[]
  setStatus: (status: OBSStatus) => void
  setScenes: (scenes: Scene[]) => void
  setSources: (sources: Source[]) => void
  updateScene: (sceneName: string) => void
}

export const useOBSStore = create<OBSStore>((set) => ({
  status: null,
  scenes: [],
  sources: [],
  setStatus: (status) => set({ status }),
  setScenes: (scenes) => set({ scenes }),
  setSources: (sources) => set({ sources }),
  updateScene: (sceneName) =>
    set((state) => ({
      status: state.status
        ? { ...state.status, current_scene: sceneName }
        : null,
      scenes: state.scenes.map((scene) => ({
        ...scene,
        active: scene.name === sceneName,
      })),
    })),
}))
