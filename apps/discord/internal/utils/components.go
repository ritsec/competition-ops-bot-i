package utils

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

// ComponentHandler handles a component interaction (modal submit, button, etc.)
type ComponentHandler func(s *discordgo.Session, i *discordgo.InteractionCreate)

var (
	// componentHandlers maps a component's CustomID to the function that handles it
	componentHandlers = make(map[string]ComponentHandler)
	componentMu       sync.RWMutex
)

// RegisterComponent adds a handler for the given CustomID
func RegisterComponent(id string, handler ComponentHandler) {
	componentMu.Lock()
	defer componentMu.Unlock()
	componentHandlers[id] = handler
}

// UnregisterComponent removes the handler for the given CustomID
func UnregisterComponent(id string) {
	componentMu.Lock()
	defer componentMu.Unlock()
	delete(componentHandlers, id)
}

// HandleComponent runs the handler registered for the CustomID, if any
func HandleComponent(s *discordgo.Session, i *discordgo.InteractionCreate, id string) {
	componentMu.RLock()
	handler, ok := componentHandlers[id]
	componentMu.RUnlock()

	if ok {
		handler(s, i)
	}
}

// modalValue returns the value of the text input with the given CustomID
func modalValue(components []discordgo.MessageComponent, id string) string {
	for _, c := range components {
		switch c := c.(type) {
		case *discordgo.ActionsRow:
			if v := modalValue(c.Components, id); v != "" {
				return v
			}
		case *discordgo.TextInput:
			if c.CustomID == id {
				return c.Value
			}
		}
	}
	return ""
}
