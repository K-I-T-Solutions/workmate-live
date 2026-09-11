package commands

import "time"

// CommandContext holds information about the command invocation
type CommandContext struct {
	User          string   // lowercase login
	DisplayName   string
	Message       string   // full message
	Args          []string // arguments after the command
	IsModerator   bool
	IsBroadcaster bool
	Source        string // "chat" or "ui"
}

// CommandResult holds the result of a command execution
type CommandResult struct {
	Command  string `json:"command"`
	Response string `json:"response"`
	Success  bool   `json:"success"`
}

// Command defines a chat command
type Command struct {
	Name        string
	Description string
	ModOnly     bool
	Cooldown    time.Duration // 0 = no cooldown
	Run         func(ctx *CommandContext) (string, error)
}

// CommandInfo is the public representation of a command for the API
type CommandInfo struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ModOnly     bool    `json:"mod_only"`
	Builtin     bool    `json:"builtin"`
	Response    string  `json:"response"`
	Enabled     bool    `json:"enabled"`
	Cooldown    float64 `json:"cooldown"` // seconds
	Count       int     `json:"count"`
}
