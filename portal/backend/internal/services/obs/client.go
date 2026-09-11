package obs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/events"
	"github.com/andreykaipov/goobs/api/requests/record"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/andreykaipov/goobs/api/requests/stream"
)

// ErrNotConnected wird zurückgegeben, wenn eine Anfrage ohne stehende
// OBS-Verbindung abgesetzt wird.
var ErrNotConnected = fmt.Errorf("not connected to OBS")

// Client spricht direkt mit dem OBS-WebSocket und hält die Verbindung über
// einen Reconnect-Loop offen. Für OBS auf einem entfernten Rechner ist
// stattdessen RemoteController vorgesehen, der über einen Agent geht.
type Client struct {
	host           string
	port           int
	password       string
	reconnectDelay time.Duration

	mu        sync.RWMutex
	client    *goobs.Client
	connected bool

	cbMu          sync.RWMutex
	eventCallback func(map[string]interface{})
}

// NewClient creates a new OBS client speaking directly to the OBS WebSocket
func NewClient(host string, port int, password string, reconnectDelay time.Duration) *Client {
	if reconnectDelay <= 0 {
		reconnectDelay = 5 * time.Second
	}

	return &Client{
		host:           host,
		port:           port,
		password:       password,
		reconnectDelay: reconnectDelay,
	}
}

// SetEventCallback sets the callback invoked for every translated OBS event
func (c *Client) SetEventCallback(cb func(map[string]interface{})) {
	c.cbMu.Lock()
	defer c.cbMu.Unlock()
	c.eventCallback = cb
}

func (c *Client) emit(event map[string]interface{}) {
	c.cbMu.RLock()
	cb := c.eventCallback
	c.cbMu.RUnlock()

	if cb != nil {
		cb(event)
	}
}

// Run hält die OBS-Verbindung bis zum Abbruch des Kontexts offen und verbindet
// nach einem Abriss automatisch neu. Blockiert; per Goroutine starten.
func (c *Client) Run(ctx context.Context) {
	for {
		if err := c.connect(); err != nil {
			log.Printf("OBS: connect failed: %v (retry in %s)", err, c.reconnectDelay)
		} else {
			log.Printf("OBS: connected to %s:%d", c.host, c.port)
			c.emit(map[string]interface{}{"type": "connection_changed", "connected": true})

			// Blockiert, bis die Verbindung abreißt.
			c.readEvents(ctx)

			c.setDisconnected()
			c.emit(map[string]interface{}{"type": "connection_changed", "connected": false})
			log.Printf("OBS: connection lost")
		}

		select {
		case <-ctx.Done():
			c.Disconnect()
			return
		case <-time.After(c.reconnectDelay):
		}
	}
}

func (c *Client) connect() error {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)

	client, err := goobs.New(addr, goobs.WithPassword(c.password))
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.client = client
	c.connected = true
	c.mu.Unlock()

	return nil
}

func (c *Client) setDisconnected() {
	c.mu.Lock()
	if c.client != nil {
		_ = c.client.Disconnect()
	}
	c.client = nil
	c.connected = false
	c.mu.Unlock()
}

// Disconnect closes the OBS WebSocket connection
func (c *Client) Disconnect() error {
	c.setDisconnected()
	return nil
}

// IsConnected returns the connection status
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// conn liefert den aktiven goobs-Client oder ErrNotConnected.
func (c *Client) conn() (*goobs.Client, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.client == nil {
		return nil, ErrNotConnected
	}

	return c.client, nil
}

// GetVersion returns OBS version information
func (c *Client) GetVersion() (string, error) {
	client, err := c.conn()
	if err != nil {
		return "", err
	}

	version, err := client.General.GetVersion()
	if err != nil {
		return "", err
	}

	return version.ObsVersion, nil
}

// GetScenes returns all available scenes
func (c *Client) GetScenes() ([]Scene, error) {
	client, err := c.conn()
	if err != nil {
		return nil, err
	}

	resp, err := client.Scenes.GetSceneList()
	if err != nil {
		return nil, err
	}

	sceneList := make([]Scene, 0, len(resp.Scenes))
	for i, scene := range resp.Scenes {
		sceneList = append(sceneList, Scene{
			Name:      scene.SceneName,
			Active:    scene.SceneName == resp.CurrentProgramSceneName,
			Index:     i,
			SceneUUID: scene.SceneUuid,
		})
	}

	return sceneList, nil
}

// GetCurrentScene returns the current active scene
func (c *Client) GetCurrentScene() (string, error) {
	client, err := c.conn()
	if err != nil {
		return "", err
	}

	resp, err := client.Scenes.GetCurrentProgramScene()
	if err != nil {
		return "", err
	}

	return resp.CurrentProgramSceneName, nil
}

// SwitchScene switches to a different scene
func (c *Client) SwitchScene(sceneName string) error {
	client, err := c.conn()
	if err != nil {
		return err
	}

	_, err = client.Scenes.SetCurrentProgramScene(&scenes.SetCurrentProgramSceneParams{
		SceneName: &sceneName,
	})

	return err
}

// GetSources returns all sources in the given scene
func (c *Client) GetSources(sceneName string) ([]Source, error) {
	client, err := c.conn()
	if err != nil {
		return nil, err
	}

	resp, err := client.SceneItems.GetSceneItemList(&sceneitems.GetSceneItemListParams{
		SceneName: &sceneName,
	})
	if err != nil {
		return nil, err
	}

	sourceList := make([]Source, 0, len(resp.SceneItems))
	for _, item := range resp.SceneItems {
		sourceList = append(sourceList, Source{
			Name:    item.SourceName,
			Type:    item.SourceType,
			Visible: item.SceneItemEnabled,
		})
	}

	return sourceList, nil
}

// ToggleSourceVisibility toggles the visibility of a source
func (c *Client) ToggleSourceVisibility(sceneName, sourceName string, visible bool) error {
	client, err := c.conn()
	if err != nil {
		return err
	}

	items, err := client.SceneItems.GetSceneItemList(&sceneitems.GetSceneItemListParams{
		SceneName: &sceneName,
	})
	if err != nil {
		return err
	}

	var itemID *int
	for _, item := range items.SceneItems {
		if item.SourceName == sourceName {
			id := item.SceneItemID
			itemID = &id
			break
		}
	}

	if itemID == nil {
		return fmt.Errorf("source not found: %s", sourceName)
	}

	_, err = client.SceneItems.SetSceneItemEnabled(&sceneitems.SetSceneItemEnabledParams{
		SceneName:        &sceneName,
		SceneItemId:      itemID,
		SceneItemEnabled: &visible,
	})

	return err
}

// GetStreamStatus returns the current streaming status
func (c *Client) GetStreamStatus() (*StreamStatus, error) {
	client, err := c.conn()
	if err != nil {
		return nil, err
	}

	resp, err := client.Stream.GetStreamStatus()
	if err != nil {
		return nil, err
	}

	return &StreamStatus{
		Active:       resp.OutputActive,
		Reconnecting: resp.OutputReconnecting,
		Duration:     int64(resp.OutputDuration) / 1000, // ms -> s
		Bytes:        int64(resp.OutputBytes),
	}, nil
}

// StartStreaming starts the OBS stream
func (c *Client) StartStreaming() error {
	client, err := c.conn()
	if err != nil {
		return err
	}

	_, err = client.Stream.StartStream(&stream.StartStreamParams{})
	return err
}

// StopStreaming stops the OBS stream
func (c *Client) StopStreaming() error {
	client, err := c.conn()
	if err != nil {
		return err
	}

	_, err = client.Stream.StopStream(&stream.StopStreamParams{})
	return err
}

// GetRecordStatus returns the current recording status
func (c *Client) GetRecordStatus() (*RecordingStatus, error) {
	client, err := c.conn()
	if err != nil {
		return nil, err
	}

	resp, err := client.Record.GetRecordStatus()
	if err != nil {
		return nil, err
	}

	return &RecordingStatus{
		Active:   resp.OutputActive,
		Paused:   resp.OutputPaused,
		Duration: int64(resp.OutputDuration) / 1000, // ms -> s
		Bytes:    int64(resp.OutputBytes),
	}, nil
}

// StartRecording starts OBS recording
func (c *Client) StartRecording() error {
	client, err := c.conn()
	if err != nil {
		return err
	}

	_, err = client.Record.StartRecord(&record.StartRecordParams{})
	return err
}

// StopRecording stops OBS recording
func (c *Client) StopRecording() error {
	client, err := c.conn()
	if err != nil {
		return err
	}

	_, err = client.Record.StopRecord(&record.StopRecordParams{})
	return err
}

// PauseRecording pauses OBS recording
func (c *Client) PauseRecording() error {
	client, err := c.conn()
	if err != nil {
		return err
	}

	_, err = client.Record.PauseRecord(&record.PauseRecordParams{})
	return err
}

// ResumeRecording resumes OBS recording
func (c *Client) ResumeRecording() error {
	client, err := c.conn()
	if err != nil {
		return err
	}

	_, err = client.Record.ResumeRecord(&record.ResumeRecordParams{})
	return err
}

// GetStatus returns the overall OBS status
func (c *Client) GetStatus() (*OBSStatus, error) {
	if !c.IsConnected() {
		return &OBSStatus{Connected: false}, nil
	}

	version, err := c.GetVersion()
	if err != nil {
		return nil, err
	}

	currentScene, err := c.GetCurrentScene()
	if err != nil {
		return nil, err
	}

	streamStatus, err := c.GetStreamStatus()
	if err != nil {
		log.Printf("OBS: failed to get stream status: %v", err)
		streamStatus = &StreamStatus{}
	}

	recordStatus, err := c.GetRecordStatus()
	if err != nil {
		log.Printf("OBS: failed to get record status: %v", err)
		recordStatus = &RecordingStatus{}
	}

	return &OBSStatus{
		Connected:    true,
		Version:      version,
		CurrentScene: currentScene,
		Streaming:    streamStatus,
		Recording:    recordStatus,
	}, nil
}

// readEvents übersetzt OBS-Events in das Portal-Format und blockiert, bis der
// Event-Channel schließt (Verbindungsabriss) oder der Kontext abgebrochen wird.
func (c *Client) readEvents(ctx context.Context) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	if client == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-client.IncomingEvents:
			if !ok {
				return
			}

			switch e := event.(type) {
			case *events.CurrentProgramSceneChanged:
				c.emit(map[string]interface{}{
					"type":       "scene_changed",
					"scene_name": e.SceneName,
				})
			case *events.StreamStateChanged:
				c.emit(map[string]interface{}{
					"type":   "stream_state_changed",
					"active": e.OutputActive,
				})
			case *events.RecordStateChanged:
				c.emit(map[string]interface{}{
					"type":   "record_state_changed",
					"active": e.OutputActive,
				})
			case *events.SceneItemEnableStateChanged:
				c.emit(map[string]interface{}{
					"type":          "source_visibility_changed",
					"scene_name":    e.SceneName,
					"scene_item_id": e.SceneItemId,
					"visible":       e.SceneItemEnabled,
				})
			}
		}
	}
}
