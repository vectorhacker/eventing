package eventing

import "context"

// Command represents a change request
type Command interface {
	AggregateID() string
}

// CommandHandler is anything that can take in a command a return
// events based on it
type CommandHandler interface {
	Apply(context.Context, Command) ([]Event, error)
}
