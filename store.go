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

type memoryStore map[string][]Record

func (m memoryStore) Load(_ context.Context, id string, start, end int) ([]Record, error) {
	records, ok := m[id]
	if !ok {
		return []Record{}, nil
	}

	records = records[start:]
	if end > 0 {
		records = records[:end]
	}
	return records, nil
}

func (m memoryStore) Save(_ context.Context, id string, records []Record) error {
	old, ok := m[id]
	if !ok {
		old = []Record{}
	}
	m[id] = append(old, records...)
	return nil
}
