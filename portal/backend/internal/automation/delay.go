package automation

import (
	"fmt"
	"time"
)

func executeDelay(params map[string]interface{}) error {
	var seconds float64
	switch v := params["seconds"].(type) {
	case float64:
		seconds = v
	case int:
		seconds = float64(v)
	default:
		return fmt.Errorf("delay: missing or invalid 'seconds' param")
	}
	if seconds <= 0 {
		return fmt.Errorf("delay: 'seconds' must be > 0")
	}
	if seconds > 300 {
		seconds = 300
	}
	time.Sleep(time.Duration(seconds * float64(time.Second)))
	return nil
}
