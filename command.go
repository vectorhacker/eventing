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

// CommandModel implements the Command interface
type CommandModel struct {
	ID string
}

// AggregateID implements the Command interface
func (c CommandModel) AggregateID() string {
	return c.ID
}
