package eventing

import "context"

// Store is an event repository. This is a helper object for loading/saving events
type Store struct {
	recordStore RecordStore
	serializer  Serializer
}

// New creates a new Store
func New(recordStore RecordStore, serializer Serializer) *Store {
	return &Store{
		recordStore: recordStore,
		serializer:  serializer,
	}
}

// LoadAll loads all the events in the store for an stream
func (s *Store) LoadAll(ctx context.Context, id string) ([]Event, error) {
	return s.Load(ctx, id, 0, EndOfStream)
}

// Load loads up events from the event store
func (s *Store) Load(ctx context.Context, id string, startVersion, endVersion int) ([]Event, error) {
	records, err := s.recordStore.Load(ctx, id, startVersion, endVersion)
	if err != nil {
		return nil, err
	}

	SortRecords(records)
	events := make([]Event, len(records))
	for i, record := range records {
		event, err := s.serializer.Deserialize(record)
		if err != nil {
			return nil, err
		}

		events[i] = event
	}

	return events, nil
}

// Save saves events into the event store
func (s *Store) Save(ctx context.Context, id string, events []Event) error {
	records := make([]Record, len(events))

	for i, event := range events {
		record, err := s.serializer.Serialize(event)
		if err != nil {
			return err
		}

		records[i] = record
	}

	SortRecords(records)
	return s.recordStore.Save(ctx, id, records)
}
