package obs

import (
	"fmt"
	"log"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/events"
	"github.com/andreykaipov/goobs/api/requests/record"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/andreykaipov/goobs/api/requests/stream"
)

// Client wraps the goobs client for OBS WebSocket communication
type Client struct {
	client        *goobs.Client
	host          string
	port          int
	password      string
	connected     bool
	eventCallback func(interface{})
}

// NewClient creates a new OBS client
func NewClient(host string, port int, password string) *Client {
	return &Client{
		host:     host,
		port:     port,
		password: password,
	}
}

// Connect establishes connection to OBS WebSocket
func (c *Client) Connect() error {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)

	client, err := goobs.New(addr, goobs.WithPassword(c.password))
	if err != nil {
		return fmt.Errorf("failed to connect to OBS: %w", err)
	}

	c.client = client
	c.connected = true

	// Start listening for events
	go c.listenForEvents()

	log.Printf("Connected to OBS at %s", addr)
	return nil
}

// Disconnect closes the OBS WebSocket connection
func (c *Client) Disconnect() error {
	if c.client != nil {
		c.connected = false
		return c.client.Disconnect()
	}
	return nil
}

// IsConnected returns the connection status
func (c *Client) IsConnected() bool {
	return c.connected
}

// SetEventCallback sets the callback function for OBS events
func (c *Client) SetEventCallback(callback func(interface{})) {
	c.eventCallback = callback
}

// GetVersion returns OBS version information
func (c *Client) GetVersion() (string, error) {
	if !c.connected {
		return "", fmt.Errorf("not connected to OBS")
	}

	version, err := c.client.General.GetVersion()
	if err != nil {
		return "", err
	}

	return version.ObsVersion, nil
}

// GetScenes returns all available scenes
func (c *Client) GetScenes() ([]Scene, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to OBS")
	}

	resp, err := c.client.Scenes.GetSceneList()
	if err != nil {
		return nil, err
	}

	var sceneList []Scene
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
	if !c.connected {
		return "", fmt.Errorf("not connected to OBS")
	}

	resp, err := c.client.Scenes.GetCurrentProgramScene()
	if err != nil {
		return "", err
	}

	return resp.CurrentProgramSceneName, nil
}

// SwitchScene switches to a different scene
func (c *Client) SwitchScene(sceneName string) error {
	if !c.connected {
		return fmt.Errorf("not connected to OBS")
	}

	_, err := c.client.Scenes.SetCurrentProgramScene(&scenes.SetCurrentProgramSceneParams{
		SceneName: &sceneName,
	})

	return err
}

// GetSources returns all sources in the current scene
func (c *Client) GetSources(sceneName string) ([]Source, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to OBS")
	}

	resp, err := c.client.SceneItems.GetSceneItemList(&sceneitems.GetSceneItemListParams{
		SceneName: &sceneName,
	})
	if err != nil {
		return nil, err
	}

	var sourceList []Source
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
	if !c.connected {
		return fmt.Errorf("not connected to OBS")
	}

	// Get scene item ID first
	items, err := c.client.SceneItems.GetSceneItemList(&sceneitems.GetSceneItemListParams{
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

	_, err = c.client.SceneItems.SetSceneItemEnabled(&sceneitems.SetSceneItemEnabledParams{
		SceneName:        &sceneName,
		SceneItemId:      itemID,
		SceneItemEnabled: &visible,
	})

	return err
}

// GetStreamStatus returns the current streaming status
func (c *Client) GetStreamStatus() (*StreamStatus, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to OBS")
	}

	resp, err := c.client.Stream.GetStreamStatus()
	if err != nil {
		return nil, err
	}

	return &StreamStatus{
		Active:        resp.OutputActive,
		Reconnecting:  resp.OutputReconnecting,
		Duration:      int64(resp.OutputDuration) / 1000, // Convert ms to seconds
		Bytes:         int64(resp.OutputBytes),
	}, nil
}

// StartStreaming starts the OBS stream
func (c *Client) StartStreaming() error {
	if !c.connected {
		return fmt.Errorf("not connected to OBS")
	}

	_, err := c.client.Stream.StartStream(&stream.StartStreamParams{})
	return err
}

// StopStreaming stops the OBS stream
func (c *Client) StopStreaming() error {
	if !c.connected {
		return fmt.Errorf("not connected to OBS")
	}

	_, err := c.client.Stream.StopStream(&stream.StopStreamParams{})
	return err
}

// GetRecordStatus returns the current recording status
func (c *Client) GetRecordStatus() (*RecordingStatus, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to OBS")
	}

	resp, err := c.client.Record.GetRecordStatus()
	if err != nil {
		return nil, err
	}

	return &RecordingStatus{
		Active:   resp.OutputActive,
		Paused:   resp.OutputPaused,
		Duration: int64(resp.OutputDuration) / 1000, // Convert ms to seconds
		Bytes:    int64(resp.OutputBytes),
	}, nil
}

// StartRecording starts OBS recording
func (c *Client) StartRecording() error {
	if !c.connected {
		return fmt.Errorf("not connected to OBS")
	}

	_, err := c.client.Record.StartRecord(&record.StartRecordParams{})
	return err
}

// StopRecording stops OBS recording
func (c *Client) StopRecording() error {
	if !c.connected {
		return fmt.Errorf("not connected to OBS")
	}

	_, err := c.client.Record.StopRecord(&record.StopRecordParams{})
	return err
}

// PauseRecording pauses OBS recording
func (c *Client) PauseRecording() error {
	if !c.connected {
		return fmt.Errorf("not connected to OBS")
	}

	_, err := c.client.Record.PauseRecord(&record.PauseRecordParams{})
	return err
}

// ResumeRecording resumes OBS recording
func (c *Client) ResumeRecording() error {
	if !c.connected {
		return fmt.Errorf("not connected to OBS")
	}

	_, err := c.client.Record.ResumeRecord(&record.ResumeRecordParams{})
	return err
}

// GetStatus returns the overall OBS status
func (c *Client) GetStatus() (*OBSStatus, error) {
	if !c.connected {
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
		log.Printf("Failed to get stream status: %v", err)
		streamStatus = &StreamStatus{}
	}

	recordStatus, err := c.GetRecordStatus()
	if err != nil {
		log.Printf("Failed to get record status: %v", err)
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

// listenForEvents listens for OBS events and triggers callbacks
func (c *Client) listenForEvents() {
	if c.client == nil {
		return
	}

	for event := range c.client.IncomingEvents {
		if !c.connected {
			break
		}

		// Only broadcast if we have a callback
		if c.eventCallback != nil {
			switch e := event.(type) {
			case *events.CurrentProgramSceneChanged:
				c.eventCallback(map[string]interface{}{
					"type":       "scene_changed",
					"scene_name": e.SceneName,
				})
			case *events.StreamStateChanged:
				c.eventCallback(map[string]interface{}{
					"type":   "stream_state_changed",
					"active": e.OutputActive,
				})
			case *events.RecordStateChanged:
				c.eventCallback(map[string]interface{}{
					"type":   "record_state_changed",
					"active": e.OutputActive,
				})
			case *events.SceneItemEnableStateChanged:
				c.eventCallback(map[string]interface{}{
					"type":         "source_visibility_changed",
					"scene_name":   e.SceneName,
					"source_name":  e.SceneItemId,
					"visible":      e.SceneItemEnabled,
				})
			default:
				// Log other events for debugging
				log.Printf("OBS Event: %T", e)
			}
		}
	}
}
