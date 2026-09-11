package automation

import "fmt"

func executeChatMessage(exec TwitchExecutor, params map[string]interface{}, vars map[string]string) error {
	if exec == nil {
		return fmt.Errorf("twitch:chat_message: Twitch not connected")
	}
	msg, ok := params["message"].(string)
	if !ok || msg == "" {
		return fmt.Errorf("twitch:chat_message: missing 'message' param")
	}
	return exec.SendChatMessage(applyVars(msg, vars))
}
