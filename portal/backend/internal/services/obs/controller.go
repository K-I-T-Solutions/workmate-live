package obs

// Controller ist die Schnittstelle, über die das Portal OBS steuert —
// unabhängig davon, ob die Verbindung direkt zum OBS-WebSocket geht (Client)
// oder über einen Agent auf einem entfernten Rechner läuft (RemoteController).
type Controller interface {
	IsConnected() bool
	GetStatus() (*OBSStatus, error)
	GetScenes() ([]Scene, error)
	GetCurrentScene() (string, error)
	SwitchScene(sceneName string) error
	GetSources(sceneName string) ([]Source, error)
	ToggleSourceVisibility(sceneName, sourceName string, visible bool) error
	StartStreaming() error
	StopStreaming() error
	StartRecording() error
	StopRecording() error
	PauseRecording() error
	ResumeRecording() error
}

// Beide Implementierungen erfüllen den Vertrag.
var (
	_ Controller = (*Client)(nil)
	_ Controller = (*RemoteController)(nil)
)
