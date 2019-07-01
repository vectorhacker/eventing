package eventing

import "context"

// Repository is an event repository. This is a helper object for loading/saving events
type Repository struct {
	store      Store
	serializer Serializer
}

// New creates a new Repository
func New(store Store, serializer Serializer) *Repository {
	return &Repository{
		store:      store,
		serializer: serializer,
	}
}

// Load loads up events from the event store
func (r *Repository) Load(ctx context.Context, id string, from, to int) ([]Event, error) {
	records, err := r.store.Load(ctx, id, from, to)
	if err != nil {
		return nil, err
	}

	Sort(records)
	events := make([]Event, len(records))
	for i, record := range records {
		event, err := r.serializer.Deserialize(record)
		if err != nil {
			return nil, err
		}

		events[i] = event
	}

	return events, nil
}

// Save saves events into the event store
func (r *Repository) Save(ctx context.Context, id string, events []Event) error {
	records := make([]Record, len(events))

	Sort(records)
	for i, event := range events {
		record, err := r.serializer.Serialize(event)
		if err != nil {
			return err
		}

		records[i] = record
	}

	return r.store.Save(ctx, id, records)
}
