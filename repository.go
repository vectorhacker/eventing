package eventing

import (
	"context"
	"errors"
	"reflect"
)

// NoVersion constants
const (
	NoVersion = 0
)

// Errors
var (
	ErrNotCommandHandler = errors.New("AggregateNotHandler")
)

// Aggregate interface
type Aggregate interface {
	On(Event) error
}

// Repository interface
type Repository struct {
	store      Store
	prototype  reflect.Type
	serializer Serializer
}

// RepositoryOption applies a different option onto the repository
type RepositoryOption func(*Repository)

// WithSerializer adds a custom serializer to the repository
func WithSerializer(s Serializer) RepositoryOption {
	return func(r *Repository) {
		r.serializer = s
	}
}

// WithStore adds a custom store to the repository
func WithStore(s Store) RepositoryOption {
	return func(r *Repository) {
		r.store = s
	}
}

// New creates a new repository with the given aggregate and options
func New(prototype Aggregate, options ...RepositoryOption) *Repository {
	t := typeOf(prototype)

	r := &Repository{
		prototype:  t,
		serializer: NewJSONSerializer(),
		store:      &memoryStore{},
	}

	for _, opt := range options {
		opt(r)
	}

	return r
}

func (r *Repository) new() Aggregate {
	return reflect.New(r.prototype).Interface().(Aggregate)
}

// Load loads up an aggregate from events
func (r *Repository) Load(ctx context.Context, aggregateID string) (Aggregate, error) {
	a := r.new()

	records, err := r.store.Load(ctx, aggregateID, 0, EndOfStream)
	if err != nil {
		return nil, err
	}

	SortRecords(records)
	for _, record := range records {
		event, err := r.serializer.Deserialize(record)
		if err != nil {
			return nil, err
		}

		if err := a.On(event); err != nil {
			return nil, err
		}
	}

	return a, nil
}

// Apply is a helper function that applies a command to an aggregate and
// saves the events. The aggregate must implement the CommandHandler interface
func (r *Repository) Apply(ctx context.Context, cmd Command) (int, error) {
	a, err := r.Load(ctx, cmd.AggregateID())
	if err != nil {
		return NoVersion, err
	}

	handler, ok := a.(CommandHandler)
	if !ok {
		return NoVersion, ErrNotCommandHandler
	}

	events, err := handler.Apply(ctx, cmd)
	if err != nil {
		return NoVersion, err
	}

	err = r.Save(ctx, cmd.AggregateID(), events)
	if err != nil {
		return NoVersion, err
	}

	version := NoVersion
	if len(events) > 0 {
		version = events[len(events)-1].EventVersion()
	}

	return version, nil
}

// Save the events from an aggregate
func (r *Repository) Save(ctx context.Context, aggregateID string, events []Event) error {
	records := make([]Record, len(events))

	for i, event := range events {
		var err error
		records[i], err = r.serializer.Serialize(event)
		if err != nil {
			return err
		}
	}

	r.store.Save(ctx, aggregateID, records)

	return nil
}
