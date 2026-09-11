package automation

import "fmt"

func executeSceneSwitch(exec OBSExecutor, params map[string]interface{}) error {
	if exec == nil {
		return fmt.Errorf("obs:scene_switch: OBS not connected")
	}
	scene, ok := params["scene"].(string)
	if !ok || scene == "" {
		return fmt.Errorf("obs:scene_switch: missing 'scene' param")
	}
	return exec.SwitchScene(scene)
}

func executeSourceToggle(exec OBSExecutor, params map[string]interface{}) error {
	if exec == nil {
		return fmt.Errorf("obs:source_toggle: OBS not connected")
	}
	source, _ := params["source"].(string)
	if source == "" {
		return fmt.Errorf("obs:source_toggle: missing 'source' param")
	}
	visible, _ := params["visible"].(bool)

	scene, _ := params["scene"].(string)
	if scene == "" {
		var err error
		scene, err = exec.GetCurrentScene()
		if err != nil {
			return fmt.Errorf("obs:source_toggle: failed to get current scene: %w", err)
		}
	}

	return exec.ToggleSourceVisibility(scene, source, visible)
}
