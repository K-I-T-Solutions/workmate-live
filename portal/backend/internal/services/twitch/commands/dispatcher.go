package commands

import (
	"fmt"
	"log"
	"strings"
)

// Dispatcher dispatches chat commands
type Dispatcher struct {
	commands   map[string]*Command
	cooldowns  *CooldownManager
	streamInfo StreamInfoProvider
	store      *CommandStore
	channel    string
}

// NewDispatcher creates a new command dispatcher
func NewDispatcher(streamInfo StreamInfoProvider, storePath string, channel string) *Dispatcher {
	d := &Dispatcher{
		commands:   make(map[string]*Command),
		cooldowns:  NewCooldownManager(),
		streamInfo: streamInfo,
		store:      NewCommandStore(storePath),
		channel:    channel,
	}
	RegisterCommands(d)
	d.SyncCustomCommands()
	return d
}

// Register adds a command to the dispatcher
func (d *Dispatcher) Register(cmd *Command) {
	d.commands[cmd.Name] = cmd
}

// Dispatch processes a message and executes the matching command
func (d *Dispatcher) Dispatch(ctx *CommandContext) *CommandResult {
	if !strings.HasPrefix(ctx.Message, "!") {
		return nil
	}

	parts := strings.Fields(ctx.Message)
	if len(parts) == 0 {
		return nil
	}

	cmdName := strings.TrimPrefix(parts[0], "!")
	cmdName = strings.ToLower(cmdName)

	cmd, exists := d.commands[cmdName]
	if !exists {
		return nil
	}

	// Check if custom command is disabled
	if !IsBuiltinCommand(cmdName) {
		if customCmd, ok := d.store.Get(cmdName); ok && !customCmd.Enabled {
			return nil
		}
	}

	// Check mod-only permission
	if cmd.ModOnly && !ctx.IsModerator && !ctx.IsBroadcaster {
		return nil
	}

	// Check cooldown (skip for UI source)
	if ctx.Source != "ui" && d.cooldowns.IsOnCooldown(cmd.Name, ctx.User, cmd.Cooldown) {
		return nil
	}

	// Set args
	if len(parts) > 1 {
		ctx.Args = parts[1:]
	}

	// Increment count for custom commands
	if !IsBuiltinCommand(cmdName) {
		d.store.IncrementCount(cmdName)
	}

	response, err := cmd.Run(ctx)
	if err != nil {
		log.Printf("Command %s error: %v", cmd.Name, err)
		return &CommandResult{
			Command:  cmd.Name,
			Response: "Fehler beim Ausführen des Commands",
			Success:  false,
		}
	}

	return &CommandResult{
		Command:  cmd.Name,
		Response: response,
		Success:  true,
	}
}

// ListCommands returns info about all registered commands
func (d *Dispatcher) ListCommands() []*CommandInfo {
	var result []*CommandInfo

	// Built-in commands
	for _, cmd := range d.commands {
		if IsBuiltinCommand(cmd.Name) {
			result = append(result, &CommandInfo{
				Name:        cmd.Name,
				Description: cmd.Description,
				ModOnly:     cmd.ModOnly,
				Builtin:     true,
				Enabled:     true,
				Cooldown:    cmd.Cooldown.Seconds(),
			})
		}
	}

	// Custom commands from store
	for _, cmd := range d.store.List() {
		result = append(result, &CommandInfo{
			Name:        cmd.Name,
			Description: cmd.Description,
			ModOnly:     cmd.ModOnly,
			Builtin:     false,
			Response:    cmd.Response,
			Enabled:     cmd.Enabled,
			Cooldown:    cmd.Cooldown.Seconds(),
			Count:       cmd.Count,
		})
	}

	return result
}

// RegisterCustomCommand creates a Command from a CustomCommand and registers it
func (d *Dispatcher) RegisterCustomCommand(cmd CustomCommand) {
	cmdName := cmd.Name
	cmdResponse := cmd.Response

	d.commands[cmd.Name] = &Command{
		Name:        cmd.Name,
		Description: cmd.Description,
		ModOnly:     cmd.ModOnly,
		Cooldown:    cmd.Cooldown,
		Run: func(ctx *CommandContext) (string, error) {
			customCmd, ok := d.store.Get(cmdName)
			count := 0
			if ok {
				count = customCmd.Count
			}
			return renderTemplate(cmdResponse, ctx, d.channel, count), nil
		},
	}
}

// SyncCustomCommands reloads all custom commands from the store and registers them
func (d *Dispatcher) SyncCustomCommands() {
	// Remove existing custom commands
	for name := range d.commands {
		if !IsBuiltinCommand(name) {
			delete(d.commands, name)
		}
	}

	// Register all custom commands from store
	for _, cmd := range d.store.List() {
		d.RegisterCustomCommand(cmd)
	}
}

// AddCustomCommand adds a new custom command via the store and registers it
func (d *Dispatcher) AddCustomCommand(cmd CustomCommand) error {
	if err := d.store.Add(cmd); err != nil {
		return err
	}
	d.RegisterCustomCommand(cmd)
	return nil
}

// UpdateCustomCommand updates a custom command in the store and re-registers it
func (d *Dispatcher) UpdateCustomCommand(name string, cmd CustomCommand) error {
	if err := d.store.Update(name, cmd); err != nil {
		return err
	}
	// Re-register with updated data
	updated, ok := d.store.Get(name)
	if ok {
		d.RegisterCustomCommand(*updated)
	}
	return nil
}

// DeleteCustomCommand removes a custom command from the store and dispatcher
func (d *Dispatcher) DeleteCustomCommand(name string) error {
	if err := d.store.Delete(name); err != nil {
		return err
	}
	delete(d.commands, name)
	return nil
}

// renderTemplate replaces template variables in the response string
func renderTemplate(tmpl string, ctx *CommandContext, channel string, count int) string {
	r := strings.NewReplacer(
		"{user}", ctx.DisplayName,
		"{channel}", channel,
		"{args}", strings.Join(ctx.Args, " "),
		"{count}", fmt.Sprintf("%d", count),
	)
	return r.Replace(tmpl)
}
