package commands

import (
	"fmt"
	"strings"
	"time"
)

// StreamInfoProvider provides stream information for commands
type StreamInfoProvider interface {
	GetStreamUptime() (isLive bool, uptime time.Duration, err error)
}

// RegisterCommands registers all built-in commands on the dispatcher
func RegisterCommands(d *Dispatcher) {
	d.Register(&Command{
		Name:        "ping",
		Description: "Prüft ob der Bot online ist",
		Cooldown:    5 * time.Second,
		Run: func(ctx *CommandContext) (string, error) {
			return fmt.Sprintf("@%s pong 🏓", ctx.DisplayName), nil
		},
	})

	d.Register(&Command{
		Name:        "uptime",
		Description: "Zeigt die Stream-Uptime an",
		Cooldown:    10 * time.Second,
		Run: func(ctx *CommandContext) (string, error) {
			if d.streamInfo == nil {
				return "Stream-Info nicht verfügbar", nil
			}
			isLive, uptime, err := d.streamInfo.GetStreamUptime()
			if err != nil {
				return "Fehler beim Abrufen der Uptime", nil
			}
			if !isLive {
				return "Der Stream ist gerade offline.", nil
			}
			hours := int(uptime.Hours())
			minutes := int(uptime.Minutes()) % 60
			seconds := int(uptime.Seconds()) % 60
			return fmt.Sprintf("Stream ist seit %dh %dm %ds live", hours, minutes, seconds), nil
		},
	})

	d.Register(&Command{
		Name:        "help",
		Description: "Listet alle verfügbaren Commands auf",
		Cooldown:    10 * time.Second,
		Run: func(ctx *CommandContext) (string, error) {
			var names []string
			for _, cmd := range d.commands {
				if !cmd.ModOnly {
					names = append(names, "!"+cmd.Name)
				}
			}
			return "Verfügbare Commands: " + strings.Join(names, ", "), nil
		},
	})

	d.Register(&Command{
		Name:        "say",
		Description: "Bot wiederholt den Text (nur Mods)",
		ModOnly:     true,
		Run: func(ctx *CommandContext) (string, error) {
			if len(ctx.Args) == 0 {
				return "Nutzung: !say <text>", nil
			}
			return strings.Join(ctx.Args, " "), nil
		},
	})
}
