package eventing

import (
	"context"
)

// Version constants
const (
	EndOfStream = 0
)

// Store represents an event store for serialized events
// this is used for supporting different backends
type Store interface {
	Load(context.Context, string, int, int) ([]Record, error)
	Save(context.Context, string, []Record) error
}
